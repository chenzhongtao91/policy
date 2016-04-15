package client

import (
	"fmt"
	"io/ioutil"

	"client/flags"
	"daemon"

	"github.com/codegangsta/cli"
)

var (
	daemonCmd = cli.Command{
		Name:   "daemon",
		Usage:  "start policy daemon",
		Flags:  flags.DaemonFlags,
		Action: cmdStartDaemon,
	}

	infoCmd = cli.Command{
		Name:   "info",
		Usage:  "information about policy",
		Action: cmdInfo,
	}
)

func cmdInfo(c *cli.Context) {
	if err := doInfo(c); err != nil {
		panic(err)
	}
}

func doInfo(c *cli.Context) error {
	rc, _, err := client.call("GET", "/info", nil, nil)
	if err != nil {
		return err
	}
	defer rc.Close()

	b, err := ioutil.ReadAll(rc)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func cmdStartDaemon(c *cli.Context) {
	if err := startDaemon(c); err != nil {
		panic(err)
	}
}

func startDaemon(c *cli.Context) error {
	return daemon.Start(client.addr, c)
}
