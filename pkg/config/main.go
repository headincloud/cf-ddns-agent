package config

import "flag"

type ProgramOptions struct {

	DiscoveryURL	string
	Domain			string
	Host			string
	CfAPIToken		string
	Daemon			bool
	UpdateInterval	int

}

var DefaultDiscoveryURL = "https://api.ipify.org"

func InitConfig(Options *ProgramOptions) {
	flag.StringVar(&Options.DiscoveryURL, "discovery-url", DefaultDiscoveryURL, "Specify an alternative IPv4 discovery service.")
	flag.StringVar(&Options.Domain,"domain", "", "Specify the CloudFlare domain. (example: 'mydomain.org') - REQUIRED")
	flag.StringVar(&Options.Host,"host", "", "Specify the full A record name that needs to be updated. (example: 'myhost.mydomain.org') - REQUIRED")
	flag.StringVar(&Options.CfAPIToken,"cf-api-token", "", "Specify the CloudFlare API token. Using this parameter is discouraged, and the token should be specified in CF_API_TOKEN environment variable.")
	flag.BoolVar(&Options.Daemon, "daemon", false, "Run continuously in background and perform update every <x> minutes. (see 'update-interval')")
	flag.IntVar(&Options.UpdateInterval, "update-interval", 30, "Specify interval (in minutes) for updating the DNS record when running as a daemon. (see 'daemon')")
	flag.Parse()
}

