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

package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime/debug"
	"strconv"
)

const (
	// everything is ok
	statusUp = "UP"
	// application is up, but at least one service it depends on
	// is down
	statusDegrade = "DEGRADE"
	// service is unavailable
	statusDown = "DOWN"
)

type serviceInfo struct {
	Name    string
	Version string `json:",omitempty"`
	Status  string
	Cause   string `json:",omitempty"`
}

type healthResponse struct {
	Status   string
	Services []serviceInfo
	Version  string
}

func (s *storage) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-transform")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(s.healthCheck())
}

func (s *storage) healthCheck() []byte {

	response := healthResponse{
		Status: statusUp,
	}

	badger := serviceInfo{
		Name:   "badger",
		Status: "UP",
	}

	if s.kvStore.IsClosed() {
		badger.Status = statusDown
		badger.Cause = "database is closed"
		response.Status = statusDegrade
	}

	mongodb := serviceInfo{
		Name:    "mongodb",
		Version: string(s.mongoVersion),
		Status:  statusUp,
	}

	err := s.mongoSession.Ping(context.Background(), nil)
	if err != nil {
		mongodb.Status = statusDown
		mongodb.Cause = strconv.Quote(err.Error())
		response.Status = statusDegrade
	}

	if s.backupServiceStatus.Status != statusUp {
		response.Status = statusDegrade
	}

	response.Services = []serviceInfo{
		badger,
		mongodb,
		s.backupServiceStatus,
	}

	if moduleInfo, ok := debug.ReadBuildInfo(); ok {
		for _, s := range moduleInfo.Settings {
			if s.Key == "vcs.revision" {
				response.Version = "production@" + s.Value
			}
		}
	}

	b, _ := json.Marshal(response)
	return b
}
