package metadata

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"path/filepath"
	"strconv"

	"meta/proto"
	"store"
	"util"

	"github.com/golang/protobuf/proto"
)

type UpdateStatusEvent struct {
	Devid   string
	Volid   string
	Status  int
	Backend string
	Time    string
}

func ExecuteUpdateDeviceStatus(evstr []byte) error {
	ev := UpdateStatusEvent{}
	err := ev.SetValue(evstr)
	if err != nil {
		return err
	}

	t, err := strconv.Atoi(ev.Time)
	if err != nil || t < 1460354140 {
		return NewError(EcodeEventTimeInvalid, "Not a valid time")
	}

	if ev.Status == DEVICE_READY {
		return FreeDevice(ev.Devid, ev.Backend, ev.Volid)
	}

	if ev.Status == DEVICE_INUSE {
		return UseDevice(ev.Devid, ev.Backend, ev.Volid)
	}

	return nil
}

func (use *UpdateStatusEvent) Name() string {
	return use.Devid
}

func (use *UpdateStatusEvent) SetName(name string) {
	use.Devid = name
}

func (use *UpdateStatusEvent) Value() ([]byte, error) {
	var value bytes.Buffer

	enc := gob.NewEncoder(&value)

	ev := UpdateStatusEvent{
		Devid:   use.Devid,
		Status:  use.Status,
		Backend: use.Backend,
	}

	err := enc.Encode(ev)
	if err != nil {
		return value.Bytes(), err
	}

	return value.Bytes(), nil
}

func (use *UpdateStatusEvent) SetValue(v []byte) error {
	value := bytes.NewBuffer(v)

	dec := gob.NewDecoder(value)

	var ev UpdateStatusEvent

	err := dec.Decode(&ev)
	if err != nil {
		return err
	}

	use.Devid = ev.Devid
	use.Backend = ev.Backend
	use.Status = ev.Status

	return nil
}

func getAndDecodeDevice(devid string, backend string) (*metaproto.Device, error) {
	driver := store.GetDriver()

	devicekey := GenerateInuseDeviceKey(devid, backend)

	opts := map[string]string{
		"sorted": "false",
		"qurum":  "false",
	}
	data, err := driver.Get(devicekey, opts)
	if err == nil {
		dv := &metaproto.Device{}
		err = proto.Unmarshal([]byte(data), dv)
		if err != nil {
			return nil, NewError(EcodeRequestDecodeError, err.Error())
		}
		return dv, err
	} else {
		if ValidKeyNotFoundError(err) == false {
			return nil, nil
		}
	}

	devicekey = GenerateFreeDeviceKey(devid, backend)
	data, err = driver.Get(devicekey, opts)
	if err != nil {
		return nil, NewError(EcodeBackendError, err.Error())
	}

	dv := &metaproto.Device{}
	err = proto.Unmarshal([]byte(data), dv)
	if err != nil {
		if ValidKeyNotFoundError(err) == true {
			return nil, NewError(EcodeDeviceNotFound, "device not found.")
		}
		return nil, NewError(EcodeRequestDecodeError, err.Error())
	}

	return dv, nil
}

func listDevices(backend string) ([]string, error) {
	driver := store.GetDriver()

	devicekey := GenerateInuseDeviceDriverKey(backend)
	opts := map[string]string{
		"recursive": "false",
		"sorted":    "false",
		"quorum":    "false",
	}
	fmt.Println("key: ", devicekey)
	devices, err := driver.List(devicekey, opts)
	fmt.Println("err: ", err)
	if err != nil {
		if ValidKeyNotFoundError(err) == true {
			return nil, nil
		}
		return nil, NewError(EcodeBackendError, err.Error())
	}

	devicekey = GenerateFreeDeviceDriverKey(backend)

	freedevices, err := driver.List(devicekey, opts)
	if err != nil {
		return nil, NewError(EcodeBackendError, err.Error())
	}

	for i := 0; i < len(freedevices); i++ {
		devices = append(devices, freedevices[i])
	}

	return devices, nil
}

