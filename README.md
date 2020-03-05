[![Go Report Card](https://goreportcard.com/badge/github.com/feliixx/mongoplayground)](https://goreportcard.com/report/github.com/feliixx/mongoplayground)
[![codecov](https://codecov.io/gh/feliixx/mongoplayground/branch/master/graph/badge.svg)](https://codecov.io/gh/feliixx/mongoplayground)

# Mongo Playground

Mongo playground: a simple sandbox to test and share MongoDB queries. Try it online : [**mongoplayground**](https://mongoplayground.net/)


## Limitations

  ### Size limitations

  This playground has several limitations: 

  - a database can't contain more than **10 collections**
  - a collection can't contain more than **100 documents**
  - all collections are capped to a size of **100*1024 bytes**, see [mongodb capped collections](https://docs.mongodb.com/manual/core/capped-collections/) for details 

  ### Queries

  Currently, the playground can run only `find()` and `aggregate()` queries 

  ### shell regex

  Currently, shell regex doesn't work in query. 

  so instead of 

  ```JSON5
  db.collection.find({
    "k": /pattern/
  })
  ```

  use 

  ```JSON5 
  db.collection.find({
    "k": {
      "$regex": "pattern"
    }
  })
  ```

## Credits 

This playground is heavily inspired from [The Go Playground](https://play.golang.org)

Editors are created with [ace](https://ace.c9.io/), and the documentation is styled using [github-markdown-css](https://github.com/sindresorhus/github-markdown-css)