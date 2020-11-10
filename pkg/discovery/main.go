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

package discovery

import (
	"io/ioutil"
	"net"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func DiscoverIPv4(DiscoveryURL string) (ip net.IP, err error) {
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
