[![Build Status](https://github.com/headincloud/cf-ddns-agent/actions/workflows/build.yml/badge.svg?branch=develop)](https://github.com/headincloud/cf-ddns-agent/actions?query=branch%3Adevelop)

# cf-ddns-agent

## Description

`cf-ddns-agent` is a dynamic DNS update agent for CloudFlare DNS.

## Supported platforms and downloads

At the moment, there are executables for a mix of +- different os/cpu combinations (Including Windows, MacOS, Linux, *bsd, amd64, arm, arm64, mips64,...).

Program executables can be download from the [releases](https://github.com/headincloud/cf-ddns-agent/releases) page.


**Attention Windows users: It seems some anti-virus incorrectly classify the windows executable as malware. This could be related to the fact that the go compiler produces statically linked binaries.**


## Usage

````
 ./cf-ddns-agent
  -cf-api-token string
    	Specify the CloudFlare API token. Using this parameter is discouraged, and the token should be specified in CF_API_TOKEN environment variable.
  -create
    	Create record with 'auto' TTL if it doesn't exist yet, or generate error otherwise. (default "true") (default true)
  -daemon
    	Run continuously in background and perform update every <x> minutes. (see 'update-interval')
  -discovery-url string
    	Specify an alternative IPv4 discovery service. (default "https://api.ipify.org")
  -discovery-url-v6 string
    	Specify an alternative IPv6 discovery service. (default "https://api6.ipify.org")
  -domain string
    	Specify the CloudFlare domain. (example: 'mydomain.org') - REQUIRED.
  -dry-run
    	Run in "dry-run" mode, don't actually update the record. (default "false")
  -host string
    	Specify the full hostname (FQDN) that needs to be updated. (example: 'myhost.mydomain.org') - REQUIRED.
  -ipv6
    	Enable ipv6 support and CAA record updates, check README. (default "false")
  -update-interval int
    	Specify interval (in minutes) for updating the DNS record when running as a daemon. (see 'daemon'). A minimum of 5 is enforced. (default 15)
````
`CF_API_TOKEN` should be set with a token ([Cloudflare API Tokens Guide](https://developers.cloudflare.com/api/tokens/create)) that has the permissions to update the configured DNS zone. Using the "Edit zone DNS" template should be enough.

By default, the program will update the IP (if necessary) and then exit. If the update fails, error code 1 will be returned by the program. To run it continuously, use the `-daemon` and `-update-interval` parameters.

The `-discovery-url`and `-discovery-url-v6` parameters expect a URL that returns the IPv4/IPv6 address in plain-text, without any markup.

If setting the `CF_API_TOKEN` is not possible for some reason, it is possible to specify it on the command line using the `-cf-api-token` parameter. **This is discouraged as this is not very secure!**


## Timeouts and backoff/retry

- The http client is configured with 10-second timeout. 
- In case of timeout or a 5xx http error, the request is retried 3 times (each time with a longer delay between each attempt).
- In case of a 4xx http error, no retry occurs as this probably means a configuration issue (invalid discovery url supplied).


## Ipv6 support

This software now includes ipv6 support. you can enable it with the `-ipv6=true` parameter. This will update the AAAA-record for the specified host.

**Attention: Most operating systems use a temporary ipv6 address for outbound connections. This feature needs to be disabled in order for `cf-ddns-agent` to work correctly when IPv6 support is enabled! Check your operating system's manual.**


## License and copyright

### Main software

cf-ddns-agent
Copyright (C) 2020-2025 Jeroen Jacobs/Head In Cloud BV.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation version 3.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.

### Third-party software

The following third-party software is also directly included:

- sirupsen/logrus (c) Simon Eskildsen, MIT license. See: https://github.com/sirupsen/logrus
- cloudflare/cloudflare-go (c) CloudFlare, BSD license. See: https://github.com/cloudflare/cloudflare-go
- go-playground/validator, MIT license. See: https://github.com/go-playground/validator
