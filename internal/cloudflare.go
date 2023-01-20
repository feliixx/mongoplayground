// mongoplayground: a sandbox to test and share MongoDB queries
// Copyright (C) 2022 Adrien Petel
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package internal

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type CloudflareInfo struct {
	zoneID   string
	apiToken string
}

func NewCloudflareInfo(zoneID, apiToken string) *CloudflareInfo {
	return &CloudflareInfo{
		zoneID:   zoneID,
		apiToken: apiToken,
	}
}

func (c *CloudflareInfo) clearCacheHandler(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("Authorization")
	if token != fmt.Sprintf("Bearer %s", c.apiToken) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	log.Println("clearing cloudflare cache")
	resp := c.clearCloudflareCache()
	log.Printf("result: %s", resp)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(resp)
}

func (c *CloudflareInfo) clearCloudflareCache() []byte {

	if c.zoneID == "" || c.apiToken == "" {
		return []byte("cloudflare auth not configured")
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).Dial,
			DisableKeepAlives: true,
		},
	}

	req, _ := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/purge_cache", c.zoneID),
		strings.NewReader(`{
		  "files": [
			"https://mongoplayground.net",
			"https://mongoplayground.net/",
			"https://mongoplayground.net/p/*",
		    "https://mongoplayground.net/static/*.html"
		  ]
		}`),
	)
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("fail to send request to clear cloudflare cache: %v", err)
		return []byte("fail to send request to cloudflare")
	}

	buf := bytes.NewBuffer(make([]byte, 512))
	io.Copy(buf, resp.Body)
	resp.Body.Close()

	return buf.Bytes()
}
