var templates = [
    {
        config: '[{"key":1},{"key":2}]',
        query: 'db.collection.find()',
        mode: 'bson'
    },
    {
        config: 'db={"orders":[{"_id":1,"item":"almonds","price":12,"quantity":2},{"_id":2,"item":"pecans","price":20,"quantity":1},{"_id":3}],"inventory":[{"_id":1,"sku":"almonds","description":"product 1","instock":120},{"_id":2,"sku":"bread","description":"product 2","instock":80},{"_id":3,"sku":"cashews","description":"product 3","instock":60},{"_id":4,"sku":"pecans","description":"product 4","instock":70},{"_id":5,"sku":null,"description":"Incomplete"},{"_id":6}]}',
        query: 'db.orders.aggregate([{"$lookup":{"from":"inventory","localField":"item","foreignField":"sku","as":"inventory_docs"}}])',
        mode: 'bson'
    },
    {
        config: '[{"collection":"collection","count":10,"content":{"key":{"type":"int","minInt":0,"maxInt":10}}}]',
        query: 'db.collection.find()',
        mode: 'mgodatagen'
    },
    {
        config: '[{"key":1},{"key":2}]',
        query: 'db.collection.update({"key":2},{"$set":{"updated":true}},{"multi":false,"upsert":false})',
        mode: 'bson'
    },
    {
        config: '[{"collection":"collection","count":5,"content":{"description":{"type":"fromArray","in":["Coffee and cakes","Gourmet hamburgers","Just coffee","Discount clothing","Indonesian goods"]}},"indexes":[{"name":"description_text_idx","key":{"description":"text"}}]}]',
        query: 'db.collection.find({"$text":{"$search":"coffee"}})',
        mode: 'mgodatagen'
    },
    {
        config: '[{"_id":1,"item":"ABC","price":80,"sizes":["S","M","L"]},{"_id":2,"item":"EFG","price":120,"sizes":[]},{"_id":3,"item":"IJK","price":160,"sizes":"M"},{"_id":4,"item":"LMN","price":10},{"_id":5,"item":"XYZ","price":5.75,"sizes":null}]',
        query: 'db.collection.aggregate([{"$unwind":{"path":"$sizes","preserveNullAndEmptyArrays":true}},{"$group":{"_id":"$sizes","averagePrice":{"$avg":"$price"}}},{"$sort":{"averagePrice":-1}}]).explain("executionStats")',
        mode: 'bson'
    }
]

var methodSnippet = [
    {
        caption: "find()",
        value: "find()",
        meta: "method"
    },
    {
        caption: "aggregate()",
        value: "aggregate()",
        meta: "method"
    },
    {
        caption: "update()",
        value: "update()",
        meta: "method"
    },
    {
        caption: "explain()",
        value: "explain()",
        meta: "method"
    },
]

var availableCollections = [
    {
        caption: "collection",
        value: "collection",
        meta: "collection name"
    }
]


var basicBsonSnippet = [
    {
        caption: "true",
        value: "true",
        meta: "bson keyword"
    },
    {
        caption: "false",
        value: "false",
        meta: "bson keyword"
    },
    {
        caption: "null",
        value: "null",
        meta: "bson keyword"
    },
    {
        caption: "$numberDecimal",
        value: "$numberDecimal: ",
        meta: "bson keyword"
    },
    {
        caption: "$numberDouble",
        value: "$numberDouble: ",
        meta: "bson keyword"
    },
    {
        caption: "$numberLong",
        value: "$numberLong: ",
        meta: "bson keyword"
    },
    {
        caption: "$numberInt",
        value: "$numberLong: ",
        meta: "bson keyword"
    },
    {
        caption: "$oid",
        value: "$oid: ",
        meta: "bson keyword"
    },
    {
        caption: "$regularExpression",
        value: '$regularExpression: {\n "pattern": "pattern",\n "options": "options"\n}',
        meta: "bson keyword"
    },
    {
        caption: "$timestamp",
        value: '$timestamp: {"t": 0, "i": 1}',
        meta: "bson keyword"
    },
    {
        caption: "$date",
        value: "$date: ",
        meta: "bson keyword"
    }
]

