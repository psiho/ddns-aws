package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/spf13/viper"
)

func listARecords() {
	// Build the request
	resp, err := client.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{
		HostedZoneId: &CLI.Records.List.ZoneId,
	})
	if err != nil {
		fmt.Printf("failed to list A records: %v", err)
		return
	}

	// Filter and output
	fmt.Println()
	fmt.Printf("Aliases in zone with ID %s:\n", CLI.Records.List.ZoneId)
	fmt.Println("---------------------------------------")
	for _, resource := range resp.ResourceRecordSets {
		if resource.Type == "A" {
			if recordIsActive(resource.Name) {
				fmt.Print("[] ")
			} else {
				fmt.Print("[ ] ")
			}
			fmt.Printf("%vs %s -> %s\n", *resource.TTL, *resource.Name, *resource.ResourceRecords[0].Value)
		}
	}
}

func fetchRecord(zoneID *string, name *string) (types.ResourceRecordSet, error) {
	var maxItems int32 = 1
	// Build the request
	resp, err := client.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{
		HostedZoneId:    zoneID,
		MaxItems:        &maxItems,
		StartRecordName: name,
		StartRecordType: "A",
	})
	if err != nil {
		fmt.Printf("Error fetching record: %v\n", err)
		return types.ResourceRecordSet{}, err
	}

	// Check if we got what we wwere looking for
	record := resp.ResourceRecordSets[0]
	if record.Type != "A" || *record.Name != *name {
		fmt.Printf("Record not found\n")
		return types.ResourceRecordSet{}, err
	}

	return record, nil
}

func activateRecord(zoneID *string, name *string) {
	if recordIsActive(name) {
		fmt.Printf("Record '%s' is already active. Nothing changed", *name)
		return
	}

	recordSet, err := fetchRecord(zoneID, name)
	if err != nil {
		return
	}

	conf.Active = append(conf.Active, Record{
		Name:   *recordSet.Name,
		ZoneID: *zoneID,
		TTL:    *recordSet.TTL,
		IP:     *recordSet.ResourceRecords[0].Value,
	})
	viper.Set("Active", &conf.Active)
	writeConfig()
	fmt.Printf("Record '%s' is now active\n", *name) //BUG:detect errors writing config
}

func deactivateRecord(name *string) {
	if !recordIsActive(name) {
		fmt.Printf("Record '%s' is not active. Nothing changed\n", *name)
		return
	}

	var new_arr []Record
	for _, item := range conf.Active {
		if item.Name != *name {
			new_arr = append(new_arr, item)
		}
	}
	conf.Active = new_arr
	viper.Set("Active", conf.Active)
	writeConfig()
	fmt.Printf("Record '%s' is deactivated\n", *name)
}

func updateRecordIP(name *string, ip *string) (string, error) {
	// only active records can be updated
	record, err := getRecordFromActive(name)
	if err != nil {
		if action != "server run" {
			fmt.Printf("Error updating record %s: %s", *name, err)
		}
		return "NOT_ACTIVE", err
	}

	if record.IP == *ip {
		if action != "server run" {
			fmt.Printf("No update needed. IP is the same (%s).", *ip)
		}
		return "NO_CHANGE", err
	}

	// Build the request
	resp, err := client.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: "UPSERT",
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: name,
						Type: "A",
						TTL:  &record.TTL,
						ResourceRecords: []types.ResourceRecord{
							{
								Value: ip,
							},
						},
					},
				},
			},
		},
		HostedZoneId: &record.ZoneID,
	})
	if err != nil {
		if action != "server run" {
			fmt.Printf("failed to update record IP. Error was:\n%v", err)
		}
		return "ROUTE53_UPDATE_FAIL", err
	}

	record.IP = *ip

	// fmt.Printf("BEFORE updateConfigFromStruct(%p) set conf to: %+v\n", &conf, conf)
	updateConfigActive()
	if action != "server run" {
		fmt.Printf("updated: %s -> %s (AWS status: %s)\n", *name, *ip, resp.ChangeInfo.Status)
	}
	return "UPDATE_OK", nil
}

func recordIsActive(name *string) bool {
	for _, activeRecord := range conf.Active {
		if *name == activeRecord.Name {
			return true
		}
	}
	return false
}

func getRecordFromActive(name *string) (*Record, error) {
	for i, activeRecord := range conf.Active {
		if *name == activeRecord.Name {
			return &conf.Active[i], nil
		}
	}
	return &Record{}, errors.New("Record not found in list of active records. Only active records can be updated.")
}
