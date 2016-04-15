package metadata

import (
	"fmt"
	"meta/proto"
	"path/filepath"
	"store"
	"strconv"
	"time"

	//"github.com/coreos/etcd/client"
	"github.com/golang/protobuf/proto"
)

func getAndDecodeVolume(volumeid string, driverName string) (*metaproto.Volume, error) {
	driver := store.GetDriver()

	volumekey := GenerateVolumeKey(volumeid, driverName)
	fmt.Println("volume key: ", volumekey)
	opts := map[string]string{
		"sorted": "false",
		"qurum":  "false",
	}
	data, err := driver.Get(volumekey, opts)
	if err != nil {
		// ToDo: key not found will create new Container
		if ValidKeyNotFoundError(err) == true {
			return nil, NewError(EcodeVolumeNotFound, "Volume not found.")
		}
		log.Errorf("[getAndDecodeVolume] driver.Get error: %s, key: %s", err.Error(), volumekey)
		return nil, NewError(EcodeBackendError, err.Error())
	}

	vl := &metaproto.Volume{}

	err = proto.Unmarshal([]byte(data), vl)
	if err != nil {
		log.Errorf("[getAndDecodeVolume] proto.Unmarshal error: %s, key: %s", err.Error(), volumekey)
		return nil, NewError(EcodeRequestDecodeError, err.Error())
	}

	log.Debugf("[getAndDecodeVolume] volume : %v", vl)

	return vl, nil
}

func listVolumes(driverName string) ([]string, error) {
	driver := store.GetDriver()

	volumekey := GenerateVolumeDriverKey(driverName)
	opts := map[string]string{
		"recursive": "false",
		"sorted":    "false",
		"quorum":    "false",
	}
	volumes, err := driver.List(volumekey, opts)
	if err != nil {
		if ValidKeyNotFoundError(err) == true {
			return nil, nil
		}
		return nil, NewError(EcodeRequestDecodeError, err.Error())
	}

	return volumes, nil
}

func setAndEncodeVolume(vl *metaproto.Volume, driverName string) error {
	data, err := proto.Marshal(vl)
	if err != nil {
		log.Errorf("[setAndEncodeVolume] proto.marshal error: %s", err.Error())
		return NewError(EcodeRequestEncodeError, err.Error())
	}

	driver := store.GetDriver()
	volumekey := GenerateVolumeKey(string(vl.Id), driverName)
	opts := map[string]string{
		"ttl":       "0",
		"prevValue": "",
		"prevIndex": "0",
	}
	err = driver.Set(volumekey, string(data), opts)
	if err != nil {
		log.Errorf("[setAndEncodeVolume] driver.Set error: %s", err.Error())
		return NewError(EcodeBackendError, err.Error())
	}

	return nil
}

func validVolumeID(volumeid string) bool {
	return len(volumeid) > VOLUME_ID_MIN_LENGTH
}

func IsVolumeExist(volumeid string, driverName string) (bool, error) {
	if validVolumeID(volumeid) == false {
		return false, NewError(EcodeParameterError, "Not Valid Volume ID.")
	}
	if ValidDriverName(driverName) == false {
		return false, NewError(EcodeParameterError, "Not Valid Volume Driver Name.")
	}
	driver := store.GetDriver()

	volumekey := GenerateVolumeKey(volumeid, driverName)
	opts := map[string]string{
		"sorted": "false",
		"qurum":  "false",
	}
	_, err := driver.Get(volumekey, opts)

	if err != nil {
		return true, nil
	} else {
		if ValidKeyNotFoundError(err) == true {
			return false, nil
		}
		return false, nil
	}
}

func AddVolume(vl *metaproto.Volume, driverName string) error {
	if vl == nil || validVolumeID(string(vl.Id)) == false {
		return NewError(EcodeParameterError, "Not Valid Volume Struct.")
	}
	if ValidDriverName(driverName) == false {
		return NewError(EcodeParameterError, "Not Valid Volume Driver Name.")
	}

	err := setAndEncodeVolume(vl, driverName)
	if err != nil {
		return err
	}

	// set devices status INUSE
	for i := 0; i < len(vl.Devices); i++ {
		err := UseDevice(string(vl.Devices[i].Deviceid), driverName, string(vl.Id))
		if err != nil {
			for j := 0; j <= i; j++ {
				//USQueue.Enqueue(string(vl.Devices[j].Deviceid), DEVICE_READY, driverName) //undo all done
				//USQueue.Count <- i                                                        //notify update status goroutine to handle events
				t := strconv.FormatInt(time.Now().Unix(), 10)
				ev := &UpdateStatusEvent{
					Devid:   string(vl.Devices[j].Deviceid),
					Volid:   string(vl.Id),
					Status:  DEVICE_READY,
					Backend: driverName,
					Time:    t,
				}
				PendingOps.Add(EVENT_UPDATE_DEVICE_STATUS, ev)

			}
			return err
		}
	}

	return nil
}

