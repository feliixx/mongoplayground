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
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/feliixx/mgodatagen/datagen"
	"github.com/feliixx/mgodatagen/datagen/generators"
	"github.com/feliixx/mongoextjson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// max number of collection to create at once
	maxCollNb = 10
	// max number of documents in a collection
	maxDoc = 100
	// max time a query can run before being aborted by the server
	maxQueryTime = 1 * time.Minute
	// errInvalidConfig error message when the configuration doesn't match expected format
	errInvalidConfig = `expecting an array of documents like 

[ 
  {_id: 1, k: "one"},
  {_id: 2, k: "two"}
]

or a list of collections like:

db = { 
	collection1: [ 
		{_id: 1, k: "one"},
		{_id: 2, k: "two"}
	],
	collection2: [
		{_id: 1, v: 1}
	]
}`
	errInvalidQuery    = "query must match db.coll.find(...) or db.coll.aggregate(...) or db.coll.update()"
	errPlaygroundToBig = "playground is too big"
	noDocFound         = "no document found"

	findMethod      = "find"
	aggregateMethod = "aggregate"
	updateMethod    = "update"
)

// run a query and return the results as plain text.
// the result is compacted and looks like:
//
//    [{_id:1,k:1},{_id:2,k:33}]
func (s *server) runHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	p, err := newPage(
		r.FormValue("mode"),
		r.FormValue("config"),
		r.FormValue("query"),
	)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	res, err := s.run(p)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(res)
}

func (s *server) run(p *page) ([]byte, error) {

	collectionName, method, stages, err := parseQuery(p.Query)
	if err != nil {
		return nil, fmt.Errorf("error in query:\n  %v", err)
	}

	db := s.session.Database(p.dbHash())

	// if this is an 'update' query, always re-create the database,
	// run the update and return the result of a 'find' query on the
	// same collection
	dbInfos, err := s.createDatabase(db, p.Mode, p.Config, method == updateMethod)
	if err != nil {
		return nil, fmt.Errorf("error in configuration:\n  %v", err)
	}

	// mongodb returns an empy array ( [] ) if we try to run a query on a collection
	// that doesn't exist. Check that the collection exist before running the query,
	// to return a clear error message in that case
	if !dbInfos.hasCollection(collectionName) {
		return nil, fmt.Errorf(`collection "%s" doesn't exist`, collectionName)
	}
	return runQuery(db.Collection(collectionName), method, stages)
}

func (s *server) createDatabase(db *mongo.Database, mode byte, config []byte, forceCreate bool) (dbInfo dbMetaInfo, err error) {

	s.mutex.Lock()
	defer s.mutex.Unlock()

	dbInfo, exists := s.activeDB[db.Name()]
	if !exists || forceCreate {

		collections := map[string][]bson.M{}

		switch mode {
		case mgodatagenMode:
			err = createContentFromMgodatagen(collections, config)
		case bsonMode:
			err = loadContentFromJSON(collections, config)
		}
		if err != nil {
			return dbInfo, err
		}

		dbInfo, err = fillDatabase(db, collections)
		if err != nil {
			return dbInfo, err
		}

		// if the database is empty, ie all collections contains no document,
		// we do not add the database to the activeDB map and just return
		// directly
		//
		// if the database is empty, but it's an update (ie 'forceCreate' is true),
		// it might be an upsert wich would create a database, so in doubt add the
		// database to the activeDB map
		//
		// if the database was already present ( for exemple, if an user run the
		// exact same update query twice ), but is re-created because 'forceCreate'
		// is true, it's already in the activeDB map, we return directly to avoid
		// incrementing the 'activeDatabase' counter. 'lastUsed' access is not updated,
		// but it doesn't matter because db is re-created every time
		if (dbInfo.emptyDatabase && !forceCreate) || (exists && forceCreate) {
			return dbInfo, nil
		}
		activeDatabases.Inc()
	}

	dbInfo.lastUsed = time.Now().Unix()
	s.activeDB[db.Name()] = dbInfo

	return dbInfo, nil
}

