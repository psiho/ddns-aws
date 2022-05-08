package main

import (
	"github.com/alecthomas/kong"
)

var CLI struct {
	Records struct {
		List struct {
			ZoneId string `arg:"" required:""`
		} `cmd:"" group:"Manage records:" help:"List all A records in the hosted zone"`

		Activate struct {
			ZoneId       string `arg:"" required:""`
			ResourceName string `arg:"" required:""`
		} `cmd:"" group:"Manage records:" help:"Activate record for DDNS updates"`

		Deactivate struct {
			ResourceName string `arg:"" required:""`
		} `cmd:"" group:"Manage records:" help:"Deactivate record for DDNS updates"`

		Update struct {
			ResourceName string `arg:"" required:""`
			IP           string `arg:"" optional:""`
		} `cmd:"" group:"Manage records:" help:"Manually update IP address of the record"`
	} `cmd:""`

	Zones struct {
		List struct{} `cmd:"" group:"Manage hosted zones:"`
	} `cmd:""`

	Server struct {
		Run struct{} `cmd:"" group:"Server: Run DDNS server"`
	} `cmd:""`
}

func parseFlags() string {
	ctx := kong.Parse(&CLI)
	return ctx.Command()
}
