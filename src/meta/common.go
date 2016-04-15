package metadata

import (
	"path/filepath"
	"strconv"
	"strings"

	"store"

	"github.com/Sirupsen/logrus"
	//"github.com/coreos/etcd/client"
	//etcderr "github.com/coreos/etcd/error"
)

var (
	log = logrus.WithFields(logrus.Fields{"pkg": "metadata"})
)

const (
	ROOT          = "/comet/"
	HOSTROOT      = ROOT + "/hosts/"
	DEVICEROOT    = ROOT + "/devices/"
	CONTAINERROOT = ROOT + "/containers/"
	VOLUMEROOT    = ROOT + "/volumes/"

	INUSE = "/inuse/"
	FREE  = "/free/"
)

const (
	HOST_MIN     = 0
	HOST_ONLINE  = 1
	HOST_OFFLINE = 2
	HOST_DEGRADE = 3
	HOST_ERROR   = 4

	HOST_MAX = 5
)

const (
	DEVICE_MIN     = 10
	DEVICE_INUSE   = 11
	DEVICE_READY   = 12
	DEVICE_OFFLINE = 13
	DEVICE_UNKNOWN = 14
	DEVICE_MAX     = 15

	DEVICE_ID_MIN_LENGTH = 2
)

const (
	CONTAINERONLINE  = 20
	CONTAINEROFFLINE = 21

	CONTAINER_ID_LENGTH = 64
)

const (
	VOLUME_ONLINE = 30
	VOLUME_UNKNOW = 31
	VOLUME_INUSE  = 32

	VOLUME_ID_MIN_LENGTH = 2
)

const (
	RWVolume = "rw"
	ROVolume = "ro"
)

const (
	SAN  = "SAN"
	NFS  = "NFS"
	CEPH = "CEPH"
	GFS  = "GLUSTERFS"
)

func ValidBackend(backend string) bool {
	return backend == SAN || backend == NFS || backend == CEPH || backend == GFS
}

func ValidDriverName(driverName string) bool {
	return ValidBackend(driverName)
}

func ValidKeyNotFoundError(err error) bool {
	strs := strings.Split(err.Error(), ":")
	ecode := strings.Trim(strs[0], " ")

	if ecode == "100" {
		return true
	}

	return false
}

func GenerateHostKey(ip string) string {
	hostkey, err := filepath.Abs(HOSTROOT + ip)
	if err != nil {
		return ""
	}
	return hostkey
}

func GenerateInuseDeviceKey(devid string, backend string) string {
	devicekey, err := filepath.Abs(DEVICEROOT + backend + INUSE + devid)
	if err != nil {
		return ""
	}
	return devicekey
}

func GenerateFreeDeviceKey(devid string, backend string) string {
	devicekey, err := filepath.Abs(DEVICEROOT + backend + FREE + devid)
	if err != nil {
		return ""
	}
	return devicekey
}

func GenerateInuseDeviceDriverKey(backend string) string {
	devicekey, err := filepath.Abs(DEVICEROOT + backend + INUSE)
	if err != nil {
		return ""
	}
	return devicekey
}

func GenerateFreeDeviceDriverKey(backend string) string {
	devicekey, err := filepath.Abs(DEVICEROOT + backend + FREE)
	if err != nil {
		return ""
	}
	return devicekey
}

func ParseDeviceKey(devicekey string) (string, string) {
	deviceid := filepath.Base(devicekey)
	backend := filepath.Base(filepath.Dir(filepath.Dir(devicekey)))

	return deviceid, backend
}

func GenerateContainerKey(containerid string) string {
	containerkey, err := filepath.Abs(CONTAINERROOT + containerid)
	if err != nil {
		return ""
	}

	return containerkey
}

func GenerateVolumeKey(volumeid string, driverName string) string {
	volumekey, err := filepath.Abs(VOLUMEROOT + driverName + "/" + volumeid)
	if err != nil {
		return ""
	}

	return volumekey
}

func ParseVolumekey(volumekey string) (string, string) {
	volumeid := filepath.Base(volumekey)
	driverName := filepath.Base(filepath.Dir(filepath.Dir(volumekey)))

	return volumeid, driverName
}

func GenerateVolumeDriverKey(driverName string) string {
	volumekey, err := filepath.Abs(VOLUMEROOT + driverName + "/")
	if err != nil {
		return ""
	}

	return volumekey
}

func GetHostIpFromKey(hostkey string) string {
	return filepath.Base(hostkey)
}

func IntegerToBytes(i int) []byte {
	return []byte((strconv.Itoa(i)))
}

func BytesToInteger(b []byte) (int, error) {
	return strconv.Atoi(string(b))
}

func Lock() error {
	driver := store.GetDriver()
	return driver.Lock()
}

func Unlock() error {
	driver := store.GetDriver()
	return driver.Unlock()
}
