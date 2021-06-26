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

	valid "github.com/asaskevich/govalidator"
	log "github.com/sirupsen/logrus"
)

func DiscoverIPv4(DiscoveryURL string) (ip net.IP, err error) {
	currentDelay := 10 * time.Second
	incrementDelay := 10 * time.Second
	retries := 3
	// get ip
	log.Infof("Contacting the IP discovery service (%s)...", DiscoveryURL)
	resp, retryable, err := RetryableGet(DiscoveryURL)
	if err != nil {
		log.Error(err.Error())
		if retryable {
			for count := 0; count < retries; count++ {
				log.Infof("will retry in %s", currentDelay.String())
				time.Sleep(currentDelay)
				// action
				resp, retryable, err = RetryableGet(DiscoveryURL)
				if err != nil {
					log.Error(err.Error())
					if retryable {
						currentDelay += incrementDelay
						continue
					} else {
						// if not retryable, break loop
						break
					}
				} else {
					// if no error, we can break loop as well
					break
				}
			}
		}
	}
	// if still error, return
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("could not read response from IP discovery service: %s", err.Error())
		return
	}
	if valid.IsIPv4(string(body)) {
		ip = net.ParseIP(string(body))
		if ip == nil {
			err = fmt.Errorf("could not parse received value as an IPv4 address")
			log.Error(err.Error())
			return
		}
		log.Infof("IP address received: %s", ip)
	}
	return
}

func DiscoverIPv6(DiscoveryURL string) (ip net.IP, err error) {
	currentDelay := 10 * time.Second
	incrementDelay := 10 * time.Second
	retries := 3
	// get ip
	log.Infof("Contacting the IPv6 discovery service (%s)...", DiscoveryURL)
	resp, retryable, err := RetryableGet(DiscoveryURL)
	if err != nil {
		log.Error(err.Error())
		if retryable {
			for count := 0; count < retries; count++ {
				log.Infof("will retry in %s", currentDelay.String())
				time.Sleep(currentDelay)
				// action
				resp, retryable, err = RetryableGet(DiscoveryURL)
				if err != nil {
					log.Error(err.Error())
					if retryable {
						currentDelay += incrementDelay
						continue
					} else {
						// if not retryable, break loop
						break
					}
				} else {
					// if no error, we can break loop as well
					break
				}
			}
		}
	}
	// if still error, return
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("could not read response from IP discovery service: %s", err.Error())
		return
	}
	if valid.IsIPv6(string(body)) {
		ip = net.ParseIP(string(body))
		if ip == nil {
			err = fmt.Errorf("could not parse received value as an IPv6 address")
			log.Error(err.Error())
			return
		}
		log.Infof("IP address received: %s", ip)
	}
	return
}

func RetryableGet(url string) (resp *http.Response, retryable bool, err error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err = client.Get(url)
	// connection or read time-out
	if err != nil {
		retryable = true
		return
	}

	if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
		err = fmt.Errorf("server returned HTTP error %d", resp.StatusCode)
		retryable = true
	} else if resp.StatusCode >= 400 && resp.StatusCode <= 499 {
		// We cannot recover from 4xx errors, so no need to try further.
		err = fmt.Errorf("server returned HTTP error %d", resp.StatusCode)
		retryable = false
	}
	return
}
