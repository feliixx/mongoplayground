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

	"github.com/feliixx/mongoplayground/internal"
)

const (
	badgerDir = "storage"
	backupDir = "backups"
)

func redirectTLS(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
}

func main() {
	l := log.New(os.Stdout, "", log.LstdFlags)
	s, err := internal.NewServer(l, badgerDir, backupDir)
	if err != nil {
		l.Fatalf("aborting: %v\n", err)
	}
	go func() {
		if err := http.ListenAndServe(":80", http.HandlerFunc(redirectTLS)); err != nil {
			l.Fatalf("ListenAndServe error: %v", err)
		}
	}()
	l.Fatal(http.ListenAndServeTLS(
		":443",
		"/etc/letsencrypt/live/www.mongoplayground.net/fullchain.pem",
		"/etc/letsencrypt/live/www.mongoplayground.net/privkey.pem",
		s,
	),
	)
}