func setAndEncodeDevice(devid string, dv *metaproto.Device, backend string, free bool) error {
	data, err := proto.Marshal(dv)
	if err != nil {
		return NewError(EcodeRequestDecodeError, err.Error())
	}
	log.Infof("setAndEncodeDevice %+v", dv)
	driver := store.GetDriver()

	var devicekey string
	if free == true {
		devicekey = GenerateFreeDeviceKey(devid, backend)
	} else {
		devicekey = GenerateInuseDeviceKey(devid, backend)
	}

	opts := map[string]string{
		"ttl":       "0",
		"prevValue": "",
		"prevIndex": "0",
	}
	err = driver.Set(devicekey, string(data), opts)
	if err != nil {
		return NewError(EcodeBackendError, err.Error())
	}

	return nil
}

func setDevice(devid string, ip string, port int, total int, free int, status int, identify string, backend string) error {

	dv := &metaproto.Device{
		Id:       []byte(devid),
		Host:     []byte(ip),
		Port:     IntegerToBytes(port),
		Total:    IntegerToBytes(total),
		Free:     IntegerToBytes(free),
		Status:   IntegerToBytes(status),
		Identify: []byte(identify),
		Backend:  []byte(backend),
	}

	/*
		freedev := false
		if status != DEVICE_INUSE {
			freedev = true
			dv.Volumekey = []byte(GenerateFreeDeviceKey(devid, backend))
		} else {
			dv.Volumekey = []byte(GenerateInuseDeviceKey(devid, backend))
		}
	*/
	return setAndEncodeDevice(devid, dv, backend, true)
}

func AddDevice(devid string, ip string, port int, total int, free int, status int, identify string, backend string) error {
	if len(devid) <= DEVICE_ID_MIN_LENGTH {
		log.Debugf("[AddDevice]devid length can not shorter than 2.")
		return NewError(EcodeParameterError, "devid length can not shorter than 2.")
	}
	if util.ValidIPAddr(ip) == false {
		log.Debugf("[AddDevice]Not Valid IP Addr.")
		return NewError(EcodeParameterError, "Not Valid IP Addr.")
	}
	if util.ValidPort(port) == false {
		log.Debugf("[AddDevice]Not Valid Port.")
		return NewError(EcodeParameterError, "Not Valid Port.")
	}
	if total <= 0 || free < 0 {
		log.Debugf("[AddDevice]Not Valid Capacity.")
		return NewError(EcodeParameterError, "Not Valid Capacity.")
	}
	if status <= DEVICE_MIN && status >= DEVICE_MAX {
		log.Debugf("[AddDevice]Not Valid Device Status.")
		return NewError(EcodeParameterError, "Not Valid Device Status.")
	}
	if ValidBackend(backend) == false {
		log.Debugf("[AddDevice]Not Valid Backend.")
		return NewError(EcodeParameterError, "Not Valid Backend.")
	}

	// 1. validate hostkey
	hs, err := GetHost(ip)
	if err != nil {
		return err
	}

	// 2. check device if exists
	log.Debugf("host devs:  %v", hs.Devices)
	for i := 0; i < len(hs.Devices); i++ {
		hsDevid, bk := ParseDeviceKey(string(hs.Devices[i]))
		log.Debugf("host devid = %s, backend = %s", hsDevid, bk)
		if devid == hsDevid && backend == bk {
			return NewError(EcodeDeviceExist, "Device Already Exist.")
		}
	}

	//TODO 考虑增加判断资源重复

	// 3. modify host device info
	var devicekey string
	if status == DEVICE_INUSE {
		devicekey = GenerateInuseDeviceKey(devid, backend)
	} else {
		devicekey = GenerateFreeDeviceKey(devid, backend)
	}
	devs := [][]byte{[]byte(devicekey)}
	err = AddHostDevices(ip, devs)
	if err != nil {
		return NewError(EcodeDeviceToHostsError, "Add Device To Host Error.")
	}

	// 4. add device
	err = setDevice(devid, ip, port, total, free, status, identify, backend)
	if err == nil {
		return nil
	}

	// 5. roll back
	err = DelHostDevices(ip, devs)
	if err != nil {
		//TODO what to do when roll back fail
		return NewError(EcodeDeviceToHostsError, "roll back Device To Host Error.")
	}

	return NewError(EcodeDeviceAddError, "Add Device To Host Error.")
}

