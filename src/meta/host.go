package metadata

import (
	"bytes"
	"encoding/gob"
	"path/filepath"
	"strconv"
	"time"

	"meta/proto"
	"store"
	"util"

	//"github.com/coreos/etcd/client"
	"github.com/golang/protobuf/proto"
)

type HostDeviceEvent struct {
	Ip      string
	Mode    int
	Devices [][]byte
	Time    string
}

const (
	ADD_DEVICE = 0
	DEL_DEVICE = 1
)

func ExecuteAddHostDevices(evstr []byte) error {
	if len(evstr) == 0 {
		return nil
	}

	ev := HostDeviceEvent{}
	err := ev.SetValue(evstr)
	if err != nil {
		return err
	}

	t, err := strconv.Atoi(ev.Time)
	if err != nil || t < 1460354140 {
		return NewError(EcodeEventTimeInvalid, "Not a valid time")
	}

	return AddHostDevices(ev.Ip, ev.Devices)
}

func ExecuteDelHostDevices(evstr []byte) error {
	if len(evstr) == 0 {
		return nil
	}

	ev := HostDeviceEvent{}
	err := ev.SetValue(evstr)
	if err != nil {
		return err
	}

	t, err := strconv.Atoi(ev.Time)
	if err != nil || t < 1460354140 {
		return NewError(EcodeEventTimeInvalid, "Not a valid time")
	}

	return DelHostDevices(ev.Ip, ev.Devices)
}

func (hde *HostDeviceEvent) Name() string {
	return hde.Ip
}

func (hde *HostDeviceEvent) SetName(name string) {
	hde.Ip = name
}

func (hde *HostDeviceEvent) Value() ([]byte, error) {
	var value bytes.Buffer

	enc := gob.NewEncoder(&value)

	ev := HostDeviceEvent{
		Ip:      hde.Ip,
		Mode:    hde.Mode,
		Devices: hde.Devices,
		Time:    hde.Time,
	}

	err := enc.Encode(ev)
	if err != nil {
		return value.Bytes(), err
	}

	return value.Bytes(), nil
}

func (hde *HostDeviceEvent) SetValue(v []byte) error {
	value := bytes.NewBuffer(v)

	dec := gob.NewDecoder(value)

	var ev HostDeviceEvent

	err := dec.Decode(&ev)
	if err != nil {
		return err
	}

	hde.Ip = ev.Ip
	hde.Mode = ev.Mode
	hde.Devices = ev.Devices
	hde.Time = ev.Time

	return nil
}

func getAndDecodeHost(ip string) (*metaproto.Host, error) {
	driver := store.GetDriver()

	hostkey := GenerateHostKey(ip)
	opts := map[string]string{
		"sorted": "false",
		"qurum":  "false",
	}
	data, err := driver.Get(hostkey, opts)
	if err != nil {
		if ValidKeyNotFoundError(err) == true {
			return nil, NewError(EcodeHostNotFound, "host not found.")
		}
		return nil, NewError(EcodeBackendError, err.Error())
	}

	hs := &metaproto.Host{}

	err = proto.Unmarshal([]byte(data), hs)
	if err != nil {
		return nil, NewError(EcodeRequestDecodeError, err.Error())
	}

	return hs, nil
}

func listHosts() ([]string, error) {
	driver := store.GetDriver()

	hostkey := HOSTROOT
	opts := map[string]string{
		"recursive": "false",
		"sorted":    "false",
		"quorum":    "false", //can be use for page result
	}

	hosts, err := driver.List(hostkey, opts)
	if err != nil {
		if ValidKeyNotFoundError(err) == true {
			return nil, nil
		}
		return nil, NewError(EcodeBackendError, err.Error())
	}

	return hosts, nil
}

func setAndEncodeHost(ip string, hs *metaproto.Host) error {
	data, err := proto.Marshal(hs)
	if err != nil {
		return NewError(EcodeRequestEncodeError, err.Error())
	}

	driver := store.GetDriver()

	hostkey := GenerateHostKey(ip)
	opts := map[string]string{
		"ttl":       "0",
		"prevValue": "",
		"prevIndex": "0",
	}
	err = driver.Set(hostkey, string(data), opts)
	if err != nil {
		return NewError(EcodeBackendError, err.Error())
	}

	return nil
}

func checkHostStatus(status int) bool {
	return status > HOST_MIN && status < HOST_MAX
}

func IsHostExist(ip string) (bool, error) {
	if util.ValidIPAddr(ip) == false {
		return false, NewError(EcodeParameterError, "Not Valid IP Addr.")
	}

	driver := store.GetDriver()

	hostkey := GenerateHostKey(ip)
	opts := map[string]string{
		"sorted": "false",
		"qurum":  "false",
	}
	_, err := driver.Get(hostkey, opts)

	if err != nil {
		return true, nil
	} else {
		if ValidKeyNotFoundError(err) == true {
			return false, nil
		}
		return false, nil
	}
}

func AddHost(ip string, status int, devices [][]byte) error {
	if util.ValidIPAddr(ip) == false {
		log.Debugf("[AddHost]Not Valid IP Addr.")
		return NewError(EcodeParameterError, "Not Valid IP Addr.")
	}

	if checkHostStatus(status) == false {
		log.Debugf("[AddHost]Not Valid Host Status.")
		return NewError(EcodeParameterError, "Not Valid Host Status.")
	}

	hs := &metaproto.Host{Ip: []byte(ip), Status: IntegerToBytes(status), Devices: devices}

	return setAndEncodeHost(ip, hs)
}

