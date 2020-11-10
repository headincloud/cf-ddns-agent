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

package config

import "flag"

type ProgramOptions struct {
	DiscoveryURL   string
	Domain         string
	Host           string
	CfAPIToken     string
	Daemon         bool
	UpdateInterval int
}

var DefaultDiscoveryURL = "https://api.ipify.org"

func InitConfig(Options *ProgramOptions) {
	flag.StringVar(&Options.DiscoveryURL, "discovery-url", DefaultDiscoveryURL, "Specify an alternative IPv4 discovery service.")
	flag.StringVar(&Options.Domain, "domain", "", "Specify the CloudFlare domain. (example: 'mydomain.org') - REQUIRED")
	flag.StringVar(&Options.Host, "host", "", "Specify the full A record name that needs to be updated. (example: 'myhost.mydomain.org') - REQUIRED")
	flag.StringVar(&Options.CfAPIToken, "cf-api-token", "", "Specify the CloudFlare API token. Using this parameter is discouraged, and the token should be specified in CF_API_TOKEN environment variable.")
	flag.BoolVar(&Options.Daemon, "daemon", false, "Run continuously in background and perform update every <x> minutes. (see 'update-interval')")
	flag.IntVar(&Options.UpdateInterval, "update-interval", 30, "Specify interval (in minutes) for updating the DNS record when running as a daemon. (see 'daemon')")
	flag.Parse()
}