func GetDevice(devid string, backend string) (*metaproto.Device, error) {
	if len(devid) <= DEVICE_ID_MIN_LENGTH {
		return nil, NewError(EcodeParameterError, "devid length can not shorter than 2.")
	}
	if ValidBackend(backend) == false {
		return nil, NewError(EcodeParameterError, "Not Valid Backend.")
	}

	return getAndDecodeDevice(devid, backend)
}

func GetFreeDevices(backend string) (devs []*metaproto.Device, err error) {
	if ValidBackend(backend) == false {
		return nil, NewError(EcodeParameterError, "Not Valid Backend.")
	}

	driver := store.GetDriver()

	devicekey := GenerateFreeDeviceDriverKey(backend)
	opts := map[string]string{
		"recursive": "false",
		"sorted":    "false",
		"quorum":    "false",
	}
	fmt.Println(devicekey)
	devices, err := driver.List(devicekey, opts)
	if err != nil {
		if ValidKeyNotFoundError(err) == true {
			return nil, nil
		}
		return nil, NewError(EcodeBackendError, err.Error())
	}
	for i := 0; i < len(devices); i++ {
		devid, _ := ParseDeviceKey(devices[i])
		dev, err := getAndDecodeDevice(devid, backend)
		if err != nil {
			continue
		}
		fmt.Println(dev)
		devs = append(devs, dev)
	}

	return devs, nil
}

func GetInuseDevices(backend string) (devs []*metaproto.Device, err error) {
	if ValidBackend(backend) == false {
		return nil, NewError(EcodeParameterError, "Not Valid Backend.")
	}

	driver := store.GetDriver()

	devicekey := GenerateInuseDeviceDriverKey(backend)
	opts := map[string]string{
		"recursive": "false",
		"sorted":    "false",
		"quorum":    "false",
	}
	devices, err := driver.List(devicekey, opts)
	if err != nil {
		if ValidKeyNotFoundError(err) == true {
			return nil, nil
		}
		return nil, NewError(EcodeBackendError, err.Error())
	}

	for i := 0; i < len(devices); i++ {
		devid, _ := ParseDeviceKey(devices[i])
		dev, err := getAndDecodeDevice(devid, backend)
		if err != nil {
			continue
		}

		devs = append(devs, dev)
	}

	return devs, nil
}

func ListDevices(backend string) ([]string, error) {
	if ValidBackend(backend) == false {
		return nil, NewError(EcodeParameterError, "Not Valid Backend.")
	}

	return listDevices(backend)
}

func ListDevicesName(backend string) ([]string, error) {
	if ValidBackend(backend) == false {
		return nil, NewError(EcodeParameterError, "Not Valid Backend.")
	}

	devices, err := listDevices(backend)
	if err != nil {
		return nil, err
	}

	names := []string{}
	var name string
	for i := 0; i < len(devices); i++ {
		name = filepath.Base(devices[i])
		if 0 != len(name) {
			names = append(names, name)
		}
	}

	return names, nil
}

func GetDeviceHostKey(devid string, backend string) (string, error) {
	if len(devid) <= DEVICE_ID_MIN_LENGTH {
		return "", NewError(EcodeParameterError, "devid length can not shorter than 2.")
	}
	if ValidBackend(backend) == false {
		return "", NewError(EcodeParameterError, "Not Valid Backend.")
	}

	dv, err := getAndDecodeDevice(devid, backend)
	if err != nil {
		return "", err
	}
	return string(dv.Host), nil
}

