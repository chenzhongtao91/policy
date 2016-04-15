package api

type VolumeGetRequest struct {
	VolumeId   string
	DriverName string
}

type VolumeCreateRequest struct {
	VolumeId    string
	DriverName  string
	Capacity    string
	ContainerId string
	Mode        string
}

type VolumeListRequest struct {
	DriverName string
}

type VolumeAttachRequest struct {
	VolumeId    string
	DriverName  string
	ContainerId string
	Mode        string
}

type VolumeDetachRequest struct {
	VolumeId    string
	DriverName  string
	ContainerId string
	Mode        string
}

type VolumeDeleteRequest struct {
	VolumeId   string
	DriverName string
}

type HostAddRequest struct {
	Ip string
}

type HostGetRequest struct {
	Ip string
}

type HostListRequest struct {
	Ip string
}

type HostDeleteRequest struct {
	Ip string
}

type DeviceAddRequest struct {
	ID       string
	Ip       string
	Port     int
	Total    int
	Free     int
	Status   int
	Resource string
	Backend  string
}

type DeviceGetRequest struct {
	ID      string
	Ip      string
	Port    int
	Backend string
}

type DeviceListRequest struct {
	ID      string
	Backend string
}

type DeviceDelRequest struct {
	ID      string
	Ip      string
	Port    int
	Backend string
}
