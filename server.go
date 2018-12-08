package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/feliixx/mgodatagen/datagen"
	"github.com/feliixx/mgodatagen/datagen/generators"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var (
	templates = template.Must(template.ParseFiles("playground.html"))
)

const (
	staticDir = "static/"
	badgerDir = "storage"
	// interval between two database cleanup
	cleanupInterval = 120 * time.Minute
	// if a database is not used within the last
	// expireInterval, it is removed in the next cleanup
	expireInterval = 60 * time.Minute
)

type server struct {
	mux              *http.ServeMux
	session          *mgo.Session
	storage          *badger.DB
	logger           *log.Logger
	activeDB         sync.Map
	mongodbVersion   []byte
	staticContentMap map[string]int
	staticContent    [][]byte
}

func newServer(logger *log.Logger) (*server, error) {

	session, err := mgo.Dial("mongodb://")
	if err != nil {
		return nil, fmt.Errorf("fail to connect to mongodb: %v", err)
	}
	info, _ := session.BuildInfo()
	version := []byte(info.Version)

	opts := badger.DefaultOptions
	opts.Dir = badgerDir
	opts.ValueDir = badgerDir
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	s := &server{
		mux:            http.DefaultServeMux,
		session:        session,
		storage:        db,
		activeDB:       sync.Map{},
		logger:         logger,
		mongodbVersion: version,
	}

	go func(s *server) {
		for range time.Tick(cleanupInterval) {
			s.removeExpiredDB()
		}
	}(s)

	err = s.precompile()
	if err != nil {
		return nil, err
	}

	s.mux.HandleFunc("/", s.newPageHandler)
	s.mux.HandleFunc("/p/", s.viewHandler)
	s.mux.HandleFunc("/run", s.runHandler)
	s.mux.HandleFunc("/save", s.saveHandler)
	s.mux.HandleFunc("/static/", s.staticHandler)
	s.mux.HandleFunc("/_status/healthcheck", s.healthcheckHandler)
	return s, nil
}

// remove db not used within the last expireInterval
func (s *server) removeExpiredDB() {
	now := time.Now()
	session := s.session.Copy()
	defer session.Close()
	s.activeDB.Range(func(k, v interface{}) bool {
		if now.Sub(time.Unix(v.(int64), 0)) > expireInterval {
			err := session.DB(k.(string)).DropDatabase()
			if err != nil {
				s.logger.Printf("fail to drop database %v: %v", k, err)
				return true
			}
			s.activeDB.Delete(k)
		}
		return true
	})
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// view a saved playground page identified by its ID
func (s *server) viewHandler(w http.ResponseWriter, r *http.Request) {

	id := strings.TrimPrefix(r.URL.Path, "/p/")
	p, err := s.loadPage([]byte(id))
	if err != nil {
		s.logger.Printf("requested page %s doesn't exists", id)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("this playground doesn't exist"))
		return
	}
	err = templates.Execute(w, p)
	if err != nil {
		s.logger.Printf("fail to execute template with page %s: %v", p.String(), err)
		return
	}
}

func (s *server) loadPage(id []byte) (*page, error) {
	p := &page{
		MongoVersion: s.mongodbVersion,
	}
	err := s.storage.View(func(txn *badger.Txn) error {
		item, err := txn.Get(id)
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}
		p.decode(val)
		return nil
	})
	return p, err
}

// run a query and return the results as plain text
func (s *server) runHandler(w http.ResponseWriter, r *http.Request) {

	p := &page{
		Mode:   modeByte(r.FormValue("mode")),
		Config: []byte(r.FormValue("config")),
		Query:  []byte(r.FormValue("query")),
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res, err := s.run(p)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(res)
}

// save the playground and return the playground ID
func (s *server) saveHandler(w http.ResponseWriter, r *http.Request) {

	p := &page{
		Mode:   modeByte(r.FormValue("mode")),
		Config: []byte(r.FormValue("config")),
		Query:  []byte(r.FormValue("query")),
	}

	id, val := p.ID(), p.encode()
	err := s.storage.Update(func(txn *badger.Txn) error {
		return txn.Set(id, val)
	})
	if err != nil {
		s.logger.Printf("fail to save playground %s with id %s", p.String(), id)
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "%sp/%s", r.Referer(), id)
}

// return a playground with the default configuration
func (s *server) newPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Encoding", "gzip")
	w.Write(s.staticContent[0])
}

