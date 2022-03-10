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
	"go.mongodb.org/mongo-driver/bson/mgocompat"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// max number of collection to create at once
	maxCollNb = 10
	// max number of documents in a collection
	maxDoc = 100
	// max time a query can run before being aborted by the Server
	maxQueryTime = writeTimeout - readTimeout
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
func (s *storage) runHandler(w http.ResponseWriter, r *http.Request) {

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

	res, err := s.run(r.Context(), p)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(res)
}

func (s *storage) run(context context.Context, p *page) ([]byte, error) {

	collectionName, method, stages, explainMode, err := parseQuery(p.Query)
	if err != nil {
		return nil, fmt.Errorf("error in query:\n  %v", err)
	}

	db := s.mongoSession.Database(p.dbHash())

	// if this is an 'update' query, always re-create the database,
	// run the update and return the result of a 'find' query on the
	// same collection
	forceCreate := method == updateMethod
	dbInfos, err := s.createDatabase(db, p.Mode, p.Config, forceCreate)
	if err != nil {
		return nil, fmt.Errorf("error in configuration:\n  %v", err)
	}

	// mongodb returns an empty array ( [] ) if we try to run a query on a collection
	// that doesn't exist. Check that the collection exist before running the query,
	// to return a clear error message in that case
	if !dbInfos.hasCollection(collectionName) {
		return nil, fmt.Errorf(`collection "%s" doesn't exist`, collectionName)
	}
	return runQuery(context, db.Collection(collectionName), method, stages, explainMode)
}

func (s *storage) createDatabase(db *mongo.Database, mode byte, config []byte, forceCreate bool) (dbInfo dbMetaInfo, err error) {

	s.activeDbLock.Lock()
	defer s.activeDbLock.Unlock()

	dbInfo, exists := s.activeDB[db.Name()]
	if !exists || forceCreate {

		switch mode {
		case mgodatagenMode:
			dbInfo, err = createDBFromMgodatagen(db, config)
		case bsonMode:
			dbInfo, err = createDBFromJSON(db, config)
		}
		if err != nil {
			return dbInfo, err
		}

		// if the database is empty, ie all collections contains no document,
		// we do not add the database to the activeDB map and just return
		// directly
		//
		// if the database is empty, but it's an update (ie 'forceCreate' is true),
		// it might be an upsert which would create a database, so in doubt add the
		// database to the activeDB map
		//
		// if the database was already present ( for example, if an user run the
		// exact same update query twice ), but is re-created because 'forceCreate'
		// is true, it's already in the activeDB map, we return directly to avoid
		// incrementing the 'activeDatabase' counter. 'lastUsed' access is not updated,
		// but it doesn't matter because db is re-created every time
		if (dbInfo.emptyDatabase && !forceCreate) || (exists && forceCreate) {
			return dbInfo, nil
		}
		activeDatabasesCounter.Inc()
	} else {
		// database is already created, just increase the cache hit counter
		dbCacheHit.Inc()
	}

	dbInfo.lastUsed = time.Now().Unix()
	s.activeDB[db.Name()] = dbInfo

	return dbInfo, nil
}

func createDBFromMgodatagen(db *mongo.Database, config []byte) (dbInfo dbMetaInfo, err error) {

	collConfigs, err := datagen.ParseConfig(config, true)
	if err != nil {
		return dbInfo, err
	}

	collections := map[string][]bson.M{}
	indexes := map[string][]datagen.Index{}

	mapRef := map[int][][]byte{}
	mapRefType := map[int]bsontype.Type{}

	for _, c := range collConfigs {

		ci := generators.NewCollInfo(c.Count, []int{3, 6}, 1, mapRef, mapRefType)
		if ci.Count > maxDoc || ci.Count <= 0 {
			ci.Count = maxDoc
		}
		g, err := ci.NewDocumentGenerator(c.Content)
		if err != nil {
			return dbInfo, fmt.Errorf("fail to create collection %s: %v", c.Name, err)
		}
		docs := make([]bson.M, ci.Count)
		for i := 0; i < ci.Count; i++ {

			// make a copy of the slice generated by mgodatagen to avoid
			// weird reference bug when unmarshaling like https://github.com/feliixx/mongoplayground/issues/120
			b := g.Generate()
			c := make([]byte, len(b))
			copy(c, b)

			err := bson.Unmarshal(c, &docs[i])
			if err != nil {
				return dbInfo, err
			}
		}
		collections[c.Name] = docs
		if len(c.Indexes) > 0 {
			indexes[c.Name] = c.Indexes
		}
	}
	// clean any potentially remaining data
	err = db.Drop(context.Background())
	if err != nil {
		return dbInfo, err
	}
	err = createIndexes(db, indexes)
	if err != nil {
		return dbInfo, err
	}
	return fillDatabase(db, collections)
}

