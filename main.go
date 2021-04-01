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
	"github.com/feliixx/mongoplayground/internal"
	"log"
)

const (
	badgerDir = "storage"
	backupDir = "backups"
)

func main() {
	s, err := internal.NewServer(badgerDir, backupDir)
	if err != nil {
		log.Fatalf("aborting: %v\n", err)
	}
	log.Fatal(s.ListenAndServe())
}
