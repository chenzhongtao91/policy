package daemon

import (
	"fmt"
	"net/http"
	"strconv"

	"api"
	"meta"
	"meta/proto"
	"scheduler"
)

func (s *daemon) doVolumeList(version string, w http.ResponseWriter, r *http.Request, objs map[string]string) error {
	metadata.Lock()
	defer metadata.Unlock()

	result := 0
	req := &api.VolumeListRequest{}
	resp := &api.VolumeListResponse{}

	for {
		if err := decodeRequest(r, req); err != nil {
			result = metadata.EcodeRequestDecodeError
			break
		}

		volumes, err := metadata.ListVolumes(req.DriverName)
		if err != nil {
			result = (err).(*metadata.Error).Code
			break
		}

		resp.Volumes = volumes
		break
	}

	resp.Result = strconv.Itoa(result)

	data, err := api.ResponseOutput(*resp)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

func (s *daemon) doVolumeGet(version string, w http.ResponseWriter, r *http.Request, objs map[string]string) error {
	metadata.Lock()
	defer metadata.Unlock()

	result := 0
	req := &api.VolumeGetRequest{}
	resp := &api.VolumeResponse{}
	for {
		if err := decodeRequest(r, req); err != nil {
			result = metadata.EcodeRequestDecodeError
			break
		}

		vl, err := metadata.GetVolume(req.VolumeId, req.DriverName)
		if err != nil {
			result = (err).(*metadata.Error).Code
			break
		}

		var ownerContainers []string
		for i := 0; i < len(vl.Containers); i++ {
			ownerContainers = append(ownerContainers, string(vl.Containers[i].Containerid))
		}

		fmt.Println(vl)

		resp.ID = string(vl.Id)
		resp.Status = string(vl.Status)
		resp.Capacity = string(vl.Capacity)
		resp.Writable = string(vl.Writable)
		resp.Containers = ownerContainers

		break
	}

	resp.Result = strconv.Itoa(result)
	data, err := api.ResponseOutput(*resp)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

func (s *daemon) doVolumeCreate(version string, w http.ResponseWriter, r *http.Request, objs map[string]string) error {
	metadata.Lock()
	defer metadata.Unlock()

	fmt.Println("[doVolumeCreate] ", r)

	result := 0
	req := &api.VolumeCreateRequest{}
	resp := &api.VolumeResponse{}

	for {
		if err := decodeRequest(r, req); err != nil {
			result = metadata.EcodeRequestDecodeError
			break
		}

		//TODO 通过scheduler返回合适的devices， 记录到Volume结构中，并返回
		var opts = map[string]string{"FilterCapacity": req.Capacity, "WeigherCapacity": "100", "Backend": req.DriverName, "Replica": "1"}
		ds, err := scheduler.DoScheduler(opts)
		if err != nil {
			result = metadata.EcodeSchedulerError
			break
		}

		devs := []api.DeviceIdentify{}                //存在Device结构中的后端信息
		devices := []*metaproto.Volume_AttachDevice{} //更新Volume结构的device信息
		for i := 0; i < len(ds); i++ {
			d := api.DeviceIdentify{
				IP:   string(ds[i].Host),
				Port: string(ds[i].Port),
				Dev:  string(ds[i].Identify),
			}
			devs = append(devs, d)

			device := &metaproto.Volume_AttachDevice{
				Deviceid: ds[i].Id,
				Status:   ds[i].Status,
			}
			devices = append(devices, device)
		}

		fmt.Println("[daemon] ", req)

		vl := &metaproto.Volume{
			Id:       []byte(req.VolumeId),
			Capacity: []byte(req.Capacity),
			Devices:  devices,
		}

		fmt.Println("[daemon] ", vl)

		cons := []*metaproto.Volume_OwnerContainer{}
		oc := &metaproto.Volume_OwnerContainer{Containerid: []byte(req.ContainerId)}

		if len(req.ContainerId) > 0 {
			if req.Mode == metadata.RWVolume {
				vl.Writable = []byte(req.ContainerId)
				oc.Mode = []byte(metadata.RWVolume)
				cons = append(cons, oc)
				vl.Containers = cons
			}

			if req.Mode == metadata.ROVolume {
				oc.Mode = []byte(metadata.ROVolume)
				cons = append(cons, oc)
				vl.Containers = cons
			}
		}

		fmt.Println("Volume struct: ", vl)
		err = metadata.AddVolume(vl, req.DriverName)
		if err != nil {
			result = (err).(*metadata.Error).Code
			break
		}

		resp.ID = req.VolumeId
		resp.Status = string(metadata.VOLUME_UNKNOW)
		resp.Devices = devs
		break
	}

	fmt.Println("Response: ", resp)
	resp.Result = strconv.Itoa(result)
	data, err := api.ResponseOutput(*resp)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

func (s *daemon) doVolumeAttach(version string, w http.ResponseWriter, r *http.Request, objs map[string]string) error {
	metadata.Lock()
	defer metadata.Unlock()

	result := 0
	req := &api.VolumeAttachRequest{}
	resp := &api.VolumeResponse{}

	for {
		if err := decodeRequest(r, req); err != nil {
			result = metadata.EcodeRequestDecodeError
			break
		}

		mode := []byte(metadata.ROVolume)
		if req.Mode == "rw" || req.Mode == "RW" {
			mode = []byte(metadata.RWVolume)
		} else if req.Mode == "ro" || req.Mode == "RO" {
			mode = []byte(metadata.ROVolume)
		} else {
			result = 9999
			break
		}

		oc := &metaproto.Volume_OwnerContainer{
			Containerid: []byte(req.ContainerId),
			Mode:        mode,
		}

		err := metadata.SetVolumeContainer(req.VolumeId, oc, req.DriverName, false)
		if err != nil {
			result = (err).(*metadata.Error).Code
			break
		}

		resp.ID = req.VolumeId
		resp.Status = string(metadata.VOLUME_INUSE)
		break
	}

	resp.Result = strconv.Itoa(result)
	data, err := api.ResponseOutput(*resp)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

func (s *daemon) doVolumeDetach(version string, w http.ResponseWriter, r *http.Request, objs map[string]string) error {
	metadata.Lock()
	defer metadata.Unlock()

	result := 0
	req := &api.VolumeDetachRequest{}
	resp := &api.VolumeResponse{}

	for {
		if err := decodeRequest(r, req); err != nil {
			result = metadata.EcodeRequestDecodeError
			break
		}

		err := metadata.DelVolumeContainer(req.VolumeId, req.DriverName, req.ContainerId)
		if err != nil {
			result = (err).(*metadata.Error).Code
			break
		}

		resp.ID = req.VolumeId
		resp.Status = string(metadata.VOLUME_INUSE)
		break

	}

	resp.Result = strconv.Itoa(result)
	data, err := api.ResponseOutput(*resp)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

func (s *daemon) doVolumeDelete(version string, w http.ResponseWriter, r *http.Request, objs map[string]string) error {
	metadata.Lock()
	defer metadata.Unlock()

	result := 0
	req := &api.VolumeDeleteRequest{}
	resp := &api.VolumeResponse{}

	for {

		if err := decodeRequest(r, req); err != nil {
			result = metadata.EcodeRequestDecodeError
			break
		}

		//TODO  设置volume的devices为READY状态，释放devices空间

		err := metadata.DelVolume(req.VolumeId, req.DriverName)
		if err != nil {
			result = (err).(*metadata.Error).Code
			break
		}

		resp.ID = req.VolumeId
		break

	}

	resp.Result = strconv.Itoa(result)

	data, err := api.ResponseOutput(*resp)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}