var querySnippet = [
    {
        caption: "$eq",
        value: '$eq: "value"',
        meta: "comparison operator"
    },
    {
        caption: "$gt",
        value: '$gt: "value"',
        meta: "comparison operator"
    },
    {
        caption: "$gte",
        value: '$gte: "value"',
        meta: "comparison operator"
    },
    {
        caption: "$in",
        value: '$in: ["value1", "value2"]',
        meta: "comparison operator"
    },
    {
        caption: "$let",
        value: '$let: {\n "vars": { "var": "expression" },\n "in": "expression"\n}',
        meta: "variable operator"
    },
    {
        caption: "$lt",
        value: '$lt: "value"',
        meta: "comparison operator"
    },
    {
        caption: "$lte",
        value: '$lte: "value"',
        meta: "comparison operator"
    },
    {
        caption: "$ne",
        value: '$ne: "value"',
        meta: "comparison operator"
    },
    {
        caption: "$nin",
        value: '$nin: ["value1", "value2"',
        meta: "comparison operator"
    },
    {
        caption: "$not",
        value: "$not: { }",
        meta: "logical operator"
    },
    {
        caption: "$nor",
        value: '$nor: [ { "expression1" }, { "expression2" } ]',
        meta: "logical operator"
    },
    {
        caption: "$and",
        value: '$and: [ { "expression1" }, { "expression2" } ]',
        meta: "logical operator"
    },
    {
        caption: "$or",
        value: '$or: [ { "expression1" }, { "expression2" } ]',
        meta: "logical operator"
    },
    {
        caption: "$exists",
        value: '$exists: "value"',
        meta: "element operator"
    },
    {
        caption: "$type",
        value: '$type: "bson type"',
        meta: "element operator"
    },
    {
        caption: "$expr",
        value: '$expr: { "expression" }',
        meta: "evaluation operator"
    },
    {
        caption: "$jsonSchema",
        value: '$jsonSchema: { "schema" }',
        meta: "evaluation operator"
    },
    {
        caption: "$mod",
        value: '$mod: [ "divisor", "remainder" ]',
        meta: "evaluation operator"
    },
    {
        caption: "$regex",
        value: '$regex: "pattern"',
        meta: "evaluation operator"
    },
    {
        caption: "$where",
        value: '$where: "code"',
        meta: "evaluation operator"
    },
    {
        caption: "$geoIntersects",
        value: '$geoIntersects: {\n "$geometry": {\n  "type": "GeoJSON type",\n  "coordinates": [  ]\n }\n}',
        meta: "geospatial operator"
    },
    {
        caption: "$geoWithin",
        value: '$geoWithin: {\n "$geometry": {\n  "type": "Polygon",\n  "coordinates": [  ]\n }\n}',
        meta: "geospatial operator"
    },
    {
        caption: "$near",
        value: '$near: {\n "$geometry": {\n  "type": "Point",\n  "coordinates": [ "long", "lat" ]\n }, "$maxDistance": 10, "$minDistance": 1\n}',
        meta: "geospatial operator"
    },
    {
        caption: "$nearSphere",
        value: '$nearSphere: {\n "$geometry": {\n  "type": "Point",\n  "coordinates": [ "long", "lat" ]\n }, "$maxDistance": 10, "$minDistance": 1\n}',
        meta: "geospatial operator"
    },
    {
        caption: "$box",
        value: "$box:  [ [ 0, 0 ], [ 100, 100 ] ]",
        meta: "geospatial operator"
    },
    {
        caption: "$center",
        value: '$center: [ [ "x", "y" ] , "radius" ]',
        meta: "geospatial operator"
    },
    {
        caption: "$centerSphere",
        value: '$centerSphere: [ [ "x", "y" ] , "radius" ]',
        meta: "geospatial operator"
    },
    {
        caption: "$geometry",
        value: '$geometry: {\n "type": "Polygon",\n "coordinates": [ ]\n}',
        meta: "geospatial operator"
    },
    {
        caption: "$maxDistance",
        value: "$maxDistance: 10",
        meta: "geospatial operator"
    },
    {
        caption: "$minDistance",
        value: "$minDistance: 10",
        meta: "geospatial operator"
    },
    {
        caption: "$polygon",
        value: "$polygon: [ [ 0 , 0 ], [ 3 , 6 ], [ 6 , 0 ] ]",
        meta: "geospatial operator"
    },
    {
        caption: "$all",
        value: '$all: [ "value1" , "value2" ]',
        meta: "array operator"
    },
    {
        caption: "$elemMatch",
        value: '$elemMatch: { "query1", "query2" }',
        meta: "array operator"
    },
    {
        caption: "$size",
        value: "$size: 1",
        meta: "array operator"
    },
    {
        caption: "$bitsAllClear",
        value: '$bitsAllClear: [ "pos1", "pos2" ]',
        meta: "bitwise operator"
    },
    {
        caption: "$bitsAllSet",
        value: '$bitsAllSet: [ "pos1", "pos2" ]',
        meta: "bitwise operator"
    },
    {
        caption: "$bitsAnyClear",
        value: '$bitsAnyClear: [ "pos1", "pos2" ]',
        meta: "bitwise operator"
    },
    {
        caption: "$bitsAnySet",
        value: '$bitsAnySet: [ "pos1", "pos2" ]',
        meta: "bitwise operator"
    },
    {
        caption: "$slice",
        value: "$slice: 2",
        meta: "projection operator"
    },
]

