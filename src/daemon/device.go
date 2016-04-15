package daemon

import (
	"fmt"
	"net/http"
	"strconv"

	"api"
	"meta"
)

func (s *daemon) doDeviceGet(version string, w http.ResponseWriter, r *http.Request, objs map[string]string) error {
	metadata.Lock()
	defer metadata.Unlock()

	result := 0
	req := &api.DeviceGetRequest{}
	resp := &api.DeviceResponse{}

	for {
		if err := decodeRequest(r, req); err != nil {
			result = metadata.EcodeRequestDecodeError
			break
		}

		dv, err := metadata.GetDevice(req.ID, req.Backend)
		if err != nil {
			result = (err).(*metadata.Error).Code
			break
		}

		resp.ID = string(dv.Id)
		resp.IP = string(dv.Host)
		resp.Capacity = string(dv.Total)
		resp.Resource = string(dv.Identify)

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

func (s *daemon) doDeviceList(version string, w http.ResponseWriter, r *http.Request, objs map[string]string) error {
	metadata.Lock()
	defer metadata.Unlock()

	result := 0
	req := &api.DeviceListRequest{}
	resp := &api.DeviceListResponse{}

	for {
		if err := decodeRequest(r, req); err != nil {
			result = metadata.EcodeRequestDecodeError
			break
		}

		devices, err := metadata.ListDevices(req.Backend)
		if err != nil {
			result = (err).(*metadata.Error).Code
			break
		}
		fmt.Println(devices)
		resp.Devices = devices

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

func (s *daemon) doDeviceAdd(version string, w http.ResponseWriter, r *http.Request, objs map[string]string) error {
	metadata.Lock()
	defer metadata.Unlock()

	result := 0
	req := &api.DeviceAddRequest{}
	resp := &api.DeviceResponse{}

	for {
		if err := decodeRequest(r, req); err != nil {
			result = metadata.EcodeRequestDecodeError
			break
		}

		err := metadata.AddDevice(req.ID, req.Ip, req.Port, req.Total, req.Free, req.Status, req.Resource, req.Backend)
		if err != nil {
			result = (err).(*metadata.Error).Code
			break
		}

		resp.ID = req.ID
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

func (s *daemon) doDeviceDel(version string, w http.ResponseWriter, r *http.Request, objs map[string]string) error {
	metadata.Lock()
	defer metadata.Unlock()

	result := 0
	req := &api.DeviceDelRequest{}
	resp := &api.DeviceResponse{}

	for {
		if err := decodeRequest(r, req); err != nil {
			result = metadata.EcodeRequestDecodeError
			break
		}

		err := metadata.DelDevice(req.ID, req.Backend)
		if err != nil {
			result = (err).(*metadata.Error).Code
			break
		}

		resp.ID = req.ID
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
