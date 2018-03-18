var configArea
var queryArea
var resultArea

function initCodeArea() {
    configArea = ace.edit(document.getElementById("config"), {
        "mode": "ace/mode/json",
        "fontSize": "16px"
    })
    queryArea = ace.edit(document.getElementById("query"), {
        "mode": "ace/mode/javascript",
        "fontSize": "16px"
    })
    resultArea = ace.edit(document.getElementById("result"), {
        "mode": "ace/mode/json",
        "fontSize": "16px",
        "readOnly": true,
        "showLineNumbers": false,
        "showGutter": false,
        "useWorker": false,
        "highlightActiveLine": false
    })
    configArea.setValue(formatConfig(2), -1)

    document.getElementById("config").removeAttribute("hidden")
    document.getElementById("query").removeAttribute("hidden")
    document.getElementById("result").removeAttribute("hidden")

    configArea.getSession().on('change', function () {
        redirect()
    })
    queryArea.getSession().on('change', function () {
        redirect()
    })

    var r = new XMLHttpRequest()
    r.open("GET", "/static/docs.html", true)
    r.onreadystatechange = function () {
        if (r.readyState !== 4) { return }
        if (r.status === 200) {
            document.getElementById("docContent").innerHTML = r.responseText
        }
    }
    r.send(null)
}

function getMode() {
    var select = document.getElementById("mode")
    return select.options[select.selectedIndex].text
}

function setMode() {
    if (getMode() === "json") {
        configArea.session.setMode("ace/mode/javascript")
        document.getElementById("config").classList.add("ignoreWarnings")
    } else {
        configArea.session.setMode("ace/mode/json")
        document.getElementById("config").classList.remove("ignoreWarnings")
    }
    redirect()
}

function redirect() {
    window.history.replaceState({}, "MongoDB playground", "/")
    document.getElementById("link").style.visibility = "hidden"
    document.getElementById("link").value = ""
    document.getElementById("share").disabled = false
}

function showDoc(doShow) {
    if (doShow && document.getElementById("docDiv").style.display === "inline") {
        doShow = false
    }
    document.getElementById("docDiv").style.display = doShow ? "inline" : "none"
    document.getElementById("queryDiv").style.display = doShow ? "none" : "inline"
    document.getElementById("resultDiv").style.display = doShow ? "none" : "inline"
    if (!doShow) {
        redirect()
    }
}

function run(doSave) {
    if (isCorrect()) {
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
                    var link = document.getElementById("link")
                    link.value = r.responseText
                    link.style.visibility = "visible"
                    document.getElementById("share").disabled = true
                } else {
                    var resultNb = document.getElementById("resultNb")
                    try {
                        var results = JSON.parse(r.responseText)
                        resultNb.innerHTML = results.length + " results"
                        resultArea.setValue(JSON.stringify(results, null, 2), -1)
                    } catch (e) {
                        resultArea.setValue(r.responseText, -1)
                    }
                }
            }
        }
        r.send("mode=" + getMode() + "&config=" + encodeURIComponent(formatConfig(0)) + "&query=" + encodeURIComponent(queryArea.getValue()))
    }
}

function isCorrect() {
    showDoc(false)
    resultArea.setValue("", -1)
    document.getElementById("resultNb").innerHTML = "0 result"

    var errors = document.querySelectorAll(".ace_error")
    if (errors.length > 0) {
        resultArea.setValue("error(s) found in config or query", -1)
        return false
    }
    configArea.setValue(formatConfig(2), -1)
    var content = queryArea.getValue().trim()
    var match = /^db\..*\.(find|aggregate)\([\s\S]*\)$/.test(content)
    if (!match) {
        resultArea.setValue("invalid query: \nmust match db.coll.find(...) or db.coll.aggregate(...)", -1)
        return false
    }
    queryArea.setValue(js_beautify(queryArea.getValue(), {}), -1)
    return true
}

function formatConfig(indentSize) {
    if (getMode() === "json") {
        return js_beautify(configArea.getValue(), {
            "indent_size": indentSize,
            "indent_char": indentSize === 0 ? "" : " ",
            "unescape_strings": true,
            "preserve_newlines": false
        })
    } else {
        return JSON.stringify(JSON.parse(configArea.getValue()), null, indentSize)
    }
}