var aggregationSnippet = [
    {
        caption: "$abs",
        value: '$abs: -1',
        meta: "arithmetic operator"
    },
    {
        caption: "$accumulator",
        value: '$accumulator: {\n "init": "code",\n "initArgs": "array expression",\n "accumulate": "code",\n "accumulateArgs": "array expression",\n "merge": "code",\n "finalize": "code",\n "lang": "string"\n}',
        meta: "accumulation operator"
    },
    {
        caption: "$acos",
        value: '$acos: "expression"',
        meta: "trigonometry operator"
    },
    {
        caption: "$acosh",
        value: '$acosh: "expression"',
        meta: "trigonometry operator"
    },
    {
        caption: "$add",
        value: '$add: [ "expression1", "expression2" ]',
        meta: "arithmetic operator"
    },
    {
        caption: "$addFields",
        value: '$addFields: { "newField": "expression" }',
        meta: "aggregation stage"
    },
    {
        caption: "$addToSet",
        value: '$addToSet: "expression"',
        meta: "accumulation operator"
    },
    {
        caption: "$allElementsTrue",
        value: '$allElementsTrue: [ "expression" ]',
        meta: "set operator"
    },
    {
        caption: "$and",
        value: '$and: [ "expression1", "expression2" ]',
        meta: "boolean operator"
    },
    {
        caption: "$anyElementTrue",
        value: '$anyElementTrue: [ "expression" ]',
        meta: "set operator"
    },
    {
        caption: "$arrayElemAt",
        value: '$arrayElemAt: [ "array", "idx" ]',
        meta: "array operator"
    },
    {
        caption: "$arrayToObject",
        value: '$arrayToObject: "expression"',
        meta: "array operator"
    },
    {
        caption: "$asin",
        value: '$asin: "expression"',
        meta: "trigonometry operator"
    },
    {
        caption: "$asinh",
        value: '$asinh: "expression"',
        meta: "trigonometry operator"
    },
    {
        caption: "$atan",
        value: '$atan: "expression"',
        meta: "trigonometry operator"
    },
    {
        caption: "$atan2",
        value: '$atan2: [ "expression 1", "expression 2" ]',
        meta: "trigonometry operator"
    },
    {
        caption: "$atanh",
        value: '$atanh: "expression"',
        meta: "trigonometry operator"
    },
    {
        caption: "$avg",
        value: '$avg: "expression"',
        meta: "accumulation operator"
    },
    {
        caption: "$binarySize",
        value: '$binarySize: "string or binData"',
        meta: "size operator"
    },
    {
        caption: "$bsonSize",
        value: '$bsonSize: "object"',
        meta: "size operator"
    },
    {
        caption: "$bucket",
        value: '$bucket: {\n "groupBy": "expression",\n "boundaries": [ "lowerbound1", "lowerbound2" ],\n "default": "literal",\n "output": {\n  "output1": "$accumulator expression",\n  "outputN": "$accumulator expression"\n }\n}',
        meta: "aggregation stage"
    },
    {
        caption: "$bucketAuto",
        value: '$bucketAuto: {\n "groupBy": "expression",\n "buckets": 2,\n "output": {\n "output1": "$accumulator expression"},\n "granularity": "string"\n}',
        meta: "aggregation stage"
    },
    {
        caption: "$ceil",
        value: '$ceil: 3.3',
        meta: "arithmetic operator"
    },
    {
        caption: "$cmp",
        value: '$cmp: [ "expression1", "expression2" ]',
        meta: "comparison operator"
    },
    {
        caption: "$concat",
        value: '$concat: [ "expression1", "expression2" ]',
        meta: "string operator"
    },
    {
        caption: "$concatArrays",
        value: '$concatArrays: [ "array1", "array2" ]',
        meta: "array operator"
    },
    {
        caption: "$cond",
        value: '$cond: {\n "if": "boolean-expression",\n "then": "true-case",\n "else": "false-case" }',
        meta: "conditional operator"
    },
    {
        caption: "$convert",
        value: '$convert: {\n "input": "expression",\n "to": "type expression",\n "onError": "expression",\n "onNull": "expression"\n}',
        meta: "type operator"
    },
    {
        caption: "$cos",
        value: '$cos: "expression"',
        meta: "trigonometry operator"
    },
    {
        caption: "$count",
        value: '$count: "string"',
        meta: "aggregation stage"
    },
    {
        caption: "$dateFromParts",
        value: '$dateFromParts : {\n "year": "year", "month": "month", "day": "day",\n "hour": "hour", "minute": "minute", "second": "second",\n "millisecond": "ms", "timezone": "tzExpression"\n}',
        meta: "date operator"
    },
    {
        caption: "$dateFromString",
        value: '$dateFromString: {\n "dateString": "dateStringExpression",\n "format": "formatStringExpression",\n "timezone": "tzExpression",\n "onError": "onErrorExpression",\n "onNull": "onNullExpression"\n}',
        meta: "string operator"
    },
    {
        caption: "$dateToParts",
        value: '$dateToParts: {\n "date" : "dateExpression",\n "timezone" : "timezone",\n "iso8601" : "boolean"\n}',
        meta: "date operator"
    },
    {
        caption: "$dateToString",
        value: '$dateToString: {\n "date": "dateExpression",\n "format": "formatString",\n "timezone": "tzExpression",\n "onNull": "expression"\n}',
        meta: "string operator"
    },
    {
        caption: "$dayOfMonth",
        value: '$dayOfMonth: "dateExpression"',
        meta: "date operator"
    },
    {
        caption: "$dayOfWeek",
        value: '$dayOfWeek: "dateExpression"',
        meta: "date operator"
    },
    {
        caption: "$dayOfYear",
        value: '$dayOfYear: "dateExpression"',
        meta: "date operator"
    },
    {
        caption: "$degreesToRadians",
        value: '$degreesToRadians: "expression"',
        meta: "trigonometry operator"
    },
    {
        caption: "$divide",
        value: '$divide: [ "expression1", "expression2" ]',
        meta: "arithmetic operator"
    },
    {
        caption: "$eq",
        value: '$eq: [ "expression1", "expression2" ]',
        meta: "comparison operator"
    },
    {
        caption: "$exists",
        value: "$exists: true",
        meta: "aggregation operator"
    },
    {
        caption: "$exp",
        value: '$exp: "exponent"',
        meta: "arithmetic operator"
    },
    {
        caption: "$facet",
        value: '$facet:\n{\n "outputField1": [ "stage1", "stage2" ]\n}',
        meta: "aggregation stage"
    },
    {
        caption: "$filter",
        value: '$filter: { "input": "array", "as": "string", "cond": "expression" }',
        meta: "array operator"
    },
    {
        caption: "$first",
        value: '$first: "expression"',
        meta: "array operator"
    },
    {
        caption: "$floor",
        value: '$floor: 1',
        meta: "arithmetic operator"
    },
    {
        caption: "$function",
        value: '$function: {\n "body": "code",\n "args": "array expression",\n "lang": "js"\n}',
        meta: "aggregation operator"
    },
    {
        caption: "$geoNear",
        value: '$geoNear: { "TODO" }',
        meta: "aggregation stage"
    },
    {
        caption: "$graphLookup",
        value: '$graphLookup: {\n "from": "collection",\n "startWith": "expression",\n "connectFromField": "string",\n "connectToField": "string",\n "as": "string",\n "maxDepth": 2,\n "depthField": "string",\n "restrictSearchWithMatch": "document"\n}',
        meta: "aggregation stage"
    },
    {
        caption: "$group",
        value: '$group: {\n "_id": "group by expression",\n "field": { "accumulator" : "expression" }\n}',
        meta: "aggregation stage"
    },
    {
        caption: "$gt",
        value: '$gt: [ "expression1", "expression2" ]',
        meta: "comparison operator"
    },
    {
        caption: "$gte",
        value: '$gte: [ "expression1", "expression2" ]',
        meta: "comparison operator"
    },
    {
        caption: "$hour",
        value: '$hour: "dateExpression"',
        meta: "date operator"
    },
    {
        caption: "$ifNull",
        value: '$ifNull: [ "expression", "replacement-expression-if-null" ]',
        meta: "conditional operator"
    },
    {
        caption: "$in",
        value: '$in: [ "expression", "array expression" ]',
        meta: "array operator"
    },
    {
        caption: "$indexOfArray",
        value: '$indexOfArray: [ "array expression", "search expression", "start", "end" ]',
        meta: "array operator"
    },
    {
        caption: "$indexOfBytes",
        value: '$indexOfBytes: [ "string expression", "substring expression", "start", "end" ]',
        meta: "string operator"
    },
    {
        caption: "$indexOfCP",
        value: '$indexOfCP: [ "string expression", "substring expression", "start", "end" ]',
        meta: "string operator"
    },
    {
        caption: "$isArray",
        value: '$isArray: [ "expression" ]',
        meta: "array operator"
    },
    {
        caption: "$isNumber",
        value: '$isNumber: "expression"',
        meta: "type operator"
    },
    {
        caption: "$isoDayOfWeek",
        value: '$isoDayOfWeek: "dateExpression"',
        meta: "date operator"
    },
    {
        caption: "$isoWeek",
        value: '$isoWeek: "dateExpression"',
        meta: "date operator"
    },
    {
        caption: "$isoWeekYear",
        value: '$isoWeekYear: "dateExpression"',
        meta: "date operator"
    },
    {
        caption: "$last",
        value: '$last: "expression"',
        meta: "array operator"
    },
    {
        caption: "$limit",
        value: '$limit: "positive integer"',
        meta: "aggregation stage"
    },
    {
        caption: "$ln",
        value: '$ln: 10',
        meta: "arithmetic operator"
    },
    {
        caption: "$log",
        value: '$log: [ 100, 10 ]',
        meta: "arithmetic operator"
    },
    {
        caption: "$log10",
        value: '$log10: 4',
        meta: "arithmetic operator"
    },
    {
        caption: "$lookup",
        value: '$lookup: {\n "from": "collection to join",\n "localField": "field from the input documents",\n "foreignField": "field from the documents of the from collection",\n "as": "output array field"\n}',
        meta: "aggregation stage"
    },
    {
        caption: "$lt",
        value: '$lt: [ "expression1", "expression2" ]',
        meta: "comparison operator"
    },
    {
        caption: "$lte",
        value: '$lte: [ "expression1", "expression2" ]',
        meta: "comparison operator"
    },
    {
        caption: "$ltrim",
        value: '$ltrim: { "input": "string",  "chars": "string" }',
        meta: "string operator"
    },
    {
        caption: "$map",
        value: '$map: { "input": "expression", "as": "string", "in": "expression" }',
        meta: "array operator"
    },
    {
        caption: "$match",
        value: "$match: { }",
        meta: "aggregation stage"
    },
    {
        caption: "$max",
        value: '$max: "expression"',
        meta: "accumulation operator"
    },
    {
        caption: "$merge",
        value: '$merge: {\n "into": "collection",\n "on": "identifier field",\n "let": "variables",\n "whenMatched": "replace|keepExisting|merge|fail|pipeline",\n "whenNotMatched": "insert|discard|fail"\n}',                     // Optional\n}",
        meta: "aggregation stage"
    },
    {
        caption: "$mergeObjects",
        value: '$mergeObjects: "document"',
        meta: "object operator"
    },
    {
        caption: "$millisecond",
        value: '$millisecond: "dateExpression"',
        meta: "date operator"
    },
    {
        caption: "$min",
        value: '$min: "expression"',
        meta: "accumulation operator"
    },
    {
        caption: "$minute",
        value: '$minute: "dateExpression"',
        meta: "date operator"
    },
    {
        caption: "$mod",
        value: '$mod: [ "expression1", "expression2" ]',
        meta: "arithmetic operator"
    },
    {
        caption: "$month",
        value: '$month: "dateExpression"',
        meta: "date operator"
    },
    {
        caption: "$multiply",
        value: '$multiply: [ "expression1", "expression2" ]',
        meta: "arithmetic operator"
    },
    {
        caption: "$ne",
        value: '$ne: [ "expression1", "expression2" ]',
        meta: "comparison operator"
    },
    {
        caption: "$not",
        value: '$not: [ "expression" ]',
        meta: "boolean operator"
    },
    {
        caption: "$objectToArray",
        value: '$objectToArray: "object"',
        meta: "object operator"
    },
    {
        caption: "$or",
        value: '$or: [ "expression1", "expression2" ]',
        meta: "boolean operator"
    },
    {
        caption: "$out",
        value: '$out: { "db": "output-db", "coll": "output-collection" }',
        meta: "aggregation stage"
    },
    {
        caption: "$pow",
        value: '$pow: [ "number", "exponent" ]',
        meta: "arithmetic operator"
    },
    {
        caption: "$project",
        value: "$project: { }",
        meta: "aggregation stage"
    },
    {
        caption: "$push",
        value: '$push: "expression"',
        meta: "accumulation operator"
    },
    {
        caption: "$radiansToDegrees",
        value: '$radiansToDegrees: "expression"',
        meta: "trigonometry operator"
    },
    {
        caption: "$range",
        value: '$range: [ "start", "end", "non-zero step" ]',
        meta: "array operator"
    },
    {
        caption: "$redact",
        value: '$redact: "expression"',
        meta: "aggregation stage"
    },
    {
        caption: "$reduce",
        value: '$reduce: { "input": "array", "initialValue": "expression", "in": "expression" }',
        meta: "array operator"
    },
    {
        caption: "$regexFind",
        value: '$regexFind: { "input": "expression", "regex": "expression", "options": "expression" }',
        meta: "string operator"
    },
    {
        caption: "$regexFindAll",
        value: '$regexFindAll: { "input": "expression", "regex": "expression", "options": "expression" }',
        meta: "string operator"
    },
    {
        caption: "$regexMatch",
        value: '$regexMatch: { "input": "expression" , "regex": "expression", "options": "expression" }',
        meta: "string operator"
    },
    {
        caption: "$replaceAll",
        value: '$replaceAll: { "input": "expression", "find": "expression", "replacement": "expression" }',
        meta: "string operator"
    },
    {
        caption: "$replaceOne",
        value: '$replaceOne: { "input": "expression", "find": "expression", "replacement": "expression" }',
        meta: "string operator"
    },
    {
        caption: "$replaceRoot",
        value: '$replaceRoot: { "newRoot": "replacementDocument" }',
        meta: "aggregation stage"
    },
    {
        caption: "$replaceWith",
        value: '$replaceWith: "replacementDocument"',
        meta: "aggregation stage"
    },
    {
        caption: "$reverseArray",
        value: '$reverseArray: "array expression"',
        meta: "array operator"
    },
    {
        caption: "$round",
        value: '$round : [ "number", "place" ]',
        meta: "arithmetic operator"
    },
    {
        caption: "$rtrim",
        value: '$rtrim: { "input": "string", chars: "string" }',
        meta: "string operator"
    },
    {
        caption: "$sample",
        value: '$sample: { "size": "positive integer" }',
        meta: "aggregation stage"
    },
    {
        caption: "$second",
        value: '$second: "dateExpression"',
        meta: "date operator"
    },
    {
        caption: "$set",
        value: '$set: { "newField": "expression" }',
        meta: "aggregation stage"
    },
    {
        caption: "$setDifference",
        value: '$setDifference: [ "expression1", "expression2" ]',
        meta: "set operator"
    },
    {
        caption: "$setEquals",
        value: '$setEquals: [ "expression1", "expression2" ]',
        meta: "set operator"
    },
    {
        caption: "$setIntersection",
        value: '$setIntersection: [ "array1", "array2" ]',
        meta: "set operator"
    },
    {
        caption: "$setIsSubset",
        value: '$setIsSubset: [ "expression1", "expression2" ]',
        meta: "set operator"
    },
    {
        caption: "$setUnion",
        value: '$setUnion: [ "expression1", "expression2" ]',
        meta: "set operator"
    },
    {
        caption: "$sin",
        value: '$sin: "expression"',
        meta: "trigonometry operator"
    },
    {
        caption: "$size",
        value: '$size: "expression"',
        meta: "array operator"
    },
    {
        caption: "$skip",
        value: "$skip",
        meta: "aggregation stage"
    },
    {
        caption: "$slice",
        value: '$slice: [ "array", "n" ]',
        meta: "array operator"
    },
    {
        caption: "$sort:",
        value: "$sort: { }",
        meta: "aggregation stage"
    },
    {
        caption: "$sortByCount",
        value: '$sortByCount:  "expression"',
        meta: "aggregation stage"
    },
    {
        caption: "$split",
        value: '$split: [ "string expression", "delimiter" ]',
        meta: "string operator"
    },
    {
        caption: "$sqrt",
        value: '$sqrt: 12',
        meta: "arithmetic operator"
    },
    {
        caption: "$stdDevPop",
        value: '$stdDevPop: "expression"',
        meta: "accumulation operator"
    },
    {
        caption: "$stdDevSamp",
        value: '$stdDevSamp: "expression"',
        meta: "accumulation operator"
    },
    {
        caption: "$strLenBytes",
        value: '$strLenBytes: "string expression"',
        meta: "string operator"
    },
    {
        caption: "$strLenCP",
        value: '$strLenCP: "string expression"',
        meta: "string operator"
    },
    {
        caption: "$strcasecmp",
        value: '$strcasecmp: [ "expression1", "expression2" ]',
        meta: "string operator"
    },
    {
        caption: "$substr",
        value: '$substr: [ "string", "start", "length" ]',
        meta: "string operator"
    },
    {
        caption: "$substrBytes",
        value: '$substrBytes: [ "string expression", "byte index", "byte count" ]',
        meta: "string operator"
    },
    {
        caption: "$substrCP",
        value: '$substrCP: [ "string expression", "code point index", "code point count" ]',
        meta: "string operator"
    },
    {
        caption: "$subtract",
        value: '$subtract: [ "expression1", "expression2" ]',
        meta: "arithmetic operator"
    },
    {
        caption: "$sum",
        value: '$sum: "expression"',
        meta: "accumulation operator"
    },
    {
        caption: "$switch",
        value: '$switch: {\n "branches": [\n { "case": "expression", "then": "expression" } \n]\n}',
        meta: "conditional operator"
    },
    {
        caption: "$tan",
        value: '$tan: "expression"',
        meta: "trigonometry operator"
    },
    {
        caption: "$toBool",
        value: '$toBool: "expression"',
        meta: "type operator"
    },
    {
        caption: "$toDate",
        value: '$toDate: "expression"',
        meta: "type operator"
    },
    {
        caption: "$toDecimal",
        value: '$toDecimal: "expression"',
        meta: "type operator"
    },
    {
        caption: "$toDouble",
        value: '$toDouble: "expression"',
        meta: "type operator"
    },
    {
        caption: "$toInt",
        value: '$toInt: "expression"',
        meta: "type operator"
    },
    {
        caption: "$toLong",
        value: '$toLong: "expression"',
        meta: "type operator"
    },
    {
        caption: "$toLower",
        value: '$toLower: "expression"',
        meta: "string operator"
    },
    {
        caption: "$toObjectId",
        value: '$toObjectId: "expression"',
        meta: "type operator"
    },
    {
        caption: "$toString",
        value: '$toString: "expression"',
        meta: "type operator"
    },
    {
        caption: "$toUpper",
        value: '$toUpper: "expression"',
        meta: "string operator"
    },
    {
        caption: "$trim",
        value: '$trim: { "input": "string",  "chars": "string" }',
        meta: "string operator"
    },
    {
        caption: "$trunc",
        value: '$trunc : [ "number", "place" ]',
        meta: "arithmetic operator"
    },
    {
        caption: "$type",
        value: '$type: "expression"',
        meta: "type operator"
    },
    {
        caption: "$unionWith",
        value: '$unionWith: { "coll": "collection", "pipeline": [ "stage1" ] }',
        meta: "aggregation stage"
    },
    {
        caption: "$unset",
        value: '$unset: "field"',
        meta: "aggregation stage"
    },
    {
        caption: "$unwind",
        value: '$unwind: "field path"',
        meta: "aggregation stage"
    },
    {
        caption: "$week",
        value: '$week: "dateExpression"',
        meta: "date operator"
    },
    {
        caption: "$where",
        value: '$where: "code"',
        meta: "aggregation operator"
    },
    {
        caption: "$year",
        value: '$year: "dateExpression"',
        meta: "date operator"
    },
    {
        caption: "$zip",
        value: '$zip: {\n "inputs": [ "array expression1" ],\n "useLongestLength": "boolean",\n "defaults":  "array expression"\n}',
        meta: "array operator"
    }]


