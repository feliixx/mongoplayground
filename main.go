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
	"time"

	"github.com/feliixx/mongoplayground/internal"

	"github.com/spf13/viper"
)

const (
	badgerDir = "storage"
	backupDir = "backups"
)

func main() {

	loadConfig()
	setLogger()

	s, err := internal.NewServer(
		viper.GetString("mongo.uri"),
		viper.GetBool("mongo.dropFirst"),
		badgerDir,
		backupDir,
		loadSmtp(),
	)
	if err != nil {
		log.Fatalf("aborting: %v\n", err)
	}

	if !viper.GetBool("https.enabled") {
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
		viper.GetString("https.fullchain"),
		viper.GetString("https.privkey"),
	))
}

func loadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.SetDefault("https.enabled", false)
	viper.SetDefault("mongo.uri", "mongodb://localhost:27017")
	viper.SetDefault("mongo.dropFirst", false)
	viper.SetDefault("logging.loki.host", "")
	viper.SetDefault("mail.enabled", false)
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("error while loading conf: %v", err)
	}
}

func setLogger() {

	if viper.GetString("logging.loki.host") != "" {

		logger := internal.NewLokiLogger(
			viper.GetString("logging.loki.host"),
			viper.GetInt("logging.loki.port"),
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
}

func loadSmtp() *internal.MailInfo {

	if viper.GetBool("mail.enabled") {

		return internal.NewMailInfo(
			viper.GetString("mail.smtp.host"),
			viper.GetInt("mail.smtp.port"),
			viper.GetString("mail.smtp.from"),
			viper.GetString("mail.smtp.pwd"),
			viper.GetString("mail.sendTo"),
		)
	} else {
		return nil
	}
}

func redirectTLS(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
}
