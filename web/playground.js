var doRedirect = true

function redirect() {
    clearLines()
    if (doRedirect) {
        window.history.replaceState({}, "MongoDB playground", "/")
        document.getElementById("link").style.visibility = "hidden"
        document.getElementById("link").value = ""
        doRedirect = false
        document.getElementById("share").disabled = false
    }
}

function loadDocs() {
    var r = new XMLHttpRequest()
    r.open("GET", "/static/docs.html", true)
    r.onreadystatechange = function () {
        if (r.readyState !== 4) {
            return
        }
        if (r.status === 200) {
            document.getElementById("docContent").innerHTML = r.responseText
        }
    }
    r.send(null)
}

function showDoc(doShow) {
    if (doShow && document.getElementById("docDiv").style.display === "inline") {
        doShow = false
    }
    document.getElementById("docDiv").style.display = doShow ? "inline" : "none"
    document.getElementById("queryDiv").style.display = doShow ? "none" : "inline"
    document.getElementById("resultDiv").style.display = doShow ? "none" : "inline"
}

function run(doSave) {
    if (isCorrect()) {
        var config = document.getElementById("config").value
        var query = document.getElementById("query").value
        var r = new XMLHttpRequest()
        r.open("POST", doSave ? "/save/" : "/run/")
        r.setRequestHeader("Content-Type", "application/x-www-form-urlencoded")
        r.onreadystatechange = function () {
            if (r.readyState !== 4) {
                return
            }
            if (r.status === 200) {
                if (doSave) {
                    window.history.replaceState({}, "MongoDB playground", r.responseText)
                    doRedirect = true
                    var link = document.getElementById("link")
                    link.value = r.responseText
                    link.style.visibility = "visible"
                    document.getElementById("share").disabled = true
                } else {
                    var resultArea = document.getElementById("result")
                    var resultNb = document.getElementById("resultNb")
                    try {
                        var results = JSON.parse(r.responseText)
                        resultArea.classList.remove("red")
                        resultNb.innerHTML = results.length + " results"
                        resultArea.value = JSON.stringify(results, null, 2)
                    } catch (e) {
                        error(r.responseText, "result")
                    }
                }
            }
        }
        config = JSON.stringify(JSON.parse(config))
        var select = document.getElementById("mode")
        r.send("mode=" + select.options[select.selectedIndex].text + "&config=" + encodeURIComponent(config) + "&query=" + encodeURIComponent(query))
    }
}

function isCorrect() {

    showDoc(false)

    var config = document.getElementById("config")
    var content = format(config.value.trim(), "config")
    if (content !== "invalid") {
        config.value = content
    } else {
        return false
    }

    var query = document.getElementById("query")
    content = query.value.trim()
    if (content.slice(-1) === ";") {
        content = content.substring(0, content.length - 1)
    }
    var match = /^db\..*\.(find|aggregate)\([\s\S]*\)$/.test(content)
    if (!match) {
        error("invalid query: \nmust match db.coll.find(...) or db.coll.aggregate(...)", "query")
        return false
    }
    var queryStart = content.substring(0, content.indexOf("(") + 1)
    var queryEnd = content.substring(content.length - 1, content.length)
    content = content.substring(content.indexOf("(") + 1, content.length - 1).trim()
    content = format(content, "query")
    if (content !== "invalid") {
        query.value = queryStart + content + queryEnd
        clearLines()
        return true
    } else {
        return false
    }
}

function format(content, textArea) {
    if (content === "") {
        return ""
    }
    try {
        var obj = JSON.parse(content)
        return JSON.stringify(obj, null, 2)
    } catch (e) {
        error("invalid " + textArea + ":\n" + e, textArea)
        return "invalid"
    }
}

function error(errorMsg, textArea) {
    var line = errorMsg.match(/line [0-9]*/)
    if (line !== null && textArea !== "result") {
        var nb = Number(line[0].replace("line ", "")) - 1
        var lineDiv = document.getElementById(textArea + "Lines")
        lineDiv.childNodes[nb].classList.add("red")
    }
    var resultArea = document.getElementById("result")
    resultArea.classList.add("red")
    resultArea.value = errorMsg
    document.getElementById("resultNb").innerHTML = "0 result"
}

function clearLines() {
    document.querySelectorAll(".lines > div").forEach(function(element) {
        element.classList.remove("red")
    })
}

function scrollArea(textAreaId) {
    var textArea = document.getElementById(textAreaId)
    var scrollTop = textArea.scrollTop
    var clientHeight = textArea.clientHeight

    var linesDiv = textArea.parentNode.querySelector(".lines")
    linesDiv.style.marginTop = (-scrollTop) + "px"
    var lineNo = textArea.getAttribute("data-lineNo")
    lineNo = fillOutLines(linesDiv, scrollTop + clientHeight, lineNo)
    textArea.setAttribute("data-lineNo", lineNo)
}

function fillOutLines(linesDiv, h, lineNo) {
    while (linesDiv.clientHeight < h) {
        var divNo = document.createElement('div')
        divNo.innerHTML = lineNo
        linesDiv.appendChild(divNo)
        lineNo++
    }
    return lineNo
}