func createContentFromMgodatagen(collections map[string][]bson.M, config []byte) error {

	collConfigs, err := datagen.ParseConfig(config, true)
	if err != nil {
		return err
	}

	mapRef := map[int][][]byte{}
	mapRefType := map[int]bsontype.Type{}

	for _, c := range collConfigs {

		ci := generators.NewCollInfo(c.Count, []int{3, 6}, 1, mapRef, mapRefType)
		if ci.Count > maxDoc || ci.Count <= 0 {
			ci.Count = maxDoc
		}
		g, err := ci.NewDocumentGenerator(c.Content)
		if err != nil {
			return fmt.Errorf("fail to create collection %s: %v", c.Name, err)
		}
		docs := make([]bson.M, ci.Count)
		for i := 0; i < ci.Count; i++ {
			err := bson.Unmarshal(g.Generate(), &docs[i])
			if err != nil {
				return err
			}
		}
		collections[c.Name] = docs
	}
	return nil
}

func loadContentFromJSON(collections map[string][]bson.M, config []byte) error {

	switch detailBsonMode(config) {
	case bsonSingleCollection:
		var docs []bson.M
		err := mongoextjson.Unmarshal(config, &docs)

		collections["collection"] = docs
		return err

	case bsonMultipleCollection:
		return mongoextjson.Unmarshal(config[3:], &collections)

	default:
		return errors.New(errInvalidConfig)
	}
}

func fillDatabase(db *mongo.Database, collections map[string][]bson.M) (dbInfo dbMetaInfo, err error) {

	if len(collections) > maxCollNb {
		return dbInfo, fmt.Errorf("max number of collection in a database is %d, but was %d", maxCollNb, len(collections))
	}
	// clean any potentially remaining data
	db.Drop(context.Background())

	dbInfo = dbMetaInfo{
		collections:   make(sort.StringSlice, 0, len(collections)),
		emptyDatabase: true,
	}
	// order the collections by name, so the order of creation is
	// garenteed to be always the same
	for name := range collections {
		dbInfo.collections = append(dbInfo.collections, name)
	}
	dbInfo.collections.Sort()

	base := 0
	for _, name := range dbInfo.collections {

		docs := collections[name]
		if len(docs) == 0 {
			continue
		}
		dbInfo.emptyDatabase = false

		if len(docs) > maxDoc {
			docs = docs[:maxDoc]
		}
		// if no _id is specified, we insert fake objectID that are
		// garenteed to be the same from one run to another, so the
		// output of a specific config is garenteed to always be the
		// same, at least in bson mode
		var toInsert = make([]interface{}, len(docs))
		for i, doc := range docs {
			if _, hasID := doc["_id"]; !hasID {
				doc["_id"] = seededObjectID(int32(base + i))
			}
			toInsert[i] = doc
		}

		opts := options.InsertMany().SetOrdered(true)
		_, err := db.Collection(name).InsertMany(context.Background(), toInsert, opts)
		if err != nil {
			// In some case, a collection can be partially created even if some write failed
			//
			// for example: [{_id:1},{_id:1}]
			//
			// -> the first write will succeed, but the second will fail, so a collection
			// containing only one record will be created, and an error will be returned
			//
			// Because fillDatabase returns an error, the hash of the database (ie db.name)
			// is not put in server.activeDB, so it can't be deleted from server.removeExpiredDB
			//
			// to avoid this kind of leaks, drop the db immediately if there is an error
			db.Drop(context.Background())
			return dbInfo, err
		}
		base += len(docs)
	}
	return dbInfo, nil
}

func seededObjectID(n int32) primitive.ObjectID {

	// using date = uint32(time.Date(2018, 02, 26, 0, 0, 0, 0, time.UTC).Unix())

	return [12]byte{
		byte(90),  // date << 24
		byte(147), // date << 16
		byte(78),  // date << 8
		byte(0),   // date
		byte(1),   // 1,2,3 for hostname bytes
		byte(2),
		byte(3),
		byte(4), // 4,5 for pid bytes
		byte(5),
		byte(n >> 16), // Increment, 3 bytes, big endian
		byte(n >> 8),
		byte(n),
	}
}

