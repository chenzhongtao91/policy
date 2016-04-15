package daemon

import (
	"net/http"
	"strconv"

	"api"
	"meta"
)

func (s *daemon) doHostGet(version string, w http.ResponseWriter, r *http.Request, objs map[string]string) error {
	metadata.Lock()
	defer metadata.Unlock()

	result := 0
	req := &api.HostGetRequest{}
	resp := &api.HostResponse{}

	for {
		if err := decodeRequest(r, req); err != nil {
			result = metadata.EcodeRequestDecodeError
			break
		}

		hs, err := metadata.GetHost(req.Ip)
		if err != nil {
			result = (err).(*metadata.Error).Code
			break
		}

		resp.IP = req.Ip
		resp.Status = string(hs.Status)
		for i := 0; i < len(hs.Devices); i++ {
			resp.Devs = append(resp.Devs, string(hs.Devices[i]))
		}
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

func (s *daemon) doHostList(version string, w http.ResponseWriter, r *http.Request, objs map[string]string) error {
	metadata.Lock()
	defer metadata.Unlock()

	result := 0
	req := &api.HostListRequest{}
	resp := &api.HostListResponse{}

	for {
		if err := decodeRequest(r, req); err != nil {
			result = metadata.EcodeRequestDecodeError
			break
		}

		hosts, err := metadata.ListHostsName()
		if err != nil {
			result = (err).(*metadata.Error).Code
			break
		}

		resp.IPs = hosts

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

func (s *daemon) doHostAdd(version string, w http.ResponseWriter, r *http.Request, objs map[string]string) error {
	metadata.Lock()
	defer metadata.Unlock()

	result := 0
	req := &api.HostAddRequest{}
	resp := &api.HostResponse{}

	for {
		if err := decodeRequest(r, req); err != nil {
			result = metadata.EcodeRequestDecodeError
			break
		}

		//TODO 检查Host是否已经存在

		devs := [][]byte{}
		err := metadata.AddHost(req.Ip, metadata.HOST_ONLINE, devs)
		if err != nil {
			result = (err).(*metadata.Error).Code
			break
		}
		resp.IP = req.Ip
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

func (s *daemon) doHostDel(version string, w http.ResponseWriter, r *http.Request, objs map[string]string) error {
	metadata.Lock()
	defer metadata.Unlock()

	result := 0
	req := &api.HostListRequest{}
	resp := &api.HostResponse{}

	for {
		if err := decodeRequest(r, req); err != nil {
			result = metadata.EcodeRequestDecodeError
			break
		}

		//检查Host中的Device是否有正在被使用的情况

		err := metadata.DelHost(req.Ip)
		if err != nil {
			result = (err).(*metadata.Error).Code
			break
		}

		resp.IP = req.Ip
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
