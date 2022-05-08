package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
)

var conf Config
var client route53.Client
var action string

func main() {
	action = parseFlags()
	initConfig()
	loadConfig()
	client = initClient()

	// Execute action based on flags
	// log.Println("action was: ", action)
	switch action {
	case "records list <zone-id>":
		listARecords()
	case "records activate <zone-id> <resource-name>":
		activateRecord(&CLI.Records.Activate.ZoneId, &CLI.Records.Activate.ResourceName)
	case "records deactivate <resource-name>":
		deactivateRecord(&CLI.Records.Deactivate.ResourceName)
	case "records update <resource-name>", "records update <resource-name> <ip>":
		updateRecordIP(&CLI.Records.Update.ResourceName, &CLI.Records.Update.IP)

	case "zones list":
		listZones()

	case "server run":
		startConfigWatcher() //config changes are watched only in server mode
		startServer()
	}
}

func initClient() route53.Client {
	// load client config
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("unable to load SDK config: %v", err)
	}

	// Using the Config value, create the Route53 client
	return *route53.NewFromConfig(cfg)
}
