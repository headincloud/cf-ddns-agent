package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/headincloud/cf-ddns-agent/pkg/config"
	"github.com/headincloud/cf-ddns-agent/pkg/discovery"
	"github.com/headincloud/cf-ddns-agent/pkg/util"
	log "github.com/sirupsen/logrus"
)

// will be replaced during build-phase with actual git-based version info
var Version = "local"
var Options config.ProgramOptions

const (
	AppName = "cf-ddns-agent - IP update-agent for CloudFlare DNS"
)

func main() {
	config.InitConfig(&Options)

	if len(os.Args) == 1 {
		flag.PrintDefaults()
		return
	}

	if flag.Arg(0) == "version" {
		fmt.Printf("%s/%s\n", path.Base(os.Args[0]), Version)
		return
	}

	log.Infof("%s (version %s) is starting...\n", AppName, Version)
	log.Infof("IP Discovery service url is set to: %s", Options.DiscoveryURL)

	if Options.Domain == "" || Options.Host == "" {
		log.Fatal("Both --domain and --host  must be set!")
	}

	if Options.CfAPIToken != "" {
		log.Warning("CloudFlare API token specified via command-line parameter instead of CF_API_TOKEN environment variable. This is insecure!")
	} else {
		Options.CfAPIToken = os.Getenv("CF_API_TOKEN")
	}

	// set log format to include timestamp, even when TTY is attached.
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true},
	)

	// if not running as daemon, we exit the program with an appropriate error-code
	if !Options.Daemon {
		err := PerformUpdate()
		defer Exit(err)
	} else {
		ticker := time.NewTicker(time.Duration(Options.UpdateInterval) * time.Minute)
		quit := make(chan struct{})
		for {
			select {
			case <-ticker.C:
				_ = PerformUpdate()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}
}

func Exit(err error) {
	if err != nil {
		defer os.Exit(1)
	}
}

func PerformUpdate() (err error) {
	// get ip and update
	MyIP, err := discovery.DiscoverIPv4(Options.DiscoveryURL)
	if err != nil {
		log.Errorf("An error was encountered during IP discovery. Check previous log entries for more details.")
	} else {
		err = util.PerformRecordUpdate(Options.CfAPIToken, Options.Domain, Options.Host, MyIP)
		if err != nil {
			log.Error("An error was encountered during updating of the DNS record. Check previous log entries for more details.")
		}
	}
	return
}
