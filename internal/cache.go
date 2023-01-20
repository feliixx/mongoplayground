// mongoplayground: a sandbox to test and share MongoDB queries
// Copyright (C) 2023 Adrien Petel
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
	"sort"
	"sync"
)

type cache struct {
	sync.Mutex
	list map[string]dbMetaInfo
}

type dbMetaInfo struct {
	// list of collections in the database
	collections sort.StringSlice
	// last usage of this database, stored as Unix time
	lastUsed int64
	// true if database is ready to use: 
	// either is created on the server, or has a config error
	ready bool
	// any error that occured while creating the db
	err error
}

func (d *dbMetaInfo) hasCollection(collectionName string) bool {
	for _, name := range d.collections {
		if name == collectionName {
			return true
		}
	}
	return false
}
