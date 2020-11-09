package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"

	"github.com/cloudflare/cloudflare-go"
	log "github.com/sirupsen/logrus"
)

// will be replaced during build-phase with actual git-based version info
var Version = "local"
var DefaultDiscoveryURL = "https://api.ipify.org"

var options struct {

	DiscoveryURL	string
	Domain			string
	Host			string
	CfAPIToken		string
}

func init() {
	flag.StringVar(&options.DiscoveryURL, "discovery-url", DefaultDiscoveryURL, "Specify an alternative IPv4 discovery service.")
	flag.StringVar(&options.Domain,"domain", "", "Specify the CloudFlare domain. (example: 'mydomain.org') - REQUIRED")
	flag.StringVar(&options.Host,"host", "", "Specify the full A record name that needs to be updated. (example: 'myhost.mydomain.org') - REQUIRED")
	flag.StringVar(&options.CfAPIToken,"cf-api-token", "", "Specify the CloudFlare API token. Using this parameter is discouraged, and the token should be specified in CF_API_TOKEN environment variable.")
	flag.Parse()
}

const (
	AppName = "cf-ddns-agent - IP update-agent for CloudFlare DNS"
)

func main() {
	if len(os.Args) == 1 {
		flag.PrintDefaults()
		return
	}

	if flag.Arg(0) == "version" {
		fmt.Printf("%s/%s\n", path.Base(os.Args[0]), Version)
		return
	}

	log.Infof("%s (version %s) is starting...\n", AppName, Version)
	log.Infof("IP Discovery service url is set to: %s", options.DiscoveryURL)

	if options.Domain == "" || options.Host =="" {
		log.Fatal("Both --domain and --host  must be set!")
	}

	if options.CfAPIToken != "" {
		log.Warning("CloudFlare API token specified via command-line parameter instead of CF_API_TOKEN environment variable. This is insecure!")
	} else {
		options.CfAPIToken = os.Getenv("CF_API_TOKEN")
	}

	// set log format to include timestamp, even when TTY is attached.
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true},
	)

	// get ip
	resp, err := http.Get(options.DiscoveryURL)
	if err != nil {
		log.Errorf("Could not connect to IP discovery service: %s", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Could not read response from IP discovery service: %s", err.Error())
	}
	MyIP := net.ParseIP(string(body))
	if MyIP == nil {
		log.Errorf("Could not parse received value as an IP address: %s", string(body))
	}
	log.Infof("IP address received: %s", MyIP)

	api, err := cloudflare.NewWithAPIToken(options.CfAPIToken)
	if err != nil {
		log.Fatal(err)
	}

	// check current setting
	id, err := api.ZoneIDByName(options.Domain)

	foo := cloudflare.DNSRecord{
		Name: options.Host,
		Type: "A",
	}
	records, err := api.DNSRecords(id, foo)
	if err != nil {
		log.Errorf("Error encountered while checking current value of %s: %s", options.Host, err.Error())
	}
	for _, record := range records {
		CurrentIP := net.ParseIP(record.Content)
		if CurrentIP.Equal(MyIP) {
			log.Infof("IP address up to date for record %s (type %s). No DNS change necessary.", record.Name, record.Type)
		} else {
			log.Infof("Updating IP address of record %s (type %s) to %s", record.Name, record.Type, MyIP)
			record.Content = MyIP.String()
			err = api.UpdateDNSRecord(id, record.ID, record)
			if err != nil {
				log.Errorf("Error updating DNS record for %s (type %s) to %s: %s", record.Name, record.Type, MyIP, err.Error())
			} else {
				log.Infof("IP address of record record %s (type %s) successfully updated.", record.Name, record.Type)
			}
		}
	}
}
