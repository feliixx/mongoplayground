##   Summary

- [Create a database](#user-content-create-a-database)
  - [with bson documents](#user-content-from-bson-documents)
  - [with random data](#user-content-from-mgodatagen)
- [Limitations](#user-content-limitations)
- [Report an issue / contribute](#user-content-report-an-issue-and-contribute)
- [Credits](#user-content-credits)

# Create a database 

## From BSON documents

It is possible to create a collection from an array of BSON documents. If no collection name is specified, documents will be inserted in a collection named **`collection`**, for example

```JSON5
[
  {
    "_id": 1, 
    "k": "value"
  },
  {
    "_id": 2, 
    "k": "someOtherValue"
  }
]
```

It is possible to create **multiple collections** in `bson` mode with custom names like this

```JSON5
db={
  "coll1": [
    {
      "_id": 1, 
      "k": "value"
    }, 
    {
      "_id": 2, 
      "k": "someOtherValue"
    }
  ], 
  "coll2": [
    {
      "_id": 1, 
      "k2": "value"
    }
  ]
}
```

This will create two collections named `coll1` and `coll2`


## From mgodatagen

You can create random documents using **[mgodatagen](github.com/feliixx/mgodatagen)**. Select `mgodatagen` mode and create a 
custom configuration file. 

The config file is an array of JSON documents, where each documents holds the configuration 
for a collection to create.

```JSON5
[
  // first collection to create 
  {  
   "collection": <string>,            // required, collection name
   "count": <int>,                    // required, number of document to insert in the collection 
   "content": {                       // required, the actual schema to generate documents   
     "fieldName1": <generator>,       // optional, see Generator below
     "fieldName2": <generator>,
     ...
   }
  },
  // second collection to create 
  {
    ...
  }
]
```

## Generator types  

Generators have a common structure: 

```JSON5
"fieldName": {                 // required, field name in generated document
  "type": <string>,            // required, type of the field 
  "nullPercentage": <int>,     // optional, int between 0 and 100. Percentage of documents 
                               // that will have this field
  "maxDistinctValue": <int>,   // optional, maximum number of distinct values for this field
  "typeParam": ...             // specific parameters for this type
}
```

List of main `<generator>` types: 

- [string](#user-content-string)
- [int](#user-content-int)
- [long](#user-content-long)
- [double](#user-content-double)
- [decimal](#user-content-decimal)
- [boolean](#user-content-boolean)
- [objectId](#user-content-objectid)
- [array](#user-content-array)
- [object](#user-content-object)
- [binary](#user-content-binary) 
- [date](#user-content-date) 

List of custom `<generator>` types: 

- [position](#user-content-position)
- [constant](#user-content-constant)
- [autoincrement](#user-content-autoincrement)
- [reference](#user-content-ref)
- [fromArray](#user-content-fromarray)
- [countAggregator](#user-content-countaggregator)
- [valueAggregator](#user-content-valueaggregator)
- [boundAggregator](#user-content-boundaggregator)

List of [Faker](https://github.com/manveru/faker) `<generator>` types: 

- [CellPhoneNumber](#user-content-faker)
- [City](#user-content-faker)
- [CityPrefix](#user-content-faker)
- [CitySuffix](#user-content-faker)
- [CompanyBs](#user-content-faker)
- [CompanyCatchPhrase](#user-content-faker)
- [CompanyName](#user-content-faker)
- [CompanySuffix](#user-content-faker)
- [Country](#user-content-faker)
- [DomainName](#user-content-faker)
- [DomainSuffix](#user-content-faker)
- [DomainWord](#user-content-faker)
- [Email](#user-content-faker)
- [FirstName](#user-content-faker)
- [FreeEmail](#user-content-faker)
- [JobTitle](#user-content-faker)
- [LastName](#user-content-faker)
- [Name](#user-content-faker)
- [NamePrefix](#user-content-faker)
- [NameSuffix](#user-content-faker)
- [PhoneNumber](#user-content-faker)
- [PostCode](#user-content-faker)
- [SafeEmail](#user-content-faker)
- [SecondaryAddress](#user-content-faker)
- [State](#user-content-faker)
- [StateAbbr](#user-content-faker)
- [StreetAddress](#user-content-faker)
- [StreetName](#user-content-faker)
- [StreetSuffix](#user-content-faker)
- [URL](#user-content-faker)
- [UserName](#user-content-faker)



### String

Generate random string of a certain length. String is composed of char within this list: 
`abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_`

```JSON5
"fieldName": {
    "type": "string",          // required
    "nullPercentage": <int>,   // optional 
    "maxDistinctValue": <int>, // optional
    "unique": <bool>,          // optional, see details below 
    "minLength": <int>,        // required,  must be >= 0 
    "maxLength": <int>         // required,  must be >= minLength
}
```

#### Unique String

If `unique` is set to true, the field will only contains unique strings. Unique strings 
have a **fixed length**, `minLength` is taken as length for the string. 
There is  `64^x`  possible unique string for strings of length `x`. This number has to 
be inferior or equal to the number of documents you want to generate. 
For example, if you want unique strings of length 3, there is `64 * 64 * 64 = 262144` possible 
strings.

They will look like 

```
"aaa",
"aab",
"aac",
"aad",
...
```

### Int 

Generate random int within bounds. 

```JSON5
"fieldName": {
    "type": "int",             // required
    "nullPercentage": <int>,   // optional 
    "maxDistinctValue": <int>, // optional
    "minInt": <int>,           // required
    "maxInt": <int>            // required, must be >= minInt
}
```

### Long 

Generate random long within bounds. 

```JSON5
"fieldName": {
    "type": "long",            // required
    "nullPercentage": <int>,   // optional 
    "maxDistinctValue": <int>, // optional
    "minLong": <long>,         // required
    "maxLong": <long>          // required, must be >= minLong
}
```

### Double

Generate random double within bounds. 

```JSON5
"fieldName": {
    "type": "double",          // required
    "nullPercentage": <int>,   // optional
    "maxDistinctValue": <int>, // optional 
    "minDouble": <double>,     // required
    "maxDouble": <double>      // required, must be >= minDouble
}
```

### Decimal

Generate random decimal128.

```JSON5
"fieldName": {
    "type": "decimal",         // required
    "nullPercentage": <int>,   // optional
    "maxDistinctValue": <int>, // optional 
}
```

### Boolean

Generate random boolean.

```JSON5
"fieldName": {
    "type": "boolean",         // required
    "nullPercentage": <int>,   // optional 
    "maxDistinctValue": <int>  // optional
}
```

### ObjectId

Generate random and unique ObjectId.

```JSON5
"fieldName": {
    "type": "objectId",        // required
    "nullPercentage": <int>,   // optional
    "maxDistinctValue": <int>  // optional 
}
```

### Array

Generate a random array of bson object.

```JSON5
"fieldName": {
    "type": "array",             // required
    "nullPercentage": <int>,     // optional
    "maxDistinctValue": <int>,   // optional
    "size": <int>,               // required, size of the array 
    "arrayContent": <generator>  // genrator use to create element to fill the array.
                                 // can be of any type scpecified in generator types
}
```

### Object

Generate random nested object.

```JSON5
"fieldName": {
    "type": "object",                    // required
    "nullPercentage": <int>,             // optional
    "maxDistinctValue": <int>,           // optional
    "objectContent": {                   // required, list of generator used to 
       "nestedFieldName1": <generator>,  // generate the nested document 
       "nestedFieldName2": <generator>,
       ...
    }
}
```

### Binary 

Generate random binary data of length within bounds.

```JSON5
"fieldName": {
    "type": "binary",           // required
    "nullPercentage": <int>,    // optional 
    "maxDistinctValue": <int>,  // optional
    "minLength": <int>,         // required,  must be >= 0 
    "maxLength": <int>          // required,  must be >= minLength
}
```

### Date 

Generate a random date (stored as [`ISODate`](https://docs.mongodb.com/manual/reference/method/Date/) ).

`startDate` and `endDate` are string representation of a Date following RFC3339: 

**format**: "yyyy-MM-ddThh:mm:ss+00:00" or "yyyy-MM-ddThh:mm:ssZ"


```JSON5
"fieldName": {
    "type": "date",            // required
    "nullPercentage": <int>,   // optional 
    "maxDistinctValue": <int>, // optional
    "startDate": <string>,     // required
    "endDate": <string>        // required,  must be >= startDate
}
```

### Position

Generate a random GPS position in Decimal Degrees ( WGS 84).
eg : [40.741895, -73.989308]

```JSON5
"fieldName": {
    "type": "position",         // required
    "nullPercentage": <int>     // optional 
    "maxDistinctValue": <int>   // optional
}
```

### Constant

Add the same value to each document.

```JSON5
"fieldName": {
    "type": "constant",       // required
    "nullPercentage": <int>,  // optional
    "constVal": <object>      // required, can be of any type including object and array
                              // eg: {"k": 1, "v": "val"} 
}
```

### Autoincrement

Create an autoincremented field (type `<long>` or `<int>`).

```JSON5
"fieldName": {
    "type": "autoincrement",  // required
    "nullPercentage": <int>,  // optional
    "autoType": <string>,     // required, can be `int` or `long`
    "startLong": <long>,      // start value if autoType = long
    "startInt": <int>       // start value if autoType = int
}
```

### Ref

If a field reference an other field in an other collection, you can use a ref generator. 

generator in first collection: 

```JSON5
"fieldName":{  
    "type":"ref",               // required
    "nullPercentage": <int>,    // optional
    "maxDistinctValue": <int>,  // optional
    "id": <int>,                // required, generator id used to link
                                // field between collections
    "refContent": <generator>   // required
}
```

generator in other collections: 

```JSON5
"fieldName": {
    "type": "ref",              // required
    "nullPercentage": <int>,    // optional
    "maxDistinctValue": <int>,  // optional
    "id": <int>                 // required, same id as previous generator 
}
```

### FromArray

Randomly pick value from an array as value for the field. Currently, object in the 
array have to be of the same type.


```JSON5
"fieldName": {
    "type": "fromArray",      // required
    "nullPercentage": <int>,  // optional   
    "in": [                   // required. Can't be empty. An array of object of 
      <object>,               // any type, including object and array. 
      <object>
      ...
    ]
}
```
### CountAggregator

Count documents from `<database>.<collection>` matching a specific query. To use a 
variable of the document in the query, prefix it with `$$`.

The query can't be empty or null.



```JSON5
"fieldName": {
  "type": "countAggregator", // required
  "database": <string>,      // required, db to use to perform aggregation
  "collection": <string>,    // required, collection to use to perform aggregation
  "query": <object>          // required, query that selects which documents to count in the collection 
}
```
**Example:**

Assuming that the collection `first` contains: 

```JSON5
{"_id": 1, "field1": 1, "field2": "a" }
{"_id": 2, "field1": 1, "field2": "b" }
{"_id": 3, "field1": 2, "field2": "c" }
```

and that the generator for collection `second` is: 

```JSON5
{
  "database": "test",
  "collection": "second",
  "count": 2,
  "content": {
    "_id": {
      "type": "autoincrement",
      "autoType": "int"
      "startInt": 0
    },
    "count": {
      "type": "countAggregator",
      "database": "test",
      "collection": "first",
      "query": {
        "field1": "$$_id"
      }
    }
  }
}
```

The collection `second` will contain: 

```JSON5
{"_id": 1, "count": 2}
{"_id": 2, "count": 1}
```

### ValueAggregator 

Get distinct values for a specific field for documents from 
`<database>.<collection>` matching a specific query. To use a variable of 
the document in the query, prefix it with `$$`.

The query can't be empty or null.

```JSON5
"fieldName": {
  "type": "valueAggregator", // required
  "database": <string>,      // required, db to use to perform aggregation
  "collection": <string>,    // required, collection to use to perform aggregation
  "key": <string>,           // required, the field for which to return distinct values. 
  "query": <object>          // required, query that specifies the documents from which 
                             // to retrieve the distinct values
}
```

**Example**: 

Assuming that the collection `first` contains: 

```JSON5
{"_id": 1, "field1": 1, "field2": "a" }
{"_id": 2, "field1": 1, "field2": "b" }
{"_id": 3, "field1": 2, "field2": "c" }
```

and that the generator for collection `second` is: 

```JSON5
{
  "database": "test",
  "collection": "second",
  "count": 2,
  "content": {
    "_id": {
      "type": "autoincrement",
      "autoType": "int"
      "startInt": 0
    },
    "count": {
      "type": "valueAggregator",
      "database": "test",
      "collection": "first",
      "key": "field2",
      "values": {
        "field1": "$$_id"
      }
    }
  }
}
```

The collection `second` will contain: 

```JSON5
{"_id": 1, "values": ["a", "b"]}
{"_id": 2, "values": ["c"]}
```


### BoundAggregator 

Get lower ang higher values for a specific field for documents from 
`<database>.<collection>` matching a specific query. To use a variable of 
the document in the query, prefix it with `$$`.

The query can't be empty or null.

```JSON5
"fieldName": {
  "type": "valueAggregator", // required
  "database": <string>,      // required, db to use to perform aggregation
  "collection": <string>,    // required, collection to use to perform aggregation
  "key": <string>,           // required, the field for which to return distinct values. 
  "query": <object>          // required, query that specifies the documents from which 
                             // to retrieve lower/higer value
}
```

**Example**: 

Assuming that the collection `first` contains: 

```JSON5
{"_id": 1, "field1": 1, "field2": "0" }
{"_id": 2, "field1": 1, "field2": "10" }
{"_id": 3, "field1": 2, "field2": "20" }
{"_id": 4, "field1": 2, "field2": "30" }
{"_id": 5, "field1": 2, "field2": "15" }
{"_id": 6, "field1": 2, "field2": "200" }
```

and that the generator for collection `second` is: 

```JSON5
{
  "database": "test",
  "collection": "second",
  "count": 2,
  "content": {
    "_id": {
      "type": "autoincrement",
      "autoType": "int"
      "startInt": 0
    },
    "count": {
      "type": "valueAggregator",
      "database": "test",
      "collection": "first",
      "key": "field2",
      "values": {
        "field1": "$$_id"
      }
    }
  }
}
```

The collection `second` will contain: 

```JSON5
{"_id": 1, "values": {"m": 0, "M": 10}}
{"_id": 2, "values": {"m": 15, "M": 200}}
```

where `m` is the min value, and `M` the max value

### Faker

Generate 'real' data using [Faker library](https://github.com/manveru/faker).

```JSON5
"fieldName": {
    "type": "faker",             // required
    "nullPercentage": <int>,     // optional
    "maxDistinctValue": <int>,   // optional
    "method": <string>           // faker method to use, for example: City / Email...
}
```

If you're building large datasets (1000000+ items) you should avoid faker generators 
and use main or custom generators instead, as faker generator are way slower. 

Currently, only `"en"` locale is available.

# Limitations

### Size limitations

This playground has several limitations: 

 - a database can't contain more than **10 collections**
 - a collection can't contain more than **100 documents**
 - all collections are capped to a size of **1024*100 bytes**, see [mongodb capped collections](https://docs.mongodb.com/manual/core/capped-collections/) for details 

### Queries

Currently, the playground can run only `find()` and `aggregate()` queries. Options in aggregation queries are **not** supported.

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

### Number decimal

Currently, `NumberDecimal()` notation is not supported in bson mode.


## Report an issue and contribute

You can report issues here: [mongoplayground/issues](https://github.com/feliixx/mongoplayground/issues)

The source code is available here: [mongoplayground](https://github.com/feliixx/mongoplayground)

Contributions are welcome! 

# Credits 

This playground is heavily inspired from [The Go Playground](https://play.golang.org)

Editors are created with [ace](https://ace.c9.io/), and the documentation is styled using [github-markdown-css](https://github.com/sindresorhus/github-markdown-css)