package flags

import (
	"github.com/codegangsta/cli"
)

var (
	DaemonFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Debug log, enabled by default",
		},
		cli.StringFlag{
			Name:  "log",
			Usage: "specific output log file, otherwise output to stdout by default",
		},
		cli.StringFlag{
			Name:  "root",
			Value: "/var/lib/policy",
			Usage: "specific root directory of policy, if configure file exists, daemon specific options would be ignore",
		},
		cli.StringSliceFlag{
			Name:  "hosts",
			Value: &cli.StringSlice{},
			Usage: "hosts to be scheduled, first host in the list would be treated as master node",
		},
		cli.StringSliceFlag{
			Name:  "policy-opts",
			Value: &cli.StringSlice{},
			Usage: "Options for hosts' policy",
		},
	}
)