func GetVolume(volumeid string, driverName string) (*metaproto.Volume, error) {
	if validVolumeID(volumeid) == false {
		return nil, NewError(EcodeParameterError, "Not Valid Volume ID.")
	}
	if ValidDriverName(driverName) == false {
		return nil, NewError(EcodeParameterError, "Not Valid Volume Driver Name.")
	}

	return getAndDecodeVolume(volumeid, driverName)
}

func ListVolumes(driverName string) ([]string, error) {
	if ValidDriverName(driverName) == false {
		return nil, NewError(EcodeParameterError, "Not Valid Volume Driver Name.")
	}

	return listVolumes(driverName)
}

func ListVolumesName(driverName string) ([]string, error) {
	if ValidDriverName(driverName) == false {
		return nil, NewError(EcodeParameterError, "Not Valid Volume Driver Name.")
	}

	volumes, err := listVolumes(driverName)
	if err != nil {
		return nil, err
	}

	names := []string{}
	var name string
	for i := 0; i < len(volumes); i++ {
		name = filepath.Base(volumes[i])
		if 0 != len(name) {
			names = append(names, name)
		}
	}

	return names, nil
}

func DelVolume(volumeid string, driverName string) error {
	if validVolumeID(volumeid) == false {
		return NewError(EcodeParameterError, "Not Valid Volume ID.")
	}
	if ValidDriverName(driverName) == false {
		return NewError(EcodeParameterError, "Not Valid Volume Driver Name.")
	}

	volumekey := GenerateVolumeKey(volumeid, driverName)

	vl, err := GetVolume(volumeid, driverName)
	if err != nil {
		return err
	}

	// make sure volume if is in use
	if len(vl.Containers) != 0 {
		return NewError(EcodeVolumeInUse, "Volume in use")
	}

	driver := store.GetDriver()
	opts := map[string]string{
		"recursive": "false",
		"dir":       "false",
		"prevValue": "",
		"prevIndex": "0",
	}

	err = driver.Remove(volumekey, opts)
	if err != nil {
		if ValidKeyNotFoundError(err) == true {
			return nil
		}
		return NewError(EcodeBackendError, err.Error())
	}

	//update device READY state
	for i := 0; i < len(vl.Devices); i++ {
		err = FreeDevice(string(vl.Devices[i].Deviceid), driverName, volumeid)
		if err != nil {
			t := strconv.FormatInt(time.Now().Unix(), 10)
			ev := &UpdateStatusEvent{
				Devid:   string(vl.Devices[i].Deviceid),
				Volid:   volumeid,
				Status:  DEVICE_READY,
				Backend: driverName,
				Time:    t,
			}
			PendingOps.Add(EVENT_UPDATE_DEVICE_STATUS, ev)
		}
	}

	return nil
}

func GetVolumeWRContainer(volumeid string, driverName string) ([]byte, error) {
	if validVolumeID(volumeid) == false {
		return nil, NewError(EcodeParameterError, "Not Valid Volume ID.")
	}
	if ValidDriverName(driverName) == false {
		return nil, NewError(EcodeParameterError, "Not Valid Volume Driver Name.")
	}

	vl, err := getAndDecodeVolume(volumeid, driverName)
	if err != nil {
		return []byte{}, err
	}

	return vl.Writable, nil
}

func GetVolumeROContainers(volumeid string, driverName string) ([][]byte, error) {
	if validVolumeID(volumeid) == false {
		return nil, NewError(EcodeParameterError, "Not Valid Volume ID.")
	}
	if ValidDriverName(driverName) == false {
		return nil, NewError(EcodeParameterError, "Not Valid Volume Driver Name.")
	}

	vl, err := getAndDecodeVolume(volumeid, driverName)
	if err != nil {
		return [][]byte{}, err
	}

	containers := [][]byte{}
	var cons *metaproto.Volume_OwnerContainer
	for i := 0; i < len(vl.Containers); i++ {
		cons = vl.Containers[i]
		if string(cons.Mode) == ROVolume {
			containers = append(containers, cons.Containerid)
		}
	}

	return containers, nil
}

