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

package util

import (
	"context"
	"net"

	"github.com/cloudflare/cloudflare-go"
	log "github.com/sirupsen/logrus"
)

func UpdateCFRecord(token string, domain string, host string, recordType string, ip net.IP, dryRun bool, createMode bool) (err error) {
	// start with creating a CF api object with the token.
	api, err := cloudflare.NewWithAPIToken(token)
	if err != nil {
		log.Errorf("Error encountered while creating CloudFlare API object: %s", err.Error())
		return
	}

	// check current setting
	id, err := api.ZoneIDByName(domain)
	if err != nil {
		log.Errorf("Error encountered while performing lookup of zone %s: %s", domain, err.Error())
		return
	}

	cfRecord := cloudflare.DNSRecord{
		Name: host,
		Type: recordType,
	}
	records, err := api.DNSRecords(context.Background(),id, cfRecord)
	if err != nil {
			log.Errorf("Error encountered querying record %s: %s", host, err.Error())
			return
	}
	if len(records)==0 {
		if createMode {
			log.Infof("No record found for %s (type %s). Will attempt to create one...", host, recordType)
			// don't know why cf sdk requires a pointer to a boolean for proxified...
			proxied:=true
			newRecord := cloudflare.DNSRecord{
				Name: host,
				Type: recordType,
				Content: ip.String(),
				Proxied: &proxied,
			}
			if !dryRun {
				_, err = api.CreateDNSRecord(context.Background(), id, newRecord)
				if err != nil {
					log.Errorf("Error encountered while creating record for %s: %s", host, err.Error())
					return
				}
				log.Infof("Record created for %s: %s", host, recordType)
			} else {
				log.Infof("Skipped creation of DNS record. (dry-run mode active)")
			}
			return
		} else {
			log.Errorf("No record found for %s: %s", host, recordType)
			return
		}
	}

	for _, record := range records {
		CurrentIP := net.ParseIP(record.Content)
		if CurrentIP.Equal(ip) {
			log.Infof("IP address up to date for record %s (type %s). No DNS change necessary.", record.Name, record.Type)
		} else {
			log.Infof("Updating IP address of record %s (type %s) to %s", record.Name, record.Type, ip)
			record.Content = ip.String()
			if !dryRun {
				err = api.UpdateDNSRecord(context.Background(), id, record.ID, record)
				if err != nil {
					log.Errorf("Error updating DNS record for %s (type %s) to %s: %s", record.Name, record.Type, ip, err.Error())
				} else {
					log.Infof("IP address of record record %s (type %s) successfully updated.", record.Name, record.Type)
				}
			} else {
				log.Infof("Skip update of DNS record. (dry-run mode active)")
			}
		}
	}
	return
}
