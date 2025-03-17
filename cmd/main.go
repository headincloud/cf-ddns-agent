/*
 * Copyright (c) 2020-2025 Jeroen Jacobs/Head In Cloud BV.
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

package cmd

import (
	"context"
	"errors"
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
var Options *config.ProgramOptions

const (
	AppName = "cf-ddns-agent - IP update-agent for CloudFlare DNS"
)

func Execute() (err error) {
	Options = &config.ProgramOptions{}
	Options.Init()
	if len(os.Args) == 1 {
		fmt.Printf("%s/%s\n", path.Base(os.Args[0]), Version)
		flag.PrintDefaults()
		return nil
	}

	if flag.Arg(0) == "version" {
		fmt.Printf("%s/%s\n", path.Base(os.Args[0]), Version)
		return nil
	}

	// set log format to include timestamp, even when TTY is attached.
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true},
	)

	log.Infof("%s (version %s) is starting...\n", AppName, Version)

	// validate configuration
	if err = Options.Validate(); err != nil {
		return
	}

	// signal handler for our application
	sigChannel := make(chan os.Signal, 2)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)

	// create our cancel-context to handle ctrl-c
	ctx, cancelFunc := context.WithCancel(context.Background())

	// make sure our cancel function gets at least once
	defer func() {
		signal.Stop(sigChannel)
		close(sigChannel) // make sure no more writes happen after we stopped signaling
		cancelFunc()
	}()

	// run our handler in a separate go routine
	go func() {
		signalReceived := false
		for sig := range sigChannel {
			if signalReceived {
				// if we get here, it means another ctrl-c was received.
				log.Errorf("Forced shutdown!")
				os.Exit(1)
			}
			signalReceived = true
			log.Infof("Received %s, trying exiting gracefully. Press CTRL-C again to force shutdown.", sig)
			cancelFunc()
		}
	}()

	// if not running as daemon, we exit the program with an appropriate error-code
	if !Options.Daemon {
		err = PerformUpdate(ctx)
		if err != nil {
			// We do not threat our program being killed by ctrl-c as an error
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				err = nil
			}
		}

	} else {
		// let's do an update at daemon startup
		err = PerformUpdate(ctx)
		// now start our timer
		ticker := time.NewTicker(time.Duration(Options.UpdateInterval) * time.Minute)
		for {
			select {
			case <-ticker.C:
				err = PerformUpdate(ctx)
			case <-ctx.Done():
				log.Infof("Shutdown complete.")
				return
			}
		}
	}
	return
}

func PerformUpdate(ctx context.Context) (err error) {
	// get ip and update
	MyIPv4, err := discovery.DiscoverIPv4(ctx, Options.DiscoveryURL)
	if err != nil {
		log.Errorf("An error was encountered during IPv4 discovery. Check previous log entries for more details.")
	} else {
		err = util.UpdateCFRecord(ctx, Options, "A", MyIPv4)
		if err != nil {
			log.Error("An error was encountered during updating of the DNS A-record. Check previous log entries for more details.")
		}
	}
	if Options.Ipv6Enabled {
		MyIPv6, err := discovery.DiscoverIPv6(ctx, Options.DiscoveryURLv6)
		if err != nil {
			log.Errorf("An error was encountered during IPv6 discovery. Check previous log entries for more details.")
		} else {
			err = util.UpdateCFRecord(ctx, Options, "AAAA", MyIPv6)
			if err != nil {
				log.Error("An error was encountered during updating of the DNS AAAA-record. Check previous log entries for more details.")
			}
		}
	}
	return
}
