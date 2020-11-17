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
	"net"

	"github.com/cloudflare/cloudflare-go"
	log "github.com/sirupsen/logrus"
)

func PerformRecordUpdate(token string, domain string, host string, ipv4 net.IP, ipv6 net.IP) (err error) {
	if ipv4 != nil {
		err = UpdateCFRecord(token, domain, host, "A", ipv4)
	}

	if ipv6 != nil {
		err = UpdateCFRecord(token, domain, host, "AAAA", ipv6)
	}
	return
}

func UpdateCFRecord(token string, domain string, host string, recordType string, ip net.IP) (err error) {
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
	records, err := api.DNSRecords(id, cfRecord)
	if err != nil {
		log.Errorf("Error encountered while checking current IP of %s: %s", host, err.Error())
		return
	}
	for _, record := range records {
		CurrentIP := net.ParseIP(record.Content)
		if CurrentIP.Equal(ip) {
			log.Infof("IP address up to date for record %s (type %s). No DNS change necessary.", record.Name, record.Type)
		} else {
			log.Infof("Updating IP address of record %s (type %s) to %s", record.Name, record.Type, ip)
			record.Content = ip.String()
			err = api.UpdateDNSRecord(id, record.ID, record)
			if err != nil {
				log.Errorf("Error updating DNS record for %s (type %s) to %s: %s", record.Name, record.Type, ip, err.Error())
			} else {
				log.Infof("IP address of record record %s (type %s) successfully updated.", record.Name, record.Type)
			}
		}
	}
	return
}
