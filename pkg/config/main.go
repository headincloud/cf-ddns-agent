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

package config

import (
	"flag"
	"os"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

type ProgramOptions struct {
	DiscoveryURL   string `validate:"required,http_url"`
	DiscoveryURLv6 string `validate:"required_if=Ipv6Enabled true,http_url"`
	Ipv6Enabled    bool
	DryRun         bool
	CreateMode     bool
	Domain         string `validate:"required"`
	Host           string `validate:"required,fqdn"`
	CfAPIToken     string
	Daemon         bool
	UpdateInterval int
}

var DefaultDiscoveryURL = "https://api.ipify.org"
var DefaultDiscoveryURLv6 = "https://api6.ipify.org"

func (Options *ProgramOptions) Init() {
	flag.StringVar(&Options.DiscoveryURL, "discovery-url", DefaultDiscoveryURL, "Specify an alternative IPv4 discovery service.")
	flag.StringVar(&Options.DiscoveryURLv6, "discovery-url-v6", DefaultDiscoveryURLv6, "Specify an alternative IPv6 discovery service.")
	flag.BoolVar(&Options.Ipv6Enabled, "ipv6", false, "Enable ipv6 support and CAA record updates, check README. (default \"false\")")
	flag.StringVar(&Options.Domain, "domain", "", "Specify the CloudFlare domain. (example: 'mydomain.org') - REQUIRED.")
	flag.StringVar(&Options.Host, "host", "", "Specify the full hostname (FQDN) that needs to be updated. (example: 'myhost.mydomain.org') - REQUIRED.")
	flag.StringVar(&Options.CfAPIToken, "cf-api-token", "", "Specify the CloudFlare API token. Using this parameter is discouraged, and the token should be specified in CF_API_TOKEN environment variable.")
	flag.BoolVar(&Options.Daemon, "daemon", false, "Run continuously in background and perform update every <x> minutes. (see 'update-interval')")
	flag.IntVar(&Options.UpdateInterval, "update-interval", 15, "Specify interval (in minutes) for updating the DNS record when running as a daemon. (see 'daemon'). A minimum of 5 is enforced.")
	flag.BoolVar(&Options.DryRun, "dry-run", false, "Run in \"dry-run\" mode, don't actually update the record. (default \"false\")")
	flag.BoolVar(&Options.CreateMode, "create", true, "Create record with 'auto' TTL if it doesn't exist yet, or generate error otherwise. (default \"true\")")
	flag.Parse()
}

func (Options *ProgramOptions) Validate() (err error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err = validate.Struct(Options); err != nil {
		return
	}

	log.Infof("IPv4 Discovery service url is set to: %s", Options.DiscoveryURL)
	if Options.Ipv6Enabled {
		log.Infof("IPv6 Discovery service url is set to: %s", Options.DiscoveryURLv6)
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
	return
}
