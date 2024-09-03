package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/miekg/dns"
)

var (
	ec2Client *ec2.Client
)

func main() {
	// Initialize AWS EC2 client
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	ec2Client = ec2.NewFromConfig(cfg)

	// Create a new DNS server
	server := &dns.Server{Addr: ":53", Net: "udp"}
	dns.HandleFunc(".", handleDNSRequest)

	log.Printf("Starting DNS proxy server")
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true

	// Get the client's IP address
	clientIP, _, err := net.SplitHostPort(w.RemoteAddr().String())
	if err != nil {
		log.Printf("Error getting client IP: %v", err)
		w.WriteMsg(m)
		return
	}

	// Determine the AZ of the client (ingress controller)
	clientAZ, err := getAZFromIP(clientIP)
	if err != nil {
		log.Printf("Error determining client AZ: %v", err)
		w.WriteMsg(m)
		return
	}

	for _, question := range r.Question {
		if question.Qtype == dns.TypeA {
			ips, err := resolveAndFilterIPs(question.Name, clientAZ)
			if err != nil {
				log.Printf("Error resolving %s: %v", question.Name, err)
				continue
			}

			for _, ip := range ips {
				rr, err := dns.NewRR(fmt.Sprintf("%s A %s", question.Name, ip))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
			}
		}
	}

	w.WriteMsg(m)
}

func resolveAndFilterIPs(domain, targetAZ string) ([]string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, err
	}

	var filteredIPs []string
	for _, ip := range ips {
		if ip.To4() != nil {
			if isInTargetAZ(ip.String(), targetAZ) {
				filteredIPs = append(filteredIPs, ip.String())
			}
		}
	}

	return filteredIPs, nil
}

func isInTargetAZ(ip, targetAZ string) bool {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: []ec2.Filter{
			{
				Name:   aws.String("addresses.private-ip-address"),
				Values: []string{ip},
			},
		},
	}

	result, err := ec2Client.DescribeNetworkInterfaces(context.Background(), input)
	if err != nil {
		log.Printf("Error describing network interfaces: %v", err)
		return false
	}

	for _, ni := range result.NetworkInterfaces {
		if ni.AvailabilityZone != nil && strings.EqualFold(*ni.AvailabilityZone, targetAZ) {
			return true
		}
	}

	return false
}

func getAZFromIP(ip string) (string, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: []ec2.Filter{
			{
				Name:   aws.String("addresses.private-ip-address"),
				Values: []string{ip},
			},
		},
	}

	result, err := ec2Client.DescribeNetworkInterfaces(context.Background(), input)
	if err != nil {
		return "", fmt.Errorf("error describing network interfaces: %v", err)
	}

	if len(result.NetworkInterfaces) > 0 && result.NetworkInterfaces[0].AvailabilityZone != nil {
		return *result.NetworkInterfaces[0].AvailabilityZone, nil
	}

	return "", fmt.Errorf("unable to determine AZ for IP %s", ip)
}
