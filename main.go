// mongoplayground: a sandbox to test and share MongoDB queries
// Copyright (C) 2017 Adrien Petel
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

package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/feliixx/boa"
	"github.com/feliixx/mongoplayground/internal"
)

func main() {

	loadConfig()
	setLogger()

	s, err := internal.NewServer(
		boa.GetString("mongo.uri"),
		boa.GetBool("mongo.dropFirst"),
		loadCloudflareInfo(),
		loadMailInfo(),
		loadGoogleDriveInfo(),
	)
	if err != nil {
		log.Fatalf("aborting: %v\n", err)
	}

	if !boa.GetBool("https.enabled") {
		log.Fatal(s.ListenAndServe())
		return
	}

	go func() {
		if err := http.ListenAndServe(":80", http.HandlerFunc(redirectTLS)); err != nil {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	s.Addr = ":443"
	log.Fatal(s.ListenAndServeTLS(
		boa.GetString("https.fullchain"),
		boa.GetString("https.privkey"),
	))
}

func loadConfig() {

	f, err := os.Open("config.json")
	if err != nil {
		log.Println("config file not found")
		return
	}
	err = boa.ParseConfig(f)
	if err != nil {
		log.Printf("error while loading conf: %v", err)
	}
}

func setLogger() {

	if !boa.GetBool("loki.enabled") {
		return
	}

	logger := internal.NewLokiLogger(
		boa.GetString("loki.host"),
		boa.GetInt("loki.port"),
	)
	log.SetOutput(logger)

	go func(l *internal.LokiLogger) {
		for range time.Tick(5 * time.Minute) {
			err := l.Send()
			if err != nil {
				log.Printf("fail to send to loki: %v", err)
			}
		}
	}(logger)
}

func loadMailInfo() *internal.MailInfo {

	if !boa.GetBool("mail.enabled") {
		return nil
	}

	return internal.NewMailInfo(
		boa.GetString("mail.smtp.host"),
		boa.GetInt("mail.smtp.port"),
		boa.GetString("mail.smtp.from"),
		boa.GetString("mail.smtp.pwd"),
		boa.GetString("mail.sendTo"),
	)
}

func loadCloudflareInfo() *internal.CloudflareInfo {
	return internal.NewCloudflareInfo(
		boa.GetString("cloudflare.zone_id"),
		boa.GetString("cloudflare.api_token"),
	)
}

func loadGoogleDriveInfo() *internal.GoogleDriveInfo {

	if !boa.GetBool("google_drive.enabled") {
		return nil
	}

	return internal.NewGoogleDriveInfo(
		boa.GetString("google_drive.dir"),
		boa.GetMap("google_drive.token"),
		boa.GetMap("google_drive.credentials"),
	)
}

func redirectTLS(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
}