func SetVolumeContainer(volumeid string, vct *metaproto.Volume_OwnerContainer, driverName string, force bool) error {
	if validVolumeID(volumeid) == false {
		return NewError(EcodeParameterError, "Not Valid Volume ID.")
	}
	if vct == nil {
		return NewError(EcodeParameterError, "Not Valid Volume OwnerContainer.")
	}
	if ValidDriverName(driverName) == false {
		return NewError(EcodeParameterError, "Not Valid Volume Driver Name.")
	}

	vl, err := getAndDecodeVolume(volumeid, driverName)
	if err != nil {
		return err
	}

	if len(vl.Writable) != 0 && force == false {
		return NewError(EcodeWRContainerExist, "rw container already exists.")
	}

	var c *metaproto.Volume_OwnerContainer
	for i := 0; i < len(vl.Containers); i++ {
		c = vl.Containers[i]
		if string(c.Containerid) == string(vct.Containerid) {
			if string(c.Mode) == string(vct.Mode) {
				return nil
			}

			c.Mode = vct.Mode // change volume mode
			return setAndEncodeVolume(vl, driverName)
		}
	}

	if string(vct.Mode) == RWVolume {
		vl.Writable = vct.Containerid
	}

	vl.Containers = append(vl.Containers, vct)

	return setAndEncodeVolume(vl, driverName)
}

func DelVolumeContainer(volumeid string, driverName string, containerid string) error {
	if validVolumeID(volumeid) == false {
		return NewError(EcodeParameterError, "Not Valid Volume ID.")
	}
	if ValidDriverName(driverName) == false {
		return NewError(EcodeParameterError, "Not Valid Volume Driver Name.")
	}
	if len(containerid) != CONTAINER_ID_LENGTH {
		return NewError(EcodeParameterError, "Not Valid Container ID.")
	}

	vl, err := getAndDecodeVolume(volumeid, driverName)
	if err != nil {
		return err
	}

	newCons := []*metaproto.Volume_OwnerContainer{}
	var c *metaproto.Volume_OwnerContainer
	for i := 0; i < len(vl.Containers); i++ {
		c = vl.Containers[i]
		if string(c.Containerid) == containerid {
			vl.Writable = []byte("")
			continue
		}

		newCons = append(newCons, c)
	}

	vl.Containers = newCons

	return setAndEncodeVolume(vl, driverName)
}

func AddVolumeDevice(volumeid string, driverName string, vad *metaproto.Volume_AttachDevice) error {
	if validVolumeID(volumeid) == false {
		return NewError(EcodeParameterError, "Not Valid Volume ID.")
	}
	if ValidDriverName(driverName) == false {
		return NewError(EcodeParameterError, "Not Valid Volume Driver Name.")
	}
	if vad == nil {
		return NewError(EcodeParameterError, "Not Valid Volume Attach Device.")
	}

	vl, err := getAndDecodeVolume(volumeid, driverName)
	if err != nil {
		return err
	}

	for i := 0; i < len(vl.Devices); i++ {
		if string(vl.Devices[i].Deviceid) == volumeid {
			return nil
		}
	}

	vl.Devices = append(vl.Devices, vad)

	return setAndEncodeVolume(vl, driverName)
}

func DelVolumeDevice(volumeid string, driverName string, deviceid string) error {
	if validVolumeID(volumeid) == false {
		return NewError(EcodeParameterError, "Not Valid Volume ID.")
	}
	if ValidDriverName(driverName) == false {
		return NewError(EcodeParameterError, "Not Valid Volume Driver Name.")
	}
	if len(volumeid) <= VOLUME_ID_MIN_LENGTH {
		return NewError(EcodeParameterError, "devid length can not shorter than 2.")
	}

	pos := -1
	log.Debugf("[DelVolumeDevice] volumeid = %s,  driver = %s", volumeid, driverName)
	vl, err := getAndDecodeVolume(volumeid, driverName)
	if err != nil {
		return err
	}

	for i := 0; i < len(vl.Devices); i++ {
		if string(vl.Devices[i].Deviceid) == deviceid {
			pos = i
			break
		}
	}

	if pos == -1 {
		return NewError(EcodeVolumeDeviceMiss, deviceid)
	}

	vl.Devices = append(vl.Devices[:pos-1], vl.Devices[pos+1:]...)

	return setAndEncodeVolume(vl, driverName)
}
