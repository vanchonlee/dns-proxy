package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/miekg/dns"
)

type AZConfig struct {
	AZs map[string][]string `json:"azs"`
}

var (
	azConfig AZConfig
)

func main() {
	// Load AZ configuration
	err := loadAZConfig("az_config.json")
	if err != nil {
		log.Fatalf("Failed to load AZ configuration: %v", err)
	}

	// Create a new DNS server
	server := &dns.Server{Addr: ":53", Net: "udp"}
	dns.HandleFunc(".", handleDNSRequest)

	log.Printf("Starting DNS proxy server")
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func loadAZConfig(filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

	err = json.Unmarshal(file, &azConfig)
	if err != nil {
		return fmt.Errorf("error unmarshaling config: %v", err)
	}

	return nil
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
	clientAZ := getAZFromIP(clientIP)
	if clientAZ == "" {
		log.Printf("Unable to determine AZ for IP %s", clientIP)
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
	return getAZFromIP(ip) == targetAZ
}

func getAZFromIP(ip string) string {
	for az, ranges := range azConfig.AZs {
		for _, cidr := range ranges {
			_, ipNet, err := net.ParseCIDR(cidr)
			if err != nil {
				log.Printf("Error parsing CIDR %s: %v", cidr, err)
				continue
			}
			if ipNet.Contains(net.ParseIP(ip)) {
				return az
			}
		}
	}
	return ""
}