var updateSnippet = [
    {
        caption: "$currentDate",
        value: '$currentDate: "expression"',
        meta: "update operator"
    },
    {
        caption: "$inc",
        value: '$inc: { "field": 1 }',
        meta: "update operator"
    },
    {
        caption: "$min",
        value: '$min: "expression"',
        meta: "update operator"
    },
    {
        caption: "$max",
        value: '$max: "expression"',
        meta: "update operator"
    },
    {
        caption: "$mul",
        value: '$mul: { "field": 2 }',
        meta: "update operator"
    },
    {
        caption: "$rename",
        value: '$rename: { "field": "newName" }',
        meta: "update operator"
    },
    {
        caption: "$set",
        value: '$set: { "field": "value" }',
        meta: "update operator"
    },
    {
        caption: "$setOnInsert",
        value: '$setOnInsert: { "field": "value" }',
        meta: "update operator"
    },
    {
        caption: "$unset",
        value: '$unset: { "field": "" }',
        meta: "update operator"
    },
    {
        caption: "$addToSet",
        value: '$addToSet: "expression"',
        meta: "update operator"
    },
    {
        caption: "$pop",
        value: '$pop: "expression"',
        meta: "update operator"
    },
    {
        caption: "$pull",
        value: '$pull: "expression"',
        meta: "update operator"
    },
    {
        caption: "$push",
        value: '$push: "expression"',
        meta: "update operator"
    },
    {
        caption: "$pullAll",
        value: '$pullAll: { "field": ["value1", "value2"] }',
        meta: "update operator"
    },
    {
        caption: "$each",
        value: '$each: ["value1", "value2"]',
        meta: "update operator"
    },
    {
        caption: "$position",
        value: "$position: 0",
        meta: "update operator"
    },
    {
        caption: "$slice",
        value: "$slice: 2",
        meta: "update operator"
    },
    {
        caption: "$sort",
        value: '$sort: "expression"',
        meta: "update operator"
    },
    {
        caption: "$bit",
        value: '$bit: { "field": { "and|or|xor": 4} }',
        meta: "update operator"
    }
]

