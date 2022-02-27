var Parser = function () {

    var at,     // The index of the current character
        ch,     // The current character

        doIndent = false,
        keepComment = true,

        depth,
        needNewLine = false,

        inParenthesis,
        inNewDate,

        input,  // the string to parse
        output, // formatted result

        aggregationStagesLimit = 0,
        aggregationStages = [] // list of stages name for aggregation queries

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

    function compactAndRemoveComment(src, type, mode, nbStagesToKeep) {
        doIndent = false
        keepComment = false
        aggregationStagesLimit = nbStagesToKeep
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

        aggregationStages = []

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
            while (ch) {
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
        return input.substring(start, at - 1)
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
                pipeline()
                break
            case "{":
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

    // a pipeline is an array of stages, which are objects 
    function pipeline() {

        if (ch !== "[") {
            error("Expected an array")
        }
        next()
        white()
        if (ch === "]") {
            return next()
        }

        var stagesNb = 0
        var indexEndLastWantedStages = output.length

        while (ch) {

            stage()
            stagesNb++

            if (stagesNb === aggregationStagesLimit) {
                indexEndLastWantedStages = output.length - 1
            }

            white()
            if (ch === "]") {
                if (aggregationStagesLimit !== 0 && stagesNb > aggregationStagesLimit) {
                    output = output.slice(0, indexEndLastWantedStages)
                    output += "]"
                }
                return next()
            }
            if (ch !== ",") {
                error("Invalid array: missing closing bracket")
            }
            next()
            white()
            if (ch === "]") {
                if (aggregationStagesLimit !== 0 && stagesNb > aggregationStagesLimit) {
                    output = output.slice(0, indexEndLastWantedStages)
                    output += "]"
                }
                // trailing comma are allowed
                return next()
            }
        }
        error("Invalid array: missing closing bracket")
    }

    function stage() {
        if (ch !== "{") {
            error("Expected an object")
        }
        next()
        white()

        var keys = []
        var stageNamePushed = false

        if (ch === "}") {
            return next()
        }
        while (ch) {

            var key = anyWord()

            if (!stageNamePushed) {
                aggregationStages.push(key)
                stageNamePushed = true
            }

            white()
            next(":")
            if (keys.includes(key)) {
                error("Duplicate key '" + key + "'")
            }
            keys.push(key)
            value()

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
        if (ch === "[") {
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

    function getAggregationStages() {
        return aggregationStages
    }

    return {
        indent: indent,
        compact: compact,
        compactAndRemoveComment: compactAndRemoveComment,
        parse: parse,
        getAggregationStages: getAggregationStages
    }
}

function addCollectionSnippet(collectionName) {
    availableCollections.push({
        caption: collectionName,
        value: collectionName,
        meta: "collection name"
    })
}