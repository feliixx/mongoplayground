var compactMode = 0
var indentMode = 1

function indent(src, mode) {
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
            result += (mode === indentMode) ? newline(depth) : ""
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
                if (mode === indentMode) {
                    if (inParenthesis) {
                        result += " "
                    } else {
                        result += newline(depth)
                    }
                }
                break
            case ":":
                result += c
                if (mode === indentMode) {
                    result += " "
                }
                break
            case "}":
            case "]":
                if (needIndent) {
                    needIndent = false
                } else {
                    depth--
                    result += (mode === indentMode) ? newline(depth) : ""
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

function formatQuery(content, mode) {
    var result = content
    if (content.endsWith(";")) {
        result = content.slice(0, -1)
    }
    var correctQuery = /^db\..(\w+)\.(find|aggregate)\([\s\S]*\)$/.test(result)
    if (!correctQuery) {
        return "invalid"
    }
    if (mode === "json") {
        var parts = result.split(".")
        parts[1] = "collection"
        result = parts.join(".")
    }

    var start = result.indexOf("(") +1
    query = result.substring(start, result.length-1)
    if (query !== "" && !query.endsWith("}") && !query.endsWith("]")) {
        return "invalid"
    }
    return result
}