var configWordCompleter = {
    getCompletions: function (editor, session, pos, prefix, callback) {

        var token = session.getTokenAt(pos.row, pos.column)

        callback(null, basicBsonSnippet.map(function (snippet) {
            return {
                caption: snippet.caption,
                value: snippet.value,
                meta: snippet.meta,
                completer: {
                    insertMatch: function (editor, data) {

                        editor.removeWordLeft()

                        var start = ""
                        if (!token.value.startsWith("\"")) {
                            start = "\""
                        }

                        if (token.value.endsWith("\"")) {
                            editor.removeWordRight()
                        }

                        editor.insert(start + data.value.replace(":", "\":"))
                    }
                }
            }
        }))
    }
}

var queryWordCompleter = {

    getCompletions: function (editor, session, pos, prefix, callback) {

        var token = session.getTokenAt(pos.row, pos.column)

        var tokens = session.getTokens(pos.row)
        if (tokens.length > 3 && tokens[0].value === "db" && tokens[token.index - 1].value === ".") {
            callback(null, methodSnippet)
            return
        } else if (tokens.length === 3 && tokens[0].value === "db" && tokens[token.index - 1].value === ".") {
            callback(null, availableCollections)
            return
        }

        var wordsQuery = basicBsonSnippet

        if (editor.getSession().getLine(0).includes(".find(")) {
            wordsQuery = wordsQuery.concat(querySnippet)
        } else if (editor.getSession().getLine(0).includes(".aggregate(")) {
            wordsQuery = wordsQuery.concat(aggregationSnippet)
        } else {
            wordsQuery = wordsQuery.concat(updateSnippet)
        }

        callback(null, wordsQuery.map(function (snippet) {
            return {
                caption: snippet.caption,
                value: snippet.value,
                meta: snippet.meta,
                completer: {
                    insertMatch: function (editor, data) {

                        editor.removeWordLeft()

                        var start = ""
                        if (!token.value.startsWith("\"")) {
                            start = "\""
                        }

                        if (token.value.endsWith("\"")) {
                            editor.removeWordRight()
                        }

                        editor.insert(start + data.value.replace(":", "\":"))
                    }
                }
            }
        }))
    }
}

