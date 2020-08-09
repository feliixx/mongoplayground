[![Go Report Card](https://goreportcard.com/badge/github.com/feliixx/mongoplayground)](https://goreportcard.com/report/github.com/feliixx/mongoplayground)
[![codecov](https://codecov.io/gh/feliixx/mongoplayground/branch/master/graph/badge.svg)](https://codecov.io/gh/feliixx/mongoplayground)

# Mongo Playground

Mongo playground: a simple sandbox to test and share MongoDB queries. Try it online : [**https://mongoplayground.net**](https://mongoplayground.net/)


## Limitations

  ### Size limitations

  This playground has several limitations: 

  - a database can't contain more than **10 collections**
  - a collection can't contain more than **100 documents**

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

Favicon was created on [favicon.io](https://favicon.io/) from an emoji provided by [twemoji](https://github.com/twitter/twemoji)