func createIndexes(db *mongo.Database, indexes map[string][]datagen.Index) error {
	// see dtg.runMgoCompatCommand @ github.com/feliixx/mgodatagen/datagen
	mgoRegistry := mgocompat.NewRespectNilValuesRegistryBuilder().Build()
	for collName, idx := range indexes {
		cmd := bson.D{
			bson.E{Key: "createIndexes", Value: collName},
			bson.E{Key: "indexes", Value: idx},
		}
		_, cmdBytes, err := bson.MarshalValueWithRegistry(mgoRegistry, cmd)
		if err != nil {
			return fmt.Errorf("failed to generate mgocompat command\n  cause: %v", err)
		}
		res := db.RunCommand(context.Background(), cmdBytes)
		if res != nil && res.Err() != nil {
			return res.Err()
		}
	}
	return nil
}

func createDBFromJSON(db *mongo.Database, config []byte) (dbInfo dbMetaInfo, err error) {

	collections := map[string][]bson.M{}

	switch detailBsonMode(config) {
	case bsonSingleCollection:
		var docs []bson.M
		err = mongoextjson.Unmarshal(config, &docs)

		collections["collection"] = docs

	case bsonMultipleCollection:
		err = mongoextjson.Unmarshal(config[3:], &collections)

	default:
		err = errors.New(errInvalidConfig)
	}

	if err != nil {
		return dbInfo, err
	}

	// clean any potentially remaining data
	err = db.Drop(context.Background())
	if err != nil {
		return dbInfo, err
	}
	return fillDatabase(db, collections)
}

