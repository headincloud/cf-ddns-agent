/*
 * Copyright (c) 2020 Jeroen Jacobs/Head In Cloud BV.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as published by
 * the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"
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

	if Options.Daemon && (Options.UpdateInterval < 5) {
		log.Warnf("Update interval is set too low. It has been set to 5 minutes.")
		Options.UpdateInterval = 5
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
		// let's do an update at daemon startup
		_ = PerformUpdate()
		// now start our timer and ctrl-c handler
		ticker := time.NewTicker(time.Duration(Options.UpdateInterval) * time.Minute)
		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		for {
			select {
			case <-ticker.C:
				_ = PerformUpdate()
			case sig := <-quit:
				log.Infof("Received %s, exiting gracefully...", sig)
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
		err = util.PerformRecordUpdate(Options.CfAPIToken, Options.Domain, Options.Host, MyIP, nil)
		if err != nil {
			log.Error("An error was encountered during updating of the DNS record. Check previous log entries for more details.")
		}
	}
	return
}