function addCollectionSnippet(collectionName) {
    availableCollections.push({
        caption: collectionName,
        value: collectionName,
        meta: "collection name"
    })
}

/**
 * adapatation from https://github.com/zoltantothcom/vanilla-js-dropdown by Zoltan Toth
 *
 * @class
 * @param {(string)} options.elem - HTML id of the select.
 */
var CustomSelect = function (options) {

    var elem = document.getElementById(options.elem)
    var mainClass = 'js-Dropdown'
    var titleClass = 'js-Dropdown-title'
    var listClass = 'js-Dropdown-list'
    var selectedClass = 'is-selected'
    var openClass = 'is-open'
    var selectOptions = elem.options
    var optionsLength = selectOptions.length
    var index = 0

    var selectContainer = document.createElement('div')
    selectContainer.className = mainClass
    selectContainer.id = 'custom-' + elem.id

    var button = document.createElement('button')
    button.className = titleClass
    button.textContent = selectOptions[0].textContent

    var ul = document.createElement('ul')
    ul.className = listClass

    generateOptions(selectOptions)

    selectContainer.appendChild(button)
    selectContainer.appendChild(ul)
    selectContainer.addEventListener('click', onClick)

    // pseudo-select is ready - append it and hide the original
    elem.parentNode.insertBefore(selectContainer, elem)
    elem.style.display = 'none'

    function generateOptions(options) {
        for (var i = 0; i < options.length; i++) {

            var li = document.createElement('li')
            li.innerText = options[i].textContent
            li.setAttribute('data-value', options[i].value)
            li.setAttribute('data-index', index++)

            if (selectOptions[elem.selectedIndex].textContent === options[i].textContent) {
                li.classList.add(selectedClass)
                button.textContent = options[i].textContent
            }
            ul.appendChild(li)
        }
    }

    document.addEventListener('click', function (e) {
        if (!selectContainer.contains(e.target)) {
            close()
        }
    })

    function onClick(e) {
        e.preventDefault()

        var t = e.target

        if (t.className === titleClass) {
            toggle()
        }

        if (t.tagName === 'LI') {

            selectContainer.querySelector('.' + titleClass).innerText = t.innerText
            elem.options.selectedIndex = t.getAttribute('data-index')
            elem.dispatchEvent(new CustomEvent('change'))

            var liElem = ul.querySelectorAll('li')
            for (var i = 0; i < optionsLength; i++) {
                liElem[i].classList.remove(selectedClass)
            }
            t.classList.add(selectedClass)
            close()
        }
    }

    function toggle() {
        ul.classList.toggle(openClass)
    }

    function open() {
        ul.classList.add(openClass)
    }

    function close() {
        ul.classList.remove(openClass)
    }

    function getValue() {
        return selectContainer.querySelector('.' + titleClass).innerText
    }

    function setValue(value) {
        var liElem = ul.querySelectorAll('li')
        for (var i = 0; i < optionsLength; i++) {
            liElem[i].classList.remove(selectedClass)
            if (liElem[i].innerText == value) {
                liElem[i].classList.add(selectedClass)
            }
        }
        selectContainer.querySelector('.' + titleClass).innerText = value
    }

    return {
        toggle: toggle,
        close: close,
        open: open,
        getValue: getValue,
        setValue: setValue
    }
}