func GetDeviceNet(devid string, backend string) (string, int, error) {
	if len(devid) <= DEVICE_ID_MIN_LENGTH {
		return "", 0, NewError(EcodeParameterError, "devid length can not shorter than 2.")
	}
	if ValidBackend(backend) == false {
		return "", 0, NewError(EcodeParameterError, "Not Valid Backend.")
	}

	dv, err := getAndDecodeDevice(devid, backend)
	if err != nil {
		return "", 0, err
	}
	port, err := BytesToInteger(dv.Port)
	if err != nil {
		port = 0
	}
	return GetHostIpFromKey(string(dv.Host)), port, nil
}

func GetDeviceStatus(devid string, backend string) (int, error) {
	if len(devid) <= DEVICE_ID_MIN_LENGTH {
		return 0, NewError(EcodeParameterError, "devid length can not shorter than 2.")
	}
	if ValidBackend(backend) == false {
		return 0, NewError(EcodeParameterError, "Not Valid Backend.")
	}

	dv, err := getAndDecodeDevice(devid, backend)
	if err != nil {
		return DEVICE_UNKNOWN, err
	}
	status, err := BytesToInteger(dv.Status)
	if err != nil {
		status = DEVICE_UNKNOWN
	}
	return status, err
}

func DelDevice(devid string, backend string) error {
	if len(devid) <= DEVICE_ID_MIN_LENGTH {
		return NewError(EcodeParameterError, "devid length can not shorter than 2.")
	}
	if ValidBackend(backend) == false {
		return NewError(EcodeParameterError, "Not Valid Backend.")
	}
	log.Debugf("device id  = %s", devid, ", backend = ", backend)
	dv, err := getAndDecodeDevice(devid, backend)
	if err != nil {
		log.Errorf("[DelDevice] getAndDecodeDevice error: %s", err.Error())
		return err
	}
	log.Debugf("device = %v", dv)
	var free = false
	status, err := BytesToInteger(dv.Status)
	if err != nil {
		return err
	}
	if status == DEVICE_INUSE {
		return NewError(EcodeDeviceInUse, devid)
	}
	free = true

	// delete volume device information
	if len(string(dv.Volumekey)) > 0 {
		volumeid, driverName := ParseVolumekey(string(dv.Volumekey))
		log.Debugf("volume key = %s", dv.Volumekey)
		err = DelVolumeDevice(volumeid, driverName, devid)
		if err != nil {
			log.Errorf("[DelDevice] DelVolumeDevice error: %s", err.Error())
			return err
		}
	}

	// delete host device information
	log.Debugf("Host = %s, devid = %s", GetHostIpFromKey(string(dv.Host)), devid)
	err = DelHostDevices(GetHostIpFromKey(string(dv.Host)), [][]byte{[]byte(devid)})
	if err != nil {
		log.Errorf("[DelDevice] DelHostDevices error: %s", err.Error())
		return err
	}

	var devicekey string
	if free == true {
		devicekey = GenerateFreeDeviceKey(devid, backend)
	} else {
		devicekey = GenerateInuseDeviceKey(devid, backend)
	}
	log.Debugf("device key = %s", devicekey)
	driver := store.GetDriver()

	opts := map[string]string{
		"recursive": "false",
		"dir":       "false",
		"prevValue": "",
		"prevIndex": "0",
	}
	err = driver.Remove(devicekey, opts)
	if err != nil {
		if ValidKeyNotFoundError(err) == true {
			return nil
		}
		return NewError(EcodeBackendError, err.Error())
	}

	return nil
}

func UpdateDevice(devid string, ip string, port int, total int, free int, status int, resource string, backend string) error {
	if len(devid) <= DEVICE_ID_MIN_LENGTH {
		return NewError(EcodeParameterError, "devid length can not shorter than 2.")
	}
	if util.ValidIPAddr(ip) == false {
		return NewError(EcodeParameterError, "Not Valid IP Addr.")
	}
	if util.ValidPort(port) == false {
		return NewError(EcodeParameterError, "Not Valid Port.")
	}
	if total <= 0 || free < 0 {
		return NewError(EcodeParameterError, "Not Valid Capacity.")
	}
	if status <= DEVICE_MIN && status >= DEVICE_MAX {
		return NewError(EcodeParameterError, "Not Valid Device Status.")
	}
	if ValidBackend(backend) == false {
		return NewError(EcodeParameterError, "Not Valid Backend.")
	}

	return setDevice(devid, ip, port, total, free, status, resource, backend)
}

func UpdateDeviceNet(devid string, ip string, port int, backend string) error {
	if len(devid) <= DEVICE_ID_MIN_LENGTH {
		return NewError(EcodeParameterError, "devid length can not shorter than 2.")
	}
	if util.ValidIPAddr(ip) == false {
		return NewError(EcodeParameterError, "Not Valid IP Addr.")
	}
	if util.ValidPort(port) == false {
		return NewError(EcodeParameterError, "Not Valid Port.")
	}
	if ValidBackend(backend) == false {
		return NewError(EcodeParameterError, "Not Valid Backend.")
	}

	dv, err := getAndDecodeDevice(devid, backend)
	if err != nil {
		log.Errorf("[UpdateDeviceNet] getAndDecodeDevice error: %s", err.Error())
		return err
	}

	flag := false
	hostip := GetHostIpFromKey(string(dv.GetHost()))
	if len(hostip) != 0 && hostip != ip {
		dv.Host = []byte(GenerateHostKey(ip))
		flag = true
	}

	dvPort, err := BytesToInteger(dv.GetPort())
	if err != nil {
		return err
	}

	if port != 0 && dvPort != port {
		dv.Port = []byte((strconv.Itoa(port)))
		flag = true
	}

	if flag == true {
		free := false
		status, err := BytesToInteger(dv.Status)
		if err != nil {
			return err
		}
		if status != DEVICE_INUSE {
			free = true
		}
		return setAndEncodeDevice(devid, dv, backend, free)
	}

	return nil
}

func UpdateDeviceCapacity(devid string, total int, free int, backend string) error {
	if len(devid) <= DEVICE_ID_MIN_LENGTH {
		return NewError(EcodeParameterError, "devid length can not shorter than 2.")
	}
	if total <= 0 || free < 0 {
		return NewError(EcodeParameterError, "Not Valid Capacity.")
	}
	if ValidBackend(backend) == false {
		return NewError(EcodeParameterError, "Not Valid Backend.")
	}

	dv, err := getAndDecodeDevice(devid, backend)
	if err != nil {
		log.Errorf("[UpdateDeviceCapacity] getAndDecodeDevice error: %s", err.Error())
		return err
	}

	flag := false
	dvTotal, err := BytesToInteger(dv.GetTotal())
	if err != nil {
		return err
	}
	if dvTotal != total {
		dv.Total = IntegerToBytes(total)
		flag = true
	}

	dvFree, err := BytesToInteger(dv.GetFree())
	if err != nil {
		return err
	}

	if dvFree != free {
		dv.Free = IntegerToBytes(free)
		flag = true
	}

	if flag == true {
		freedev := false
		status, err := BytesToInteger(dv.Status)
		if err != nil {
			return err
		}
		if status != DEVICE_INUSE {
			freedev = true
		}
		return setAndEncodeDevice(devid, dv, backend, freedev)
	}

	return nil
}