// serve static ressources (css/js/html)
func (s *server) staticHandler(w http.ResponseWriter, r *http.Request) {

	name := strings.TrimPrefix(r.URL.Path, "/static/")
	sub := strings.Split(name, ".")

	contentType := "text/html; charset=utf-8"
	if len(sub) > 0 {
		switch sub[len(sub)-1] {
		case "css":
			contentType = "text/css; charset=utf-8"
		case "js":
			contentType = "application/javascript; charset=utf-8"
		}
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	pos, ok := s.staticContentMap[name]
	if !ok {
		s.logger.Printf("static resource %s doesn't exist", name)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Write(s.staticContent[pos])
}

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
	invalidConfig = "invalid configuration:\n    must be an array of documents like '[ {_id: 1} ]'\n\n    or\n\n    must match 'db = { collection: [ {_id: 1}, ... ]' }"
)

func (s *server) run(p *page) (result []byte, err error) {

	session := s.session.Copy()
	defer session.Close()

	DBHash := p.dbHash()
	db := session.DB(DBHash)

	_, exists := s.activeDB.Load(DBHash)
	if !exists {

		collections := map[string][]bson.M{}

		switch p.Mode {
		case mgodatagenMode:
			err = createContentFromMgodatagen(collections, p.Config)
		case bsonMode:
			err = loadContentFromJSON(collections, p.Config)
		}

		if err != nil {
			return nil, fmt.Errorf("error in configuration:\n  %v", err)
		}

		err := createDatabase(db, collections)
		if err != nil {
			return nil, err
		}
	}

	s.activeDB.Store(DBHash, time.Now().Unix())

	return runQuery(db, p.Query)
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

func createDatabase(db *mgo.Database, collections map[string][]bson.M) error {

	if len(collections) > maxCollNb {
		return fmt.Errorf("max number of collection in a database is %d, but was %d", maxCollNb, len(collections))
	}
	// clean any potentially remaining data
	db.DropDatabase()

	names := make(sort.StringSlice, len(collections))
	for name := range collections {
		names = append(names, name)
	}
	names.Sort()

	base := 0
	for _, name := range names {

		bulk := createBulk(db, name)

		docs := collections[name]
		if len(docs) > 0 {
			for i, doc := range docs {
				if _, hasID := doc["_id"]; !hasID {
					doc["_id"] = seededObjectID(int32(base + i))
				}
				bulk.Insert(doc)
			}
		}
		_, err := bulk.Run()
		if err != nil {
			return err
		}
		base += len(docs)
	}
	return nil
}

func createBulk(db *mgo.Database, name string) *mgo.Bulk {
	info := &mgo.CollectionInfo{
		Capped:   true,
		MaxDocs:  maxDoc,
		MaxBytes: maxBytes,
	}
	c := db.C(name)
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

func runQuery(db *mgo.Database, query []byte) ([]byte, error) {

	// query should look like
	// db.(\w*).(find|aggregate)(...)
	p := bytes.SplitN(query, []byte{'.'}, 3)
	if len(p) != 3 {
		return nil, fmt.Errorf("invalid query: \nmust match db.coll.find(...) or db.coll.aggregate(...)")
	}

	collection := db.C(string(p[1]))

	if !exist(collection) {
		return nil, fmt.Errorf(`collection "%s" doesn't exist`, p[1])
	}

	// last part of query contains the method and the stages, for example find({k:1})
	queryBytes := p[2]
	start, end := bytes.IndexByte(queryBytes, '('), bytes.LastIndexByte(queryBytes, ')')

	method := string(queryBytes[:start])

	stages, err := stages(queryBytes[start+1 : end])
	if err != nil {
		return nil, fmt.Errorf("fail to parse content of query: %v", err)
	}

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

func stages(queryBytes []byte) (stages []bson.M, err error) {

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

func exist(collection *mgo.Collection) bool {
	names, err := collection.Database.CollectionNames()
	if err != nil {
		return true
	}
	for _, name := range names {
		if name == collection.Name {
			return true
		}
	}
	return false
}

const (
	templateConfig = `[
  {
    "collection": "collection",
    "count": 10,
    "content": {
		"k": {
		  "type": "int",
		  "minInt": 0, 
		  "maxInt": 10
		}
	}
  }
]`
	templateQuery = "db.collection.find()"
)

// load static ressources (javascript, css, docs and default page)
// and compress them in order to serve them faster
func (s *server) precompile() error {

	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	zw.Name = "playground.html"
	zw.ModTime = time.Now()
	p := &page{
		Mode:         bsonMode,
		Config:       []byte(templateConfig),
		Query:        []byte(templateQuery),
		MongoVersion: s.mongodbVersion,
	}
	if err := templates.Execute(zw, p); err != nil {
		return err
	}
	if err := s.add(zw, &buf, 0); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(staticDir)
	if err != nil {
		return err
	}
	for i, f := range files {
		buf.Reset()
		zw.Reset(&buf)
		zw.Name = f.Name()
		zw.ModTime = time.Now()
		b, err := ioutil.ReadFile(staticDir + f.Name())
		if err != nil {
			return err
		}
		if _, err = zw.Write(b); err != nil {
			return err
		}
		if err := s.add(zw, &buf, i+1); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) add(zw *gzip.Writer, buf *bytes.Buffer, index int) error {
	if s.staticContent == nil {
		s.staticContent = make([][]byte, 0)
		s.staticContentMap = map[string]int{}
	}
	if err := zw.Close(); err != nil {
		return err
	}
	c := make([]byte, buf.Len())
	copy(c, buf.Bytes())
	s.staticContentMap[zw.Name] = index
	s.staticContent = append(s.staticContent, c)
	return nil
}

var statusOK = []byte(`{"status":"ok"}"`)

func (s *server) healthcheckHandler(w http.ResponseWriter, r *http.Request) {

	p := &page{
		Mode:   bsonMode,
		Config: []byte(`[{"_id":1}]`),
		Query:  []byte(templateQuery),
	}

	w.Header().Set("Content-Type", "encoding/json")

	result, err := s.run(p)
	if err != nil || bytes.Compare(bytes.TrimSuffix(result, []byte("\n")), p.Config) != 0 {
		fmt.Fprintf(w, `{"status":"unexpected result: (err: %v, result: %s"}`, err, result)
		return
	}
	w.Write(statusOK)
}