var Parser = function () {

    var at,     // The index of the current character
        ch,     // The current character

        doIndent = false,
        keepComment = true,

        depth,
        needNewLine = false,

        inParenthesis,
        inNewDate,

        input, // the string to parse
        output // formatted result

    function indent(src, type, mode) {
        doIndent = true
        keepComment = true
        parse(src, type, mode)
        return output
    }

    function compact(src, type, mode) {
        doIndent = false
        keepComment = true
        parse(src, type, mode)
        return output
    }

    function compactAndRemoveComment(src, type, mode) {
        doIndent = false
        keepComment = false
        parse(src, type, mode)
        return output
    }

    function parse(src, type, mode) {

        input = src
        output = ""
        at = 0
        ch = " "
        depth = 0
        inParenthesis = false
        inNewDate = false
        needNewLine = false

        try {
            switch (type) {
                case "config":
                    config(mode)
                    break;
                case "query":
                    query()
                    break;
                default:
                    white()
                    value()
            }
            white()
            if (ch) {
                if (ch === ";") {
                    output = output.slice(0, -1)
                    white()
                    return null
                }
                error("Unexpected remaining char after end of " + type)
            }
        } catch (err) {
            // if there's an error, keep indenting so it's easyer to
            // see where the error is
            while(ch) {
                next()
            }
            return err
        }
        return null
    }

    function next(c) {

        if (c && c !== ch) {
            error("Expected '" + c + "' instead of '" + ch + "'")
        }
        nextNoAppend()

        if (inNewDate) {
            output += ch
            return
        }

        if (ch > " ") {

            if (needNewLine && ch !== "]" && ch !== "}") {
                needNewLine = false
                depth++
                if (doIndent) {
                    output += newline()
                }
            }
            switch (ch) {
                case "{":
                case "[":
                    needNewLine = true
                    output += ch
                    break
                case ",":
                    output += ch
                    if (doIndent) {
                        if (inParenthesis) {
                            output += " "
                        } else {
                            output += newline()
                        }
                    }
                    break
                case ":":
                    output += ch
                    if (doIndent) {
                        output += " "
                    }
                    break
                case "}":
                case "]":
                    if (needNewLine) {
                        needNewLine = false
                    } else {
                        depth--
                        if (doIndent) {
                            output += newline()
                        }
                    }
                    output += ch
                    break
                default:
                    output += ch
            }
        }
    }

    function nextNoAppend() {
        ch = input.charAt(at)
        at += 1
    }

    function newline() {
        // might happen with some pathological input
        if (depth < 0) {
            return "\n"
        }
        return "\n" + "  ".repeat(depth)
    }

    function white() {
        while (ch && ch <= " ") {
            next()
        }
        if (ch === "/") {
            next()

            if (ch !== "/" && ch !== "*") {
                error('Javascript regex are not supported. Use "$regex" instead')
            }

            switch (ch) {
                case "/":
                    singleLineComment()
                    break
                case "*":
                    multiligneComment()
                    break
            }
            white()
        }
    }

    function singleLineComment() {

        output = output.slice(0, -2)

        var endIndex = input.indexOf("\n", at)
        if (endIndex === -1) {
            endIndex = input.length
        }

        var comment = input.substring(at, endIndex).trimRight()

        if (keepComment) {
            if (doIndent) {
                output += "//" + comment + newline()
            } else {
                if (output.slice(-2) === "*/") {
                    output = output.slice(0, -2)
                    output += "*" + comment + "*/"
                } else {
                    output += "/**" + comment + "*/"
                }
            }
        }

        ch = input.charAt(endIndex + 1)
        at = endIndex + 2

        if (ch > " ") {
            output += ch
        }
    }

    function multiligneComment() {

        output = output.slice(0, -2)

        var endIndex = input.indexOf("*/", at)
        if (endIndex === -1) {
            error("Unfinished multiligne comment")
        }

        nextNoAppend()
        if (ch === "*") {
            nextNoAppend()
        }

        var comment = input.substring(at - 1, endIndex)

        if (keepComment && comment !== "") {
            if (doIndent) {
                // if we're here, comment is a expected to be like /**[ first line* second line* third line]*/
                // has to be transformed into this:
                //
                // // first line
                // // second line
                // // third line
                //
                //
                comment = comment.replace(/\*/gm, newline() + "//")
                output += "//" + comment + newline()
            } else {
                output += "/**" + comment + "*/"
            }
        }

        ch = input.charAt(endIndex + 2)
        at = endIndex + 3

        if (ch > " ") {
            output += ch
        }
    }

    function anyWord() {

        if (ch === '"' || ch === "'") {
            return string()
        }
        var start = at - 1
        while (ch && ((ch >= "0" && ch <= "9") || (ch >= "a" && ch <= "z") || (ch >= "A" && ch <= "Z") || ch === "$" || ch === "_")) {
            next()
        }
        return input.substring(start, at-1)
    }

    function config(mode) {

        availableCollections = []
        white()
        if (mode === "mgodatagen" && ch !== "[") {
            error("mgodatagen config has to be an array")
        }

        if (ch === "[") {
            if (mode === "bson") {
                addCollectionSnippet("collection")
                return array()
            }
            next()
            white()
            while (ch) {
                object(true)
                white()
                if (ch === "]") {
                    return next()
                }
                if (ch === ",") {
                    next()
                    white()
                    continue
                }
                error("Invalid configuration")
            }
        }

        next("d")
        next("b")
        white()
        next("=")
        white()
        if (ch === "{") {
            next()
            while (ch) {
                collectionBson()
                white()
                if (ch === "}") {
                    return next()
                }
                if (ch !== ",") {
                    error("Invalid configuration")
                }
                next()
                white()
                if (ch === "}") {
                    // trailing comma are allowed
                    return next()
                }
            }
        }
        error("Invalid configuration:\n\nmust be an array of documents like '[ {_id: 1}, {_id: 2} ]'\n\nor\n\nmust match 'db = { collection: [ {_id: 1}, {_id: 2} ] }'")
    }

    function collectionBson() {
        white()
        var collName = anyWord()
        white()
        next(":")
        white()
        array()
        addCollectionSnippet(collName)
    }

    function query() {

        white()
        next("d")
        next("b")
        next(".")
        anyWord()
        method()
        if (ch === ".") {
            return method()
        }
    }

    function number() {

        var numberStr = ""

        if (ch === "-") {
            numberStr += ch
            next()
        }
        while (ch >= "0" && ch <= "9") {
            numberStr += ch
            next()
        }
        if (ch === ".") {
            numberStr += ch
            next()
            while (ch >= "0" && ch <= "9") {
                numberStr += ch
                next()
            }
        }
        if (ch === "e" || ch === "E") {
            numberStr += ch
            next()
            if (ch === "-" || ch === "+") {
                numberStr += ch
                next()
            }
            while (ch >= "0" && ch <= "9") {
                numberStr += ch
                next()
            }
        }
        // +{string} convert a string into a number in js: wtf
        if (isNaN(+numberStr)) {
            error("Invalid number")
        }
    }

    function string() {

        if (ch !== '"' && ch !== "'") {
            error("Expected a string")
        }

        output = output.slice(0, -1)

        var string = "",
            startStringCh = ch

        nextNoAppend()

        var prevCh = ch
        while (ch) {

            if (ch === startStringCh && prevCh !== "\\") {
                break
            }

            string += ch
            prevCh = ch
            if (ch === "\n" || ch === "\r") {
                error("Invalid string: missing terminating quote")
            }
            nextNoAppend()
        }

        if (!ch) {
            output += '"' + string
            error("Invalid string: missing terminating quote")
        }

        output += '"' + string + '"'
        next()

        return string
    }

    function word() {

        var start = at - 1
        switch (ch) {
            case "t":
                next()
                next("r")
                next("u")
                return next("e")
            case "f":
                next()
                next("a")
                next("l")
                next("s")
                return next("e")
            case "n":
                next()
                switch (ch) {
                    case "u":
                        next()
                        next("l")
                        return next("l")
                    case "e":
                        return newDate()
                }
                break;
            case "u":
                next()
                next("n")
                next("d")
                next("e")
                next("f")
                next("i")
                next("n")
                next("e")
                return next("d")
            case "O":
                return objectId()
            case "I":
                return isodate()
            case "T":
                return timestamp()
            case "B":
                return binaryData()
            case "N":
                next()
                next("u")
                next("m")
                next("b")
                next("e")
                next("r")
                switch (ch) {
                    case "D":
                        return decimal128()
                    case "L":
                        return numberLong()
                    case "I":
                        return numberInt()
                }
                error("Expecting NumberInt, NumberLong or NumberDecimal")
        }

        var end = input.indexOf("\n", start)
        error("Unknown type: '" + input.substring(start, end) + "'")
    }

    function newDate() {
        inNewDate = true
        next("e")
        next("w")
        next(" ")
        next("D")
        next("a")
        next("t")
        next("e")
        inNewDate = false
        next("(")
        white()

        switch (ch) {
            case ")":
                return next()
            case '"':
            case "'":
                string()
                break
            default:
                number()
        }
        white()
        next(")")
    }

    function objectId() {

        next("O")
        next("b")
        next("j")
        next("e")
        next("c")
        next("t")
        next("I")
        next("d")
        next("(")
        white()
        var hash = string()
        if (hash.length !== 24) {
            error("Invalid ObjectId: hash has to be 24 char long")
        }
        white()
        next(")")
    }

    function isodate() {

        next("I")
        next("S")
        next("O")
        next("D")
        next("a")
        next("t")
        next("e")
        next("(")
        white()
        string()
        white()
        next(")")
    }

    function timestamp() {
        next("T")
        next("i")
        next("m")
        next("e")
        next("s")
        next("t")
        next("a")
        next("m")
        next("p")
        next("(")
        inParenthesis = true
        white()
        if (ch === ")" || ch === ",") {
            error("Invalid timestamp: missing second since unix epoch (number)")
        }
        number()
        white()
        next(",")
        white()
        if (ch === ")") {
            error("Invalid timestamp: Missing incremental ordinal (number)")
        }
        number()
        white()
        inParenthesis = false
        next(")")
    }

    function binaryData() {
        next("B")
        next("i")
        next("n")
        next("D")
        next("a")
        next("t")
        next("a")
        next("(")
        inParenthesis = true
        white()
        if (ch === ")" || ch === ",") {
            error("Missing binary type (number)")
        }
        number()
        white()
        next(",")
        white()
        string()
        white()
        inParenthesis = false
        next(")")
    }

    function decimal128() {
        next("D")
        next("e")
        next("c")
        next("i")
        next("m")
        next("a")
        next("l")
        next("(")
        white()
        if (ch === '"' || ch === "'") {
            string()
        } else {
            number()
        }
        white()
        next(")")
    }

    function numberInt() {
        next("I")
        next("n")
        next("t")
        next("(")
        white()
        if (ch === ")") {
            error("NumberInt can't be empty")
        }
        number()
        white()
        next(")")
    }

    function numberLong() {
        next("L")
        next("o")
        next("n")
        next("g")
        next("(")
        white()
        switch (ch) {
            case '"':
            case "'":
                string()
                break
            default:
                ch >= "0" && ch <= "9" ? number() : error("NumberLong() can't be empty")
        }
        white()
        next(")")
    }

    function array() {

        if (ch !== "[") {
            error("Expected an array")
        }
        next()
        white()
        if (ch === "]") {
            return next()
        }
        while (ch) {
            value()
            white()
            if (ch === "]") {
                return next()
            }
            if (ch !== ",") {
                error("Invalid array: missing closing bracket")
            }
            next()
            white()
            if (ch === "]") {
                // trailing comma are allowed
                return next()
            }
        }
        error("Invalid array: missing closing bracket")
    }

    function object(updateAvailableCollection) {

        if (ch !== "{") {
            error("Expected an object")
        }
        next()
        white()

        var keys = []

        if (ch === "}") {
            return next()
        }
        while (ch) {

            var key = anyWord()
            white()
            next(":")
            if (keys.includes(key)) {
                error("Duplicate key '" + key + "'")
            }
            keys.push(key)
            var val = value()
            if (updateAvailableCollection && key === "collection") {
                addCollectionSnippet(val)
            }
            white()
            if (ch === "}") {
                return next()
            }
            if (ch !== ",") {
                error("Invalid object: missing closing bracket")
            }
            next()
            white()
            if (ch === "}") {
                // trailing comma are allowed
                return next()
            }
        }
        error("Invalid object: missing closing bracket")
    }

    function value() {

        white()
        switch (ch) {
            case "{":
                return object()
            case "[":
                return array()
            case '"':
            case "'":
                return string()
            case "-":
                return number()
            default:
                ch >= '0' && ch <= '9' ? number() : word()
        }
    }

    function explain() {
        next("e")
        next("x")
        next("p")
        next("l")
        next("a")
        next("i")
        next("n")
        next("(")
        white()
        if (ch === ")") {
            return next()
        }
        var explainMode = string()
        if (!["executionStats", "queryPlanner", "allPlansExecution"].includes(explainMode)) {
            error("Invalid explain mode :" + explainMode + ', expected one of ["executionStats", "queryPlanner", "allPlansExecution"] ')
        }
        white()
        next(")")
    }

    function find() {
        next("f")
        next("i")
        next("n")
        next("d")
        next("(")
        white()
        nObject(2)
        white()
        next(")")
    }

    function aggregate() {
        next("a")
        next("g")
        next("g")
        next("r")
        next("e")
        next("g")
        next("a")
        next("t")
        next("e")
        next("(")
        white()
        switch (ch) {
            case "[":
                array()
                break
            case "{" :
                nObject(-1)
                break
        }
        white()
        next(")")
    }

    function nObject(n) {
        var count = 0
        while (ch && ch === "{") {
            count++
            if (n !== -1 && count > n) {
                error("too many object, expected up to " + n)
            }
            object()
            white()
            if (ch === ",") {
                next()
                white()
            }
        }
    }

    function update() {
        next("u")
        next("p")
        next("d")
        next("a")
        next("t")
        next("e")
        next("(")
        white()
        object()
        white()
        next(",")
        white()
        if (ch === "[" ) {
            array()
        } else {
            object()
        }
        white()
        if (ch === ",") {
            next()
            if (ch === ")") {
                return next()
            }
            white()
            object()
            white()
        }
        if (ch === ",") {
            next()
            white()
        }
        next(")")
    }

    function method() {
        next(".")
        switch (ch) {
            case "f":
                return find()
            case "a":
                return aggregate()
            case "u":
                return update()
            case "e":
                return explain()
            default:
                error("Unsupported method: only find(), aggregate(), update() and explain() are supported")
        }
    }

    function error(m) {
        throw {
            message: m,
            at: at
        }
    }

    return {
        indent: indent,
        compact: compact,
        compactAndRemoveComment: compactAndRemoveComment,
        parse: parse
    }
}