func GetHost(ip string) (*metaproto.Host, error) {
	if util.ValidIPAddr(ip) == false {
		return nil, NewError(EcodeParameterError, "Not Valid IP Addr.")
	}

	return getAndDecodeHost(ip)
}

func ListHosts() ([]string, error) {
	return listHosts()
}

func ListHostsName() ([]string, error) {
	hosts, err := listHosts()
	if err != nil {
		return nil, err
	}

	names := []string{}
	var name string
	for i := 0; i < len(hosts); i++ {
		name = filepath.Base(hosts[i])
		if 0 != len(name) {
			names = append(names, name)
		}
	}

	return names, nil
}

func DelHost(ip string) error {
	if util.ValidIPAddr(ip) == false {
		return NewError(EcodeParameterError, "Not Valid IP Addr.")
	}

	//before delete Host, remove all Device in the Host.
	hs, err := getAndDecodeHost(ip)
	if err != nil {
		return err
	}

	devs := [][]byte{}
	for i := 0; i < len(hs.Devices); i++ {
		deviceid, backend := ParseDeviceKey(string(hs.Devices[i]))
		err = DelDevice(deviceid, backend)
		// remember delete fail
		if err != nil {
			if err.(*Error).Code == EcodeDeviceInUse {
				return err
			}
			devs = append(devs, hs.Devices[i])
		}
	}

	if len(devs) > 0 {
		t := strconv.FormatInt(time.Now().Unix(), 10)
		ev := &HostDeviceEvent{
			Ip:      string(hs.Ip),
			Mode:    DEL_DEVICE,
			Devices: devs,
			Time:    t,
		}
		PendingOps.Add(EVENT_DEL_HOST_DEVICE, ev)
	}

	driver := store.GetDriver()

	hostkey := GenerateHostKey(ip)
	opts := map[string]string{
		"recursive": "false",
		"dir":       "false",
		"prevValue": "",
		"prevIndex": "0",
	}
	err = driver.Remove(hostkey, opts)
	if err != nil {
		if ValidKeyNotFoundError(err) == true {
			return nil
		}
		return NewError(EcodeBackendError, err.Error())
	}

	return nil
}

func ModHostIp(oldIp string, newIp string) error {
	if util.ValidIPAddr(oldIp) == false || util.ValidIPAddr(newIp) == false {
		return NewError(EcodeParameterError, "Not Valid IP Addr.")
	}

	hs, err := getAndDecodeHost(oldIp)
	if err != nil {
		return err
	}

	hs.Ip = []byte(newIp)

	/* metadata only offer function, without logic
	for i := 0; i < len(hs.Devices); i++ {
		err = UpdateDeviceNet(string(hs.Devices[i]), newIp, 0)
		//TODO remember update fail
	}
	*/

	err = setAndEncodeHost(newIp, hs)
	if err != nil {
		return err
	}

	return DelHost(oldIp)
}

func ModHostStatus(ip string, status int) error {
	if util.ValidIPAddr(ip) == false {
		return NewError(EcodeParameterError, "Not Valid IP Addr.")
	}

	if checkHostStatus(status) == false {
		return NewError(EcodeParameterError, "Not Valid Host Status.")
	}

	hs, err := getAndDecodeHost(ip)
	if err != nil {
		return err
	}

	hs.Status = IntegerToBytes(status)

	err = setAndEncodeHost(ip, hs)
	if err != nil {
		return err
	}

	return nil
}

func AddHostDevices(ip string, devices [][]byte) error {
	if util.ValidIPAddr(ip) == false {
		return NewError(EcodeParameterError, "Not Valid IP Addr.")
	}

	if len(devices) == 0 {
		return nil
	}

	hs, err := getAndDecodeHost(ip)
	if err != nil {
		return err
	}

	for i := 0; i < len(devices); i++ {
		for j := 0; j < len(hs.Devices); j++ {
			if string(devices[i]) == string(hs.Devices[j]) {
				continue
			}
		}
		hs.Devices = append(hs.Devices, devices[i])
	}

	return setAndEncodeHost(ip, hs)
}

func DelHostDevices(ip string, devices [][]byte) error {
	log.Debugf("[DelHostDevices] ip = %s", ip)

	if util.ValidIPAddr(ip) == false {
		return NewError(EcodeParameterError, "Not Valid IP Addr.")
	}

	if len(devices) == 0 {
		return nil
	}

	hs, err := getAndDecodeHost(ip)
	if err != nil {
		return err
	}

	newdevs := [][]byte{}
	flag := false
	for i := 0; i < len(hs.Devices); i++ {
		flag = false
		for j := 0; j < len(devices); j++ {
			deviceid, _ := ParseDeviceKey(string(hs.Devices[i]))
			if string(devices[j]) == deviceid {
				flag = true
				break
			}
		}
		if flag == false {
			newdevs = append(newdevs, hs.Devices[i])
		}
	}

	hs.Devices = newdevs
	return setAndEncodeHost(ip, hs)
}
