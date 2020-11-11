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
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func DiscoverIPv4(DiscoveryURL string) (ip net.IP, err error) {
	// get ip
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(DiscoveryURL)
	if err != nil {
		log.Errorf("Could not connect to IP discovery service: %s", err.Error())
	}
	defer resp.Body.Close()
	if ! (resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		log.Errorf("Discovery returned a status code which is not in the 2XX range: %d", resp.StatusCode)
		err = errors.New("Returned status code not in 2XX range.")
	}
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
