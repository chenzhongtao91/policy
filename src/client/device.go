package client

import (
	"fmt"

	"api"
	"util"

	"github.com/codegangsta/cli"
)

var (
	DeviceCmds = cli.Command{
		Name:  "device",
		Usage: "Manage devices",
		Subcommands: []cli.Command{
			{
				Name:  "add",
				Usage: "add new device to cluster storage pool",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "id",
						Usage: "device id or unique name",
					},
					cli.StringFlag{
						Name:  "ip",
						Usage: "device host ip",
					},
					cli.StringFlag{
						Name:  "port",
						Usage: "devcie daemon port",
					},

					cli.StringFlag{
						Name:  "total",
						Usage: "device total capacity",
					},
					cli.StringFlag{
						Name:  "status",
						Usage: "device status",
					},
					cli.StringFlag{
						Name:  "resource",
						Usage: "device storage resource id:  iqn",
					},
					cli.StringFlag{
						Name:  "backend",
						Usage: "device backend storage",
					},
				},
				Action: cmdAddDevice,
			},

			{
				Name:  "inspect",
				Usage: "Inspect device information",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "id",
						Usage: "device id or unique name",
					},
					cli.StringFlag{
						Name:  "ip",
						Usage: "device host ip",
					},
					cli.StringFlag{
						Name:  "port",
						Usage: "devcie daemon port",
					},
					cli.StringFlag{
						Name:  "backend",
						Usage: "device backend storage",
					},
				},
				Action: cmdGetDevice,
			},

			{
				Name:  "list",
				Usage: "List driver volume",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "backend",
						Usage: "device backend storage",
					},
				},
				Action: cmdListDevice,
			},

			{
				Name:  "delete",
				Usage: "Delete Volume",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "id",
						Usage: "device id or unique name",
					},
					cli.StringFlag{
						Name:  "ip",
						Usage: "device host ip",
					},
					cli.StringFlag{
						Name:  "port",
						Usage: "devcie daemon port",
					},
					cli.StringFlag{
						Name:  "backend",
						Usage: "device backend storage",
					},
				},
				Action: cmdDeleteDevice,
			},
		},
	}
)

func cmdGetDevice(c *cli.Context) {
	if err := doGetDevice(c); err != nil {
		PrintErrorInfo(err)
	}
}

func getInteger(c *cli.Context, key string, force bool, err error) (int, error) {
	size, err := util.GetFlag(c, key, force, err)
	if err != nil {
		return 0, err
	}
	return util.ParseInt(size)
}

func getSize(c *cli.Context, key string, force bool, err error) (int, error) {
	size, err := util.GetFlag(c, key, force, err)
	if err != nil {
		return 0, err
	}
	return util.ParseSizeInMb(size)
}

func doGetDevice(c *cli.Context) error {
	var err error

	id, err := util.GetFlag(c, "id", true, err)
	ip, err := util.GetFlag(c, "ip", true, err)
	port, err := getInteger(c, "port", true, err)
	backend, err := util.GetFlag(c, "backend", true, err)
	if err != nil {
		return err
	}

	request := &api.DeviceGetRequest{
		ID:      id,
		Ip:      ip,
		Port:    port,
		Backend: backend,
	}

	url := "/device/"

	return sendRequestAndPrint("GET", url, request)
}

func cmdListDevice(c *cli.Context) {
	if err := doListDevice(c); err != nil {
		PrintErrorInfo(err)
	}
}

func doListDevice(c *cli.Context) error {
	var err error
	backend, err := util.GetFlag(c, "backend", true, err)
	if err != nil {
		return err
	}
	request := &api.DeviceGetRequest{
		ID:      "000000",
		Backend: backend,
	}

	url := "/device/list"

	return sendRequestAndPrint("GET", url, request)
}

func cmdAddDevice(c *cli.Context) {
	if err := doAddDevice(c); err != nil {
		PrintErrorInfo(err)
	}
}

func doAddDevice(c *cli.Context) error {
	var err error

	id, err := util.GetFlag(c, "id", true, err)
	ip, err := util.GetFlag(c, "ip", true, err)
	port, err := getInteger(c, "port", true, err)
	total, err := getSize(c, "total", true, err)
	resource, err := util.GetFlag(c, "resource", true, err)
	free, err := getSize(c, "free", false, err)
	status, err := getInteger(c, "status", false, err)
	backend, err := util.GetFlag(c, "backend", true, err)
	if err != nil {
		return err
	}

	if port == 0 || total == 0 {
		return fmt.Errorf("Parameters Error.")
	}

	request := &api.DeviceAddRequest{
		ID:       id,
		Ip:       ip,
		Port:     port,
		Total:    total,
		Free:     free,
		Status:   status,
		Resource: resource,
		Backend:  backend,
	}

	url := "/device/add"

	return sendRequestAndPrint("POST", url, request)
}

func cmdDeleteDevice(c *cli.Context) {
	if err := doDeleteDevice(c); err != nil {
		PrintErrorInfo(err)
	}
}

func doDeleteDevice(c *cli.Context) error {
	var err error

	id, err := util.GetFlag(c, "id", true, err)
	ip, err := util.GetFlag(c, "ip", true, err)
	port, err := getInteger(c, "port", false, err)
	backend, err := util.GetFlag(c, "backend", true, err)
	if err != nil {
		return err
	}

	request := &api.DeviceGetRequest{
		ID:      id,
		Ip:      ip,
		Port:    port,
		Backend: backend,
	}

	url := "/device/"

	return sendRequestAndPrint("DELETE", url, request)
}
