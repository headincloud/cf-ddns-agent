package discovery

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
)

func DiscoverIPv4(DiscoveryURL string) (ip net.IP, err error){
	// get ip
	resp, err := http.Get(DiscoveryURL)
	if err != nil {
		log.Errorf("Could not connect to IP discovery service: %s", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Could not read response from IP discovery service: %s", err.Error())
	}
	ip = net.ParseIP(string(body))
	if ip == nil {
		log.Errorf("Could not parse received value as an IP address: %s", string(body))
	}
	log.Infof("IP address received: %s", ip)
	return
}
