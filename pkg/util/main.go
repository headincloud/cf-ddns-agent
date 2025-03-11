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

package util

import (
	"context"
	"net"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/dns"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/cloudflare/cloudflare-go/v4/zones"
	log "github.com/sirupsen/logrus"
)

func UpdateCFRecord(ctx context.Context, token string, domain string, host string, recordType string, ip net.IP, dryRun bool, createMode bool) (err error) {
	// create our client
	api := cloudflare.NewClient(option.WithAPIToken(token))

	// check current setting
	zoneList, err := api.Zones.List(ctx, zones.ZoneListParams{
		Name:   cloudflare.F(domain),
		Status: cloudflare.F(zones.ZoneListParamsStatusActive),
		Match:  cloudflare.F(zones.ZoneListParamsMatchAll),
	})

	if err != nil {
		log.Errorf("Failed to retrieve zones for %s: %s", domain, err.Error())
		return
	}

	if len(zoneList.Result) == 0 {
		log.Errorf("Zone not found: %s", domain)
		return
	}

	id := zoneList.Result[0].ID

	// let's find our record
	validTypes := make(map[string]dns.RecordListParamsType)
	validTypes["A"] = dns.RecordListParamsTypeA
	validTypes["AAAA"] = dns.RecordListParamsTypeAAAA

	recordList, err := api.DNS.Records.List(ctx, dns.RecordListParams{
		ZoneID: cloudflare.F(id),
		Type:   cloudflare.F(validTypes[recordType]),
		Name: cloudflare.F(dns.RecordListParamsName{
			Exact: cloudflare.F(host),
		}),
		Match: cloudflare.F(dns.RecordListParamsMatchAll),
	})

	if err != nil {
		log.Errorf("Failed to retrieve record for %s: %s", host, err.Error())
		return
	}

	if len(zoneList.Result) == 0 {
		if !createMode {
			log.Errorf("Record not found for %s", host)
			return
		} else {
			// create record
			log.Infof("No record found for %s (type %s). Will attempt to create one...", host, recordType)
			if !dryRun {
				_, err = api.DNS.Records.New(ctx, dns.RecordNewParams{
					ZoneID: cloudflare.F(id),
					Record: dns.RecordParam{
						Name:    cloudflare.F(host),
						Type:    cloudflare.F(dns.RecordType(validTypes[recordType])),
						TTL:     cloudflare.F(dns.TTL(1)), // 1 = automatic
						Content: cloudflare.F(ip.String()),
						Proxied: cloudflare.F(true),
					},
				})
				if err != nil {
					log.Errorf("Error encountered while creating record for %s: %s", host, err.Error())
					return
				}
				log.Infof("Record created for %s: %s", host, recordType)
				return
			} else {
				log.Infof("Skipped creation of DNS record. (dry-run mode active)")
			}
		}
	} else {
		// update record
		record := recordList.Result[0]
		CurrentIP := net.ParseIP(record.Content)
		if CurrentIP.Equal(ip) {
			log.Infof("IP address up to date for record %s (type %s). No DNS change necessary.", record.Name, record.Type)
		} else {
			log.Infof("Updating IP address of record %s (type %s) to %s", record.Name, record.Type, ip)
			if !dryRun {
				_, err = api.DNS.Records.Edit(ctx, record.ID, dns.RecordEditParams{
					ZoneID: cloudflare.F(id),
					Record: dns.RecordParam{
						Content: cloudflare.F(ip.String()),
					},
				})
			} else {
				log.Infof("Skip update of DNS record. (dry-run mode active)")
			}
		}
	}
	return
}
