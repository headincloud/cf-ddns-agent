package util

import (
	"github.com/cloudflare/cloudflare-go"
	log "github.com/sirupsen/logrus"

	"net"
)

func PerformRecordUpdate(token string, domain string, host string, value net.IP) (err error) {

	// start with creating a CF api object with the token.
	api, err := cloudflare.NewWithAPIToken(token)
	if err != nil {
		log.Errorf("Error encountered while creating CloudFlare API object: %s", err.Error())
		return
	}

	// check current setting
	id, err := api.ZoneIDByName(domain)
	if err!= nil {
		log.Errorf("Error encountered while performing lookup of zone %s: %s", domain, err.Error())
		return
	}

	foo := cloudflare.DNSRecord{
		Name: host,
		Type: "A",
	}
	records, err := api.DNSRecords(id, foo)
	if err != nil {
		log.Errorf("Error encountered while checking current value of %s: %s", host, err.Error())
	}
	for _, record := range records {
		CurrentIP := net.ParseIP(record.Content)
		if CurrentIP.Equal(value) {
			log.Infof("IP address up to date for record %s (type %s). No DNS change necessary.", record.Name, record.Type)
		} else {
			log.Infof("Updating IP address of record %s (type %s) to %s", record.Name, record.Type, value)
			record.Content = value.String()
			err = api.UpdateDNSRecord(id, record.ID, record)
			if err != nil {
				log.Errorf("Error updating DNS record for %s (type %s) to %s: %s", record.Name, record.Type, value, err.Error())
			} else {
				log.Infof("IP address of record record %s (type %s) successfully updated.", record.Name, record.Type)
			}
		}
	}

	return
}