func UpdateDeviceStatus(devid string, status int, backend string, optime int) error {
	if len(devid) <= DEVICE_ID_MIN_LENGTH {
		return NewError(EcodeParameterError, "devid length can not shorter than 2.")
	}
	if status <= DEVICE_MIN && status >= DEVICE_MAX {
		return NewError(EcodeParameterError, "Not Valid Device Status.")
	}
	if ValidBackend(backend) == false {
		return NewError(EcodeParameterError, "Not Valid Backend.")
	}

	dv, err := getAndDecodeDevice(devid, backend)
	if err != nil {
		log.Errorf("[UpdateDeviceStatus] getAndDecodeDevice error: %s", err.Error())
		return err
	}

	t, err := strconv.Atoi(string(dv.Optime))
	if err != nil {
		return NewError(EcodeMetaTimeInvalid, "meta Device Optime Invalid.")
	}

	if optime != 0 && t > optime {
		return NewError(EcodeEventTimeExipre, "Event time expire.")
	}

	flag := false

	dvStatus, err := BytesToInteger(dv.GetStatus())
	if err != nil {
		return err
	}
	if dvStatus != status {
		dv.Status = IntegerToBytes(status)
		flag = true
	}

	if flag == true {
		freedev := false
		status, err := BytesToInteger(dv.Status)
		if err != nil {
			return err
		}
		if status != DEVICE_INUSE {
			freedev = true
		}
		return setAndEncodeDevice(devid, dv, backend, freedev)
	}

	return nil
}

func FreeDevice(devid string, backend string, volumeid string) error {
	if len(devid) <= DEVICE_ID_MIN_LENGTH {
		return NewError(EcodeParameterError, "devid length can not shorter than 2.")
	}
	if ValidBackend(backend) == false {
		return NewError(EcodeParameterError, "Not Valid Backend.")
	}

	dv, err := getAndDecodeDevice(devid, backend)
	if err != nil {
		log.Errorf("[FreeDevice] getAndDecodeDevice error: %s", err.Error())
		return err
	}

	dvStatus, err := BytesToInteger(dv.GetStatus())
	if err != nil {
		return err
	}

	if dvStatus != DEVICE_INUSE {
		return nil
	}

	dv.Status = IntegerToBytes(DEVICE_READY)
	dv.Volumekey = []byte{}

	driver := store.GetDriver()

	opts := map[string]string{
		"recursive": "false",
		"dir":       "false",
		"prevValue": "",
		"prevIndex": "0",
	}
	err = driver.Remove(GenerateInuseDeviceKey(devid, backend), opts)
	if err != nil {
		//TODO add to delay queue
		if ValidKeyNotFoundError(err) == true {
			return nil
		}
		return NewError(EcodeBackendError, err.Error())
	}

	return setAndEncodeDevice(devid, dv, backend, true)
}

func UseDevice(devid string, backend string, volumeid string) error {
	if len(devid) <= DEVICE_ID_MIN_LENGTH {
		return NewError(EcodeParameterError, "devid length can not shorter than 2.")
	}
	if ValidBackend(backend) == false {
		return NewError(EcodeParameterError, "Not Valid Backend.")
	}

	dv, err := getAndDecodeDevice(devid, backend)
	if err != nil {
		log.Errorf("[UseDevice] getAndDecodeDevice error: %s", err.Error())
		return err
	}

	dvStatus, err := BytesToInteger(dv.GetStatus())
	if err != nil {
		return err
	}

	if dvStatus == DEVICE_INUSE {
		return NewError(EcodeDeviceInUse, "device already in use.")
	}

	dv.Status = IntegerToBytes(DEVICE_INUSE)
	dv.Volumekey = []byte(GenerateVolumeKey(volumeid, backend))
	driver := store.GetDriver()

	opts := map[string]string{
		"recursive": "false",
		"dir":       "false",
		"prevValue": "",
		"prevIndex": "0",
	}
	err = driver.Remove(GenerateFreeDeviceKey(devid, backend), opts)
	if err != nil {
		//TODO add to delay queue
		if ValidKeyNotFoundError(err) == true {
			return nil
		}
		return NewError(EcodeBackendError, err.Error())
	}

	return setAndEncodeDevice(devid, dv, backend, false)
}
