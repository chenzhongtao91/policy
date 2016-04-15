package client

import (
	"api"
	"strconv"
	"util"

	"github.com/codegangsta/cli"
)

var (
	VolumeCmds = cli.Command{
		Name:  "volume",
		Usage: "Manage volumes",
		Subcommands: []cli.Command{
			{
				Name:  "create",
				Usage: "Create New Volume",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name",
						Usage: "New volume name",
					},
					cli.StringFlag{
						Name:  "driver",
						Usage: "volume driver",
					},
					cli.StringFlag{
						Name:  "capacity",
						Usage: "volume capacity in G",
					},
				},
				Action: cmdCreateVolume,
			},

			{
				Name:  "inspect",
				Usage: "Inspect volume information",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name",
						Usage: "New volume name",
					},
					cli.StringFlag{
						Name:  "driver",
						Usage: "volume driver",
					},
				},
				Action: cmdGetVolume,
			},

			{
				Name:  "list",
				Usage: "List driver volume",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "driver",
						Usage: "volume driver",
					},
				},
				Action: cmdListVolume,
			},

			{
				Name:  "delete",
				Usage: "Delete Volume",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name",
						Usage: "New volume name",
					},
					cli.StringFlag{
						Name:  "driver",
						Usage: "volume driver",
					},
				},
				Action: cmdDeleteVolume,
			},

			{
				Name:  "attach",
				Usage: "attach volume to container",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name",
						Usage: "New volume name",
					},
					cli.StringFlag{
						Name:  "driver",
						Usage: "volume driver",
					},
					cli.StringFlag{
						Name:  "cid",
						Usage: "container id which use volume",
					},
					cli.StringFlag{
						Name:  "mode",
						Usage: "container id which use volume with mode",
					},
				},
				Action: cmdAttachVolume,
			},

			{
				Name:  "detach",
				Usage: "detach volume from container",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name",
						Usage: "New volume name",
					},
					cli.StringFlag{
						Name:  "driver",
						Usage: "volume driver",
					},
					cli.StringFlag{
						Name:  "cid",
						Usage: "container id which use volume",
					},
					cli.StringFlag{
						Name:  "mode",
						Usage: "container id which use volume with mode",
					},
				},
				Action: cmdDetachVolume,
			},
		},
	}
)

func cmdCreateVolume(c *cli.Context) {
	if err := doCreateVolume(c); err != nil {
		PrintErrorInfo(err)
	}
}

func getCapacity(c *cli.Context, err error) (int, error) {
	size, err := util.GetFlag(c, "capacity", true, err)
	if err != nil {
		return 0, err
	}
	return util.ParseSizeInMb(size)
}

func doCreateVolume(c *cli.Context) error {
	var err error

	volumeId, err := util.GetFlag(c, "name", true, err)
	driverName, err := util.GetFlag(c, "driver", true, err)
	capacity, err := getCapacity(c, err)
	if err != nil {
		return err
	}

	request := &api.VolumeCreateRequest{
		VolumeId:   volumeId,
		DriverName: driverName,
		Capacity:   strconv.Itoa(capacity),
	}

	url := "/volume/create"

	return sendRequestAndPrint("POST", url, request)
}

func cmdGetVolume(c *cli.Context) {
	if err := doGetVolume(c); err != nil {
		PrintErrorInfo(err)
	}
}

func doGetVolume(c *cli.Context) error {
	var err error

	volumeId, err := util.GetFlag(c, "name", true, err)
	driverName, err := util.GetFlag(c, "driver", true, err)
	if err != nil {
		return err
	}

	request := &api.VolumeGetRequest{
		VolumeId:   volumeId,
		DriverName: driverName,
	}

	url := "/volume/"

	return sendRequestAndPrint("GET", url, request)
}

func cmdListVolume(c *cli.Context) {
	if err := doListVolume(c); err != nil {
		PrintErrorInfo(err)
	}
}

func doListVolume(c *cli.Context) error {
	var err error

	driverName, err := util.GetFlag(c, "driver", true, err)
	if err != nil {
		return err
	}

	request := &api.VolumeListRequest{
		DriverName: driverName,
	}

	url := "/volume/list"

	return sendRequestAndPrint("GET", url, request)
}

func cmdDeleteVolume(c *cli.Context) {
	if err := doDeleteVolume(c); err != nil {
		PrintErrorInfo(err)
	}
}

func doDeleteVolume(c *cli.Context) error {
	var err error

	volumeId, err := util.GetFlag(c, "name", true, err)
	driverName, err := util.GetFlag(c, "driver", true, err)
	if err != nil {
		return err
	}

	request := &api.VolumeDeleteRequest{
		VolumeId:   volumeId,
		DriverName: driverName,
	}

	url := "/volume/"

	return sendRequestAndPrint("DELETE", url, request)
}

func cmdAttachVolume(c *cli.Context) {
	if err := doAttachVolume(c); err != nil {
		PrintErrorInfo(err)
	}
}

func doAttachVolume(c *cli.Context) error {
	var err error

	volumeId, err := util.GetFlag(c, "name", true, err)
	driverName, err := util.GetFlag(c, "driver", true, err)
	cid, err := util.GetFlag(c, "cid", true, err)
	mode, err := util.GetFlag(c, "mode", true, err)
	if err != nil {
		return err
	}

	request := &api.VolumeAttachRequest{
		VolumeId:    volumeId,
		DriverName:  driverName,
		ContainerId: cid,
		Mode:        mode,
	}

	url := "/volume/attach"

	return sendRequestAndPrint("POST", url, request)
}

func cmdDetachVolume(c *cli.Context) {
	if err := doDetachVolume(c); err != nil {
		PrintErrorInfo(err)
	}
}

func doDetachVolume(c *cli.Context) error {
	var err error

	volumeId, err := util.GetFlag(c, "name", true, err)
	driverName, err := util.GetFlag(c, "driver", true, err)
	cid, err := util.GetFlag(c, "cid", true, err)
	mode, err := util.GetFlag(c, "mode", true, err)
	if err != nil {
		return err
	}

	request := &api.VolumeAttachRequest{
		VolumeId:    volumeId,
		DriverName:  driverName,
		ContainerId: cid,
		Mode:        mode,
	}

	url := "/volume/detach"

	return sendRequestAndPrint("POST", url, request)
}
