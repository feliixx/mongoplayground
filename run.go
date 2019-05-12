package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/feliixx/mgodatagen/datagen"
	"github.com/feliixx/mgodatagen/datagen/generators"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	// max number of collection to create at once
	maxCollNb = 10
	// max number of documents in a collection
	maxDoc = 100
	// max size of a collection
	maxBytes = maxDoc * 1024
	// noDocFound error message when no docs match the query
	noDocFound = "no document found"
	// invalidConfig error message when the configuration doesn't match expected format
	invalidConfig = `expecting an array of documents like 

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
	invalidQuery = `query must match db.coll.find(...) or db.coll.aggregate(...)`
)

// run a query and return the results as plain text.
// the result is compacted and looks like:
//
//    [{_id:1,k:1},{_id:2,k:33}]
func (s *server) runHandler(w http.ResponseWriter, r *http.Request) {

	p := newPage(
		r.FormValue("mode"),
		r.FormValue("config"),
		r.FormValue("query"),
	)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	res, err := s.run(p)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(res)
}

func (s *server) run(p *page) ([]byte, error) {

	session := s.session.Copy()
	defer session.Close()

	db := session.DB(p.dbHash())

	dbInfos, err := s.createDatabase(db, p.Mode, p.Config)
	if err != nil {
		return nil, fmt.Errorf("error in configuration:\n  %v", err)
	}
	collectionName, method, stages, err := parseQuery(p.Query)
	if err != nil {
		return nil, fmt.Errorf("error in query:\n  %v", err)
	}
	// mongodb returns an empy array ( [] ) if we try to run a query on a collection
	// that doesn't exist. Check that the collection exist before running the query,
	// to return a clear error message in that case
	if !exist(collectionName, dbInfos) {
		return nil, fmt.Errorf(`collection "%s" doesn't exist`, collectionName)
	}
	collection := db.C(collectionName)

	return runQuery(collection, method, stages)
}

func (s *server) createDatabase(db *mgo.Database, mode byte, config []byte) (dbInfo dbMetaInfo, err error) {

	s.mutex.Lock()
	defer s.mutex.Unlock()

	dbInfo, exists := s.activeDB[db.Name]
	if !exists {

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

		dbInfo = dbMetaInfo{
			collections: make([]string, 0, len(collections)),
		}
		for name := range collections {
			dbInfo.collections = append(dbInfo.collections, name)
		}

		err = fillDatabase(db, collections)
	}

	if err == nil {
		dbInfo.lastUsed = time.Now().Unix()
		s.activeDB[db.Name] = dbInfo
	}
	return dbInfo, err
}

func createContentFromMgodatagen(collections map[string][]bson.M, config []byte) error {

	collConfigs, err := datagen.ParseConfig(config, true)
	if err != nil {
		return err
	}

	mapRef := map[int][][]byte{}
	mapRefType := map[int]byte{}

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

	if bytes.HasPrefix(config, []byte("[")) {

		var docs []bson.M
		err := bson.UnmarshalJSON(config, &docs)

		collections["collection"] = docs

		return err
	}

	if bytes.HasPrefix(config, []byte("db={")) {
		return bson.UnmarshalJSON(config[3:], &collections)
	}

	return errors.New(invalidConfig)
}

func fillDatabase(db *mgo.Database, collections map[string][]bson.M) error {

	if len(collections) > maxCollNb {
		return fmt.Errorf("max number of collection in a database is %d, but was %d", maxCollNb, len(collections))
	}
	// clean any potentially remaining data
	db.DropDatabase()

	names := make(sort.StringSlice, 0, len(collections))
	for name := range collections {
		names = append(names, name)
	}
	names.Sort()

	base := 0
	for _, name := range names {

		bulk := createBulk(db, name)

		docs := collections[name]
		if len(docs) == 0 {
			continue
		}

		for i, doc := range docs {
			if _, hasID := doc["_id"]; !hasID {
				doc["_id"] = seededObjectID(int32(base + i))
			}
			bulk.Insert(doc)
		}

		_, err := bulk.Run()
		if err != nil {
			// In some case, a collection can be partially created even if some write failed
			//
			// for example: [{_id:1},{_id:1}]
			//
			// -> the first write will suceed, but the second will fail, so a collection
			// containing only one record will be created, and an error will be returned
			//
			// Because fillDatabase returns an error, the hash of the database (ie db.name)
			// is not put in server.activeDB, so it can't be deleted from server.removeExpiredDB
			//
			// to avoid this kind of leaks, drop the db immediately if there is an error
			db.DropDatabase()
			return err
		}
		base += len(docs)
	}
	activeDatabases.Inc()

	return nil
}

func createBulk(db *mgo.Database, collectionName string) *mgo.Bulk {
	info := &mgo.CollectionInfo{
		Capped:   true,
		MaxDocs:  maxDoc,
		MaxBytes: maxBytes,
	}
	c := db.C(collectionName)
	c.Create(info)

	bulk := c.Bulk()
	bulk.Unordered()

	return bulk
}

func seededObjectID(n int32) bson.ObjectId {

	// using date = uint32(time.Date(2018, 02, 26, 0, 0, 0, 0, time.UTC).Unix())

	return bson.ObjectId([]byte{
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
	})
}

// query has to match the folowing regex:
//
//   /^db\..(\w*)\.(find|aggregate)\([\s\S]*\)$/
//
// for example:
//
//   db.collection.find({k:1})
//   db.collection.aggregate([{$project:{_id:0}}])
//
//
func parseQuery(query []byte) (collectionName, method string, stages []bson.M, err error) {

	p := bytes.SplitN(query, []byte{'.'}, 3)
	if len(p) != 3 {
		return "", "", nil, errors.New(invalidQuery)
	}

	collectionName = string(p[1])

	// last part of query contains the method and the stages, for example find({k:1})
	queryBytes := p[2]
	start, end := bytes.IndexByte(queryBytes, '('), bytes.LastIndexByte(queryBytes, ')')

	method = string(queryBytes[:start])

	stages, err = unmarshalStages(queryBytes[start+1 : end])
	if err != nil {
		return "", "", nil, fmt.Errorf("fail to parse content of query: %v", err)
	}

	return collectionName, method, stages, nil
}

func runQuery(collection *mgo.Collection, method string, stages []bson.M) (result []byte, err error) {

	var docs []bson.M

	switch method {
	case "find":
		for len(stages) < 2 {
			stages = append(stages, bson.M{})
		}
		err = collection.Find(stages[0]).Select(stages[1]).All(&docs)
	case "aggregate":
		err = collection.Pipe(stages).All(&docs)
	default:
		err = fmt.Errorf("invalid method: %s", method)
	}

	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	if len(docs) == 0 {
		return []byte(noDocFound), nil
	}
	return bson.MarshalExtendedJSON(docs)
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

	err = bson.UnmarshalJSON(queryBytes, &stages)

	return stages, err
}

func exist(collectionName string, dbInfos dbMetaInfo) bool {
	for _, name := range dbInfos.collections {
		if name == collectionName {
			return true
		}
	}
	return false
}