func fillDatabase(db *mongo.Database, collections map[string][]bson.M) (dbInfo dbMetaInfo, err error) {

	if len(collections) > maxCollNb {
		return dbInfo, fmt.Errorf("max number of collection in a database is %d, but was %d", maxCollNb, len(collections))
	}

	dbInfo = dbMetaInfo{
		collections:   make(sort.StringSlice, 0, len(collections)),
		emptyDatabase: true,
	}
	// order the collections by name, so the order of creation is
	// guaranteed to be always the same
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
		// guaranteed to be the same from one run to another, so the
		// output of a specific config is guaranteed to always be the
		// same, at least in bson mode
		var toInsert = make([]interface{}, len(docs))
		for i, doc := range docs {

			// in production logs, it appears that some docs can be
			// nil at this point, triggering a panic.
			// couldn't find how it's possible yet, so just add a nil
			// check in the meantime
			if _, hasID := doc["_id"]; !hasID && doc != nil {
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

// find, aggregate and update queries are supported, with or without explain()
// once the .explain() part is stripped, the query has to match the following
// regex:
//           /^db\..(\w*)\.(find|aggregate|update)\([\s\S]*\)$/
//
// for example, those queries are valid:
//
//   db.collection.find({k:1})
//   db.collection.aggregate([{$project:{_id:0}}])
//   db.collection.update({k:1},{$set:{n:1}},{upsert:true})
//   db.collection.find({k:1}).explain()
//   db.collection.explain("executionStats").find({k:1})
//
// input is filtered from front-end side, but this should
// not panic on pathological/malformatted input
func parseQuery(query []byte) (collectionName, method string, stages []interface{}, explainMode string, err error) {

	query, explainMode = stripExplain(query)

	p := bytes.SplitN(query, []byte{'.'}, 3)
	if len(p) != 3 {
		return "", "", nil, "", errors.New(errInvalidQuery)
	}

	collectionName = string(p[1])

	// last part of query contains the method and the stages, for example find({k:1})
	queryBytes := p[2]
	start, end := bytes.IndexByte(queryBytes, '('), bytes.LastIndexByte(queryBytes, ')')

	if start == -1 || end == -1 {
		return "", "", nil, "", errors.New(errInvalidQuery)
	}

	method = string(queryBytes[:start])

	stages, err = unmarshalStages(queryBytes[start+1 : end])
	if err != nil {
		return "", "", nil, "", fmt.Errorf("fail to parse content of query: %v", err)
	}

	return collectionName, method, stages, explainMode, nil
}

func stripExplain(query []byte) (strippedQuery []byte, explainMode string) {

	startExplain := bytes.Index(query, []byte(".explain("))
	if startExplain == -1 {
		return query, ""
	}

	endExplain := bytes.Index(query[startExplain:], []byte(")"))
	if endExplain == -1 {
		return query, ""
	}

	endExplain += startExplain
	if endExplain+1 == len(query) {
		query = query[:startExplain]
	} else {
		query = append(query[:startExplain], query[endExplain+1:]...)
	}

	explainMode = string(query[startExplain+9 : endExplain])
	if len(explainMode) < 2 {
		explainMode = "queryPlanner"
	} else {
		// remove the enclosing double quote (")
		explainMode = explainMode[1 : len(explainMode)-1]
	}

	return query, explainMode
}

// most of the time, each stage is a bson.M document.
//
// however, since mongodb 4.2, the second stage of an update()
// can be an slice of bson.M
//
//   cf https://docs.mongodb.com/manual/tutorial/update-documents-with-aggregation-pipeline/
//
func unmarshalStages(queryBytes []byte) (stages []interface{}, err error) {

	if len(queryBytes) == 0 {
		return []interface{}{bson.M{}, bson.M{}}, nil
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

func runQuery(context context.Context, collection *mongo.Collection, method string, stages []interface{}, explainMode string) ([]byte, error) {

	var cmd bson.D

	switch method {
	case aggregateMethod:

		cmd = bson.D{
			{Key: aggregateMethod, Value: collection.Name()},
			{Key: "pipeline", Value: sanitize(stages)},
			{Key: "cursor", Value: bson.M{"batchSize": 1000}},
		}

	case findMethod:

		for len(stages) < 2 {
			stages = append(stages, bson.M{})
		}

		cmd = bson.D{
			{Key: findMethod, Value: collection.Name()},
			{Key: "filter", Value: stages[0]},
			{Key: "projection", Value: stages[1]},
		}

	case updateMethod:

		for len(stages) < 3 {
			stages = append(stages, bson.M{})
		}

		var err error
		multi, opts := parseUpdateOpts(stages[2])
		if multi {
			_, err = collection.UpdateMany(context, stages[0], stages[1], opts)
		} else {
			_, err = collection.UpdateOne(context, stages[0], stages[1], opts)
		}
		if err != nil {
			return nil, fmt.Errorf("fail to run update: %v", err)
		}

		cmd = bson.D{
			{Key: findMethod, Value: collection.Name()},
			{Key: "filter", Value: bson.M{}},
		}

	default:
		return nil, fmt.Errorf("invalid method: '%s'", method)
	}

	// make sure that all types of queries have a timeout,
	// even in explain mode
	cmd = append(cmd,
		bson.E{Key: "maxTimeMS", Value: maxQueryTime.Milliseconds()},
	)

	if explainMode != "" {
		cmd = bson.D{
			{Key: "explain", Value: cmd},
			{Key: "verbosity", Value: explainMode},
		}
	}

	res := collection.Database().RunCommand(context, cmd)
	if res.Err() != nil {
		return nil, fmt.Errorf("query failed: %v", res.Err())
	}

	var cursorDoc bson.M
	if err := res.Decode(&cursorDoc); err != nil {
		return nil, fmt.Errorf("fail to get result from cursor: %v", err)
	}

	if explainMode != "" {
		// not really sensitive, but it's useless as the server version already appears
		// in the footer of the site, so just remove it
		delete(cursorDoc, "serverInfo")
		delete(cursorDoc, "ok")

		return mongoextjson.Marshal(cursorDoc)
	}
	// result doc looks like
	//
	// {"cursor":{"firstBatch":[{"_id":1},{"_id":2}],"id":NumberLong(0),"ns":"dbName.collection"},"ok":1}
	docs := cursorDoc["cursor"].(bson.M)["firstBatch"].(bson.A)
	if len(docs) == 0 {
		return []byte(noDocFound), nil
	}
	return mongoextjson.Marshal(docs)
}

func parseUpdateOpts(opts interface{}) (bool, *options.UpdateOptions) {

	optsDoc, _ := opts.(map[string]interface{})

	multi, _ := optsDoc["multi"].(bool)
	upsert, _ := optsDoc["upsert"].(bool)
	arrayFilters, _ := optsDoc["arrayFilters"].([]interface{})

	return multi, options.Update().
		SetUpsert(upsert).
		SetArrayFilters(options.ArrayFilters{
			Filters: arrayFilters,
		})
}

// remove any stages that might write to another db/collection,
// to avoid leaking databases, or or other playground contamination
func sanitize(stages []interface{}) []interface{} {

	for i := 0; i < len(stages); i++ {
		stage, _ := stages[i].(map[string]interface{})
		if _, ok := stage["$out"]; ok {
			stages = append(stages[:i], stages[i+1:]...)
			i--
		}
		if _, ok := stage["$merge"]; ok {
			stages = append(stages[:i], stages[i+1:]...)
			i--
		}
	}
	return stages
}
