function compact(src) {
    return format(src, false, true)
}

function compactAndRemoveComment(src) {
    return format(src, false, false)
}

function indent(src) {
    return format(src, true, true)
}

function format(src, indent, keepComment) {
    var result = ""
    var needIndent = false
    var inParenthesis = false
    var depth = 0
    var i = 0
    var c = src.charAt(i)
    if (src.startsWith("db.")) {
        i = src.indexOf("(") + 1
        result += src.substring(0, i)
    }
    while (i < src.length) {
        c = src.charAt(i)
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

                if (src.length >= i + 1 && src.charAt(i + 1) == "/") {

                    // single ligne comment
                    if (!keepComment) {
                        i += 2
                        while (c != "\n" && i < src.length) {
                            i++
                            c = src.charAt(i)
                        }
                        continue
                    }
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
                } else if (src.length >= i + 2 && src.charAt(i + 1) == "*" && src.charAt(i + 2) == "*") {
                    // multi ligne comment

                    start = i + 3
                    i = src.indexOf("*/", start)

                    if (!keepComment) {
                        i+=2
                        c = src.charAt(i)
                        continue
                    }

                    if (i == -1) {
                        i = start
                        continue
                    }
                    comment = src.substring(start, i)

                    result += "/**"

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
                    c = src.charAt(i)
                    if (indent) {
                        result += newline(depth)
                    }
                } else if (src.length >= i + 1 && src.charAt(i + 1) == "*") {

                    start = i + 2
                    i = src.indexOf("*/", start)

                    if (!keepComment) {
                        i+=2
                        c = src.charAt(i)
                        continue
                    }
                    if (i == -1) {
                        i = start
                        continue
                    }
                    comment = src.substring(start, i)

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
                    c = src.charAt(i)
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

function formatConfig(content, mode) {

    var contentWithoutComment = compactAndRemoveComment(content)
    if (!contentWithoutComment.startsWith("[") || !contentWithoutComment.endsWith("]")) {
        if (mode === "bson") {
            var correctConfigMultipleCollections = /^\s*db\s*=\s*\{[\s\S]*\}$/.test(contentWithoutComment)
            if (!correctConfigMultipleCollections) {
                return "invalid"
            }
        } else {
            return "invalid"
        }
    }
    return content
}

function formatQuery(content) {
    var result = content
    if (content.endsWith(";")) {
        result = content.slice(0, -1)
    }
    var queryWithoutComment = compactAndRemoveComment(result)

    var correctQuery = /^db\..(\w*)\.(find|aggregate)\([\s\S]*\)$/.test(queryWithoutComment)
    if (!correctQuery) {
        return "invalid"
    }

    var start = queryWithoutComment.indexOf("(") + 1
    query = queryWithoutComment.substring(start, queryWithoutComment.length - 1)
    if (query !== "" && !query.endsWith("}") && !query.endsWith("]")) {
        return "invalid"
    }
    return result
}
