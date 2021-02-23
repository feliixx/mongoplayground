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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// everything is ok
	statusUp = "UP"
	// application is up, but at least one service it depends on
	// is down
	statusDegrade = "DEGRADE"
	// service is unavalaible
	statusDown = "DOWN"
)

type service struct {
	Name    string
	Version string `json:",omitempty"`
	Status  string
	Cause   string `json:",omitempty"`
}

type healthResponse struct {
	Status   string
	Services []service
	Version  string
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(s.healthCheck())
}

func (s *Server) healthCheck() []byte {

	response := healthResponse{
		Status: statusUp,
	}

	badger := service{
		Name:   "badger",
		Status: "UP",
	}

	if s.storage.IsClosed() {
		badger.Status = statusDown
		badger.Cause = "database is closed"
		response.Status = statusDegrade
	}

	mongodb := service{
		Name:    "mongodb",
		Version: string(s.mongodbVersion),
		Status:  statusUp,
	}

	err := s.session.Ping(context.Background(), nil)
	if err != nil {
		mongodb.Status = statusDown
		mongodb.Cause = strconv.Quote(err.Error())
		response.Status = statusDegrade
	}

	response.Services = []service{
		badger,
		mongodb,
	}

	if moduleInfo, ok := debug.ReadBuildInfo(); ok {
		response.Version = moduleInfo.Main.Version
	}

	b, _ := json.Marshal(response)
	return b
}

func getMongodVersion(client *mongo.Client) []byte {

	result := client.Database("admin").RunCommand(context.Background(), bson.M{"buildInfo": 1})

	var buildInfo struct {
		Version []byte
	}
	err := result.Decode(&buildInfo)
	if err != nil {
		return []byte("unknown")
	}
	return buildInfo.Version
}
