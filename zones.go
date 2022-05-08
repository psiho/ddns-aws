package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/route53"
)

func listZones() {
	// Build the request
	resp, err := client.ListHostedZones(context.TODO(), &route53.ListHostedZonesInput{})
	if err != nil {
		fmt.Printf("failed to list hosted zones: %v", err)
		return
	}

	// Filter and output
	fmt.Println()
	fmt.Println("Hosted zones:")
	fmt.Println("---------------------------------------")
	for _, zone := range resp.HostedZones {
		fmt.Printf("%s %s\n", *zone.Id, *zone.Name)
	}
}
