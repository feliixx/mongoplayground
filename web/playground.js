function indent(src) {
    return format(src, true, true)
}

function compact(src) {
    return format(src, false, true)
}

function compactAndRemoveComment(src) {
    return format(src, false, false)
}

function format(src, indent, keepComment) {

    if (src.endsWith(";")) {
        src = src.slice(0, -1)
    }

    var result = ""
    var needIndent = false
    var inParenthesis = false
    var depth = 0
    var i = 0

    if (src.startsWith("db.")) {
        i = src.indexOf("(") + 1
        result += src.substring(0, i)
    }
    while (i < src.length) {
        var c = src.charAt(i)
        if (c === " " || c === "\n" || c === "\t") {
            i++
            continue
        }
        if (needIndent && c !== "]" && c !== "}") {
            needIndent = false
            depth++
            result += indent ? newline(depth) : ""
        }

        switch (c) {
            case "(":
                inParenthesis = true
                result += c
                break
            case ")":
                inParenthesis = false
                result += c
                break
            case "{":
            case "[":
                needIndent = true
                result += c
                break
            case ",":
                result += c
                if (indent) {
                    if (inParenthesis) {
                        result += " "
                    } else {
                        result += newline(depth)
                    }
                }
                break
            case ":":
                result += c
                if (indent) {
                    result += " "
                }
                break
            case "}":
            case "]":
                if (needIndent) {
                    needIndent = false
                } else {
                    depth--
                    result += indent ? newline(depth) : ""
                }
                result += c
                break
            case "\"":
            case "'":
                var end = c
                result += "\""
                i++
                c = src.charAt(i)
                while (c !== end && i < src.length) {
                    result += c
                    i++
                    c = src.charAt(i)
                }
                if (i != src.length) {
                    result += "\""
                }
                break
            case "n":
                var tmp = src.substring(i, i + 9)
                if (tmp === "new Date(") {
                    result += tmp
                    i += tmp.length
                    c = src.charAt(i)
                    while (c !== ")" && i < src.length) {
                        result += c
                        i++
                        c = src.charAt(i)
                    }
                    if (i != src.length) {
                        result += ")"
                    }
                } else {
                    result += c
                }
                break
            case "/":

                // single ligne comment, starting with '//' 
                if (src.length >= i + 1 && src.charAt(i + 1) == "/") {

                    if (!keepComment) {
                        i += 2
                        while (c != "\n" && i < src.length) {
                            i++
                            c = src.charAt(i)
                        }
                        continue
                    }

                    // rewrite every line with /**...*/ 
                    result += "/**"
                    i+=2
                    c = src.charAt(i)

                    while (c != "\n" && i < src.length) {
                        result += c
                        i++
                        c = src.charAt(i)
                    }
                    result += "*/"
                    if (indent) {
                        result += newline(depth)
                    }
                // multi ligne comment type javadoc: /**...*/
                } else if (src.length >= i + 2 && src.charAt(i + 1) == "*" && src.charAt(i + 2) == "*") {

                    start = i + 3
                    i = src.indexOf("*/", start)

                    if (!keepComment) {
                        if (i == -1) {
                            i = start
                            if (src.charAt(i) == "/") {
                                i++
                            }
                        } else {
                            i+=2
                        }
                        continue
                    }

                    if (i == -1) {
                        i = start
                        if (src.charAt(i) == "/") {
                            i++
                        }
                        continue
                    }
                    comment = src.substring(start, i)

                    result += "/**"
                    // each '*' in the body of the comment means newline
                    if (!indent) {
                        result += comment.replace(/[\s]+\*/gm, "*").trimRight()
                    } else {
                        comment = comment.replace(/[\s]+\*/gm, "*").trimRight()
                        comment = comment.replace(/\*/gm, newline(depth) + "*")

                        if (comment.indexOf("*") > 0) {
                            comment += newline(depth)
                        }
                        result += comment
                    }

                    result += "*/"

                    i++
                    if (indent) {
                        result += newline(depth)
                    }
                // multiligne comment classic: /*...*/
                } else if (src.length >= i + 1 && src.charAt(i + 1) == "*") {

                    start = i + 2
                    i = src.indexOf("*/", start)

                    if (!keepComment) {
                        if (i == -1) {
                            i = start
                        } else {
                            i+=2
                        }
                        continue
                    }
                    if (i == -1) {
                        i = start
                        continue
                    }
                    comment = src.substring(start, i)

                    // rewrite the whole as /**...*/, and add a '*' at the start
                    // of every new line
                    result += "/**"

                    if (!indent) {
                        result += comment.replace(/[\s]*\n[\s]+/gm, "* ").trimRight()
                    } else {
                        comment = comment.replace(/[\s]*\n[\s]+/gm, "* ").trimRight()
                        comment = comment.replace(/\*/gm, newline(depth) + "*")

                        if (comment.indexOf("*") > 0) {
                            comment += newline(depth)
                        }
                        result += comment
                    }

                    result += "*/"

                    i++
                    if (indent) {
                        result += newline(depth)
                    }
                } else {
                    result += c
                    i++
                    c = src.charAt(i)
                    while (c !== "/" && i < src.length) {
                        result += c
                        i++
                        c = src.charAt(i)
                    }
                    if (i != src.length) {
                        result += "/"
                    }
                }
                break
            default:
                result += c
        }
        i++
    }
    return result
}

function newline(depth) {
    var line = "\n"
    for (var i = 0; i < depth; i++) {
        line += "  "
    }
    return line
}

function isConfigValid(content, mode) {

    var configWithoutComment = compactAndRemoveComment(content)

    // mgodatagen and bson single collection have an array as config
    if (!configWithoutComment.startsWith("[") || !configWithoutComment.endsWith("]")) {
        if (mode === "bson") {
            // check wether it match the multiple collection config, ie `db = {...}`
            return /^\s*db\s*=\s*\{[\s\S]*\}$/.test(configWithoutComment)
        } else {
            return false
        }
    }
    return true
}

function isQueryValid(content) {
    if (content.endsWith(";")) {
        content = content.slice(0, -1)
    }
    var queryWithoutComment = compactAndRemoveComment(content)

    var correctQuery = /^db\..(\w*)\.(find|aggregate)\([\s\S]*\)$/.test(queryWithoutComment)
    if (!correctQuery) {
        return false
    }

    var start = queryWithoutComment.indexOf("(") + 1
    query = queryWithoutComment.substring(start, queryWithoutComment.length - 1)
    if (query !== "" && !query.endsWith("}") && !query.endsWith("]")) {
        return false
    }
    return true
}
