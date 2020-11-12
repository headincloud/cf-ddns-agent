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
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func DiscoverIPv4(DiscoveryURL string) (ip net.IP, err error) {
	// get ip
	resp, err := RetryableGet(DiscoveryURL)
	if err!=nil {
		log.Errorf("An error occured when contacting the IP discovery service (%s).", DiscoveryURL)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Could not read response from IP discovery service: %s", err.Error())
		return
	}
	ip = net.ParseIP(string(body))
	if ip == nil {
		err = fmt.Errorf("Could not parse received value as an IP address.")
		log.Error(err.Error())
		return
	}
	log.Infof("IP address received: %s", ip)
	return
}

func RetryableGet(url string) (resp *http.Response, err error) {
	count:= 0
	delay := 0 * time.Second
	increment := 10* time.Second
	for count < 3 {
		if delay > (0 * time.Second) {
			time.Sleep(delay)
		}
		delay+=increment
		client := http.Client{
			Timeout: 10 * time.Second,
		}
		resp, err = client.Get(url)
		// connection or read time-out
		if err != nil {
			log.Errorf("Could not connect to url: %s. Retry in %s.", err.Error(), delay.String())
			count++
			continue
		}

		if (resp.StatusCode >= 500 && resp.StatusCode <= 599) {
			err = fmt.Errorf("Server returned HTTP error %d. Retry in %s.", resp.StatusCode, delay.String())
			log.Error(err.Error())
			count++
			continue
		} else if (resp.StatusCode >= 400 && resp.StatusCode <= 499){
			// We cannot recover from 4xx errors, so no need to try further.
			err = fmt.Errorf("Unrecoverable error encountered. Please check the url is valid (HTTP error %d). Request aborted.", resp.StatusCode)
			log.Error(err.Error())
			break
		} else {
			// in other cases, we assume we succeeded so we break the loop.
			break
		}
	}
	return
}