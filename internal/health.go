package internal

import (
	"context"
	"encoding/json"
	"net/http"
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

// Following variables will be statically linked at the time of compiling
// to link, update ldflags during build like this:
//
// go build -ldflags "-X main.BuildDate=$(date '+%Y-%m-%dT%H:%M:%S') -X main.GitCommit=$(git rev-parse HEAD) -X main.GitBranch=$(git rev-parse --abbrev-ref HEAD)"

// GitCommit holds commit hash of source tree
var GitCommit string

// GitBranch holds current branch name the code is built off
var GitBranch string

// BuildDate holds RFC3339 formatted UTC date (build time)
var BuildDate string

type service struct {
	Name    string
	Version string `json:",omitempty"`
	Status  string
	Cause   string `json:",omitempty"`
}

type buildInfo struct {
	Commit    string
	Branch    string
	BuildDate string
}

type healthResponse struct {
	Status    string
	Services  []service
	BuildInfo buildInfo
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(s.healthCheck())
}

func (s *Server) healthCheck() []byte {

	response := healthResponse{
		Status: statusUp,
		BuildInfo: buildInfo{
			Commit:    GitCommit,
			Branch:    GitBranch,
			BuildDate: BuildDate,
		},
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