// query has to match the following regex:
//
//   /^db\..(\w*)\.(find|aggregate|update)\([\s\S]*\)$/
//
// for example:
//
//   db.collection.find({k:1})
//   db.collection.aggregate([{$project:{_id:0}}])
//   db.collection.update({k:1},{$set:{n:1}},{upsert:true})
//
// input is filtered from front-end side, but this should
// not panic on pathological/malformatted input
func parseQuery(query []byte) (collectionName, method string, stages []bson.M, err error) {

	p := bytes.SplitN(query, []byte{'.'}, 3)
	if len(p) != 3 {
		return "", "", nil, errors.New(errInvalidQuery)
	}

	collectionName = string(p[1])

	// last part of query contains the method and the stages, for example find({k:1})
	queryBytes := p[2]
	start, end := bytes.IndexByte(queryBytes, '('), bytes.LastIndexByte(queryBytes, ')')

	if start == -1 || end == -1 {
		return "", "", nil, errors.New(errInvalidQuery)
	}

	method = string(queryBytes[:start])

	stages, err = unmarshalStages(queryBytes[start+1 : end])
	if err != nil {
		return "", "", nil, fmt.Errorf("fail to parse content of query: %v", err)
	}

	return collectionName, method, stages, nil
}

func unmarshalStages(queryBytes []byte) (stages []bson.M, err error) {

	if len(queryBytes) == 0 {
		return make([]bson.M, 2), nil
	}

	// because projections are allowed, transform
	// {}, {"_id": 0} into [{}, {"_id": 0}] so we
	// can parse it as a []bson.M
	if queryBytes[0] != '[' {
		b := make([]byte, 0, len(queryBytes)+2)
		b = append(b, '[')
		b = append(b, queryBytes...)
		b = append(b, ']')
		queryBytes = b
	}

	err = mongoextjson.Unmarshal(queryBytes, &stages)

	return stages, err
}

func runQuery(collection *mongo.Collection, method string, stages []bson.M) ([]byte, error) {

	var docs []bson.M
	var err error
	var cursor *mongo.Cursor

	switch method {
	case aggregateMethod:
		cursor, err = collection.Aggregate(context.Background(), stages, options.Aggregate().SetMaxTime(maxQueryTime))
	case findMethod:
		for len(stages) < 2 {
			stages = append(stages, bson.M{})
		}
		cursor, err = collection.Find(context.Background(), stages[0], options.Find().SetProjection(stages[1]).SetMaxTime(maxQueryTime))
	case updateMethod:
		for len(stages) < 3 {
			stages = append(stages, bson.M{})
		}

		multi, opts := parseUpdateOpts(stages[2])
		if multi {
			_, err = collection.UpdateMany(context.Background(), stages[0], stages[1], opts)
		} else {
			_, err = collection.UpdateOne(context.Background(), stages[0], stages[1], opts)
		}

		if err != nil {
			return nil, fmt.Errorf("fail to run update: %v", err)
		}
		cursor, err = collection.Find(context.Background(), bson.M{})

	default:
		err = fmt.Errorf("invalid method: %s", method)
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}

	if err = cursor.All(context.Background(), &docs); err != nil {
		return nil, fmt.Errorf("fail to get result from cursor: %v", err)
	}

	if len(docs) == 0 {
		return []byte(noDocFound), nil
	}
	return mongoextjson.Marshal(docs)
}

func parseUpdateOpts(optsDoc bson.M) (bool, *options.UpdateOptions) {

	multi, _ := optsDoc["multi"].(bool)

	upsert, _ := optsDoc["upsert"].(bool)
	arrayFilters, _ := optsDoc["arrayFilters"].([]interface{})

	return multi, options.Update().
		SetUpsert(upsert).
		SetArrayFilters(options.ArrayFilters{
			Filters: arrayFilters,
		})

}
