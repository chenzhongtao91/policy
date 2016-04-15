package client

import (
	"api"
	"util"

	"github.com/codegangsta/cli"
)

var (
	HostCmds = cli.Command{
		Name:  "host",
		Usage: "Manage hosts",
		Subcommands: []cli.Command{
			{
				Name:  "add",
				Usage: "add new host to cluster storage pool",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "ip",
						Usage: "host ip",
					},
				},
				Action: cmdAddHost,
			},

			{
				Name:  "inspect",
				Usage: "Inspect host information",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "ip",
						Usage: "host ip",
					},
				},
				Action: cmdGetHost,
			},

			{
				Name:   "list",
				Usage:  "List Host",
				Action: cmdListHost,
			},

			{
				Name:  "delete",
				Usage: "Delete Host",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "ip",
						Usage: "host ip",
					},
				},
				Action: cmdDeleteHost,
			},
		},
	}
)

func cmdGetHost(c *cli.Context) {
	if err := doGetHost(c); err != nil {
		PrintErrorInfo(err)
	}
}

func doGetHost(c *cli.Context) error {
	var err error

	ip, err := util.GetFlag(c, "ip", true, err)
	if err != nil {
		return err
	}

	request := &api.HostGetRequest{
		Ip: ip,
	}

	url := "/host/"

	return sendRequestAndPrint("GET", url, request)
}

func cmdListHost(c *cli.Context) {
	if err := doListHost(c); err != nil {
		PrintErrorInfo(err)
	}
}

func doListHost(c *cli.Context) error {
	request := &api.HostListRequest{
		Ip: "000000",
	}

	url := "/host/list"

	return sendRequestAndPrint("GET", url, request)
}

func cmdAddHost(c *cli.Context) {
	if err := doAddHost(c); err != nil {
		PrintErrorInfo(err)
	}
}

func doAddHost(c *cli.Context) error {
	var err error

	ip, err := util.GetFlag(c, "ip", true, err)
	if err != nil {
		return err
	}

	request := &api.HostAddRequest{
		Ip: ip,
	}

	url := "/host/add"

	return sendRequestAndPrint("POST", url, request)
}

func cmdDeleteHost(c *cli.Context) {
	if err := doDeleteHost(c); err != nil {
		PrintErrorInfo(err)
	}
}

func doDeleteHost(c *cli.Context) error {
	var err error

	ip, err := util.GetFlag(c, "ip", true, err)
	if err != nil {
		return err
	}

	request := &api.HostDeleteRequest{
		Ip: ip,
	}

	url := "/host/"

	return sendRequestAndPrint("DELETE", url, request)
}
