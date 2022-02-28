
var configEditor,
    queryEditor,
    resultEditor,

    comboMode,
    comboTemplate,
    comboStages,

    parser,

    hasChangedSinceLastRun = true,
    hasChangedSinceLastSave = true,
    isConfigHandlerDragging = false,
    isQueryHandlerDragging = false

window.onload = function () {

    parser = new Parser()

    comboStages = new CustomSelect({
        selectId: "aggregation_stages",
        onChange: function () { run() }
    })
    comboMode = new CustomSelect({
        selectId: "mode",
        onChange: function () { checkEditorContent(configEditor, "config") }
    })
    comboTemplate = new CustomSelect({
        selectId: "template",
        onChange: function () { setTemplate(comboTemplate.getSelectedIndex()) }
    })

    var commonOpts = {
        "mode": "ace/mode/mongo",
        "fontSize": "16px",
        "enableBasicAutocompletion": true,
        "enableLiveAutocompletion": true,
        "enableSnippets": true,
        "useWorker": false,
        "useSoftTabs": true,
        "tabSize": 2,
    }

    var configDiv = document.getElementById("config")
    var queryDiv = document.getElementById("query")

    configEditor = ace.edit(configDiv, commonOpts)
    queryEditor = ace.edit(queryDiv, commonOpts)
    resultEditor = ace.edit(document.getElementById("result"), {
        "mode": commonOpts.mode,
        "fontSize": commonOpts.fontSize,
        "readOnly": true,
        "showLineNumbers": false,
        "showGutter": false,
        "useWorker": false,
        "highlightActiveLine": false,
        "wrap": true
    })
    resultEditor.renderer.$cursorLayer.element.style.display = "none"

    configEditor.completers = [configWordCompleter]
    queryEditor.completers = [queryWordCompleter]

    configEditor.getSession().on("change", checkEditorContent.bind(null, configEditor, "config"))
    queryEditor.getSession().on("change", checkEditorContent.bind(null, queryEditor, "query"))

    configEditor.setValue(parser.indent(configEditor.getValue(), "config", comboMode.getValue()), -1)
    queryEditor.setValue(parser.indent(queryEditor.getValue(), "query", comboMode.getValue()), -1)

    configDiv.style.display = "inline"
    queryDiv.style.display = "inline"

    hasChangedSinceLastRun = false
    hasChangedSinceLastSave = false

    addKeyDownListener()
    addDivResizeListener()
    addButtonClickListener()
}

function addKeyDownListener() {
    document.addEventListener("keydown", function (event) {
        if ((event.ctrlKey || event.metaKey) && event.key === "Enter") {
            event.preventDefault()
            run()
        }
        if ((event.ctrlKey || event.metaKey) && event.key === "s") {
            event.preventDefault()
            formatAll(true)
        }
    })
}

function addDivResizeListener() {
    document.addEventListener("mousedown", function (e) {
        if (e.target.id === "configResizeHandler") {
            isConfigHandlerDragging = true
        }
        if (e.target.id === "queryResizeHandler") {
            isQueryHandlerDragging = true
        }
    })

    document.addEventListener("mouseup", function (e) {
        isConfigHandlerDragging = false
        isQueryHandlerDragging = false
    })

    document.addEventListener("mousemove", function (e) {
        var box
        if (isConfigHandlerDragging) {
            box = document.getElementById("configPanel")
        } else if (isQueryHandlerDragging) {
            box = document.getElementById("queryPanel")
        } else {
            return false
        }
        var pointerRelativeXpos = e.clientX - box.offsetLeft
        box.style.width = (Math.max(60, pointerRelativeXpos + 2)) + "px"
        box.style.flexGrow = "0"
    })
}

function addButtonClickListener() {
    document.getElementById("run").addEventListener("click", function (e) { run() })
    document.getElementById("format").addEventListener("click", function (e) { formatAll(true) })
    document.getElementById("share").addEventListener("click", function (e) { save() })
    document.getElementById("doc").addEventListener("click", function (e) { showDoc(true) })
}

function checkEditorContent(editor, type) {

    var errors = []
    var err = parser.parse(editor.getValue(), type, comboMode.getValue())
    if (err != null) {
        var pos = editor.getSession().getDocument().indexToPosition(err.at - 1)
        errors.push({
            row: pos.row,
            column: pos.column,
            text: err.message,
            type: "error"
        })
    }
    editor.getSession().setAnnotations(errors)
    if (type === "query") {
        comboStages.setOptions(parser.getAggregationStages())
    }

    if (!hasChangedSinceLastRun || !hasChangedSinceLastSave) {
        hasChangedSinceLastRun = true
        hasChangedSinceLastSave = true
        redirect("/", false)
    }
}

function redirect(url, showLink) {
    window.history.replaceState({}, "MongoDB playground", url)
    document.getElementById("link").style.visibility = showLink ? "visible" : "hidden"
    document.getElementById("link").value = url
    document.getElementById("share").disabled = showLink
}

function setTemplate(index) {
    comboMode.setValue(templates[index].mode)
    configEditor.setValue(parser.indent(templates[index].config, "config", comboMode.getValue()), 1)
    queryEditor.setValue(parser.indent(templates[index].query, "query", comboMode.getValue()), 1)
    resultEditor.setValue("", 1)
}

function showDoc(doShow) {

    if (doShow && !document.getElementById("docPanel").hasChildNodes()) {
        loadDocs()
    }

    if (doShow && document.getElementById("docPanel").style.display === "inline") {
        doShow = false
    }
    document.getElementById("docPanel").style.display = doShow ? "inline" : "none"
    document.getElementById("queryPanel").style.display = doShow ? "none" : "inline"
    document.getElementById("resultPanel").style.display = doShow ? "none" : "inline"
    if (!doShow && hasChangedSinceLastSave) {
        redirect("/", false)
    }
}

function loadDocs() {
    var r = new XMLHttpRequest()
    r.open("GET", "/static/docs-12.html", true)
    r.onreadystatechange = function () {
        if (r.readyState !== 4) { return }
        if (r.status === 200) {
            document.getElementById("docPanel").innerHTML = r.responseText
        }
    }
    r.send(null)
}

function run() {
    if (formatAll(false)) {

        showResult("running query...", false)

        var r = new XMLHttpRequest()
        r.open("POST", "/run")
        r.setRequestHeader("Content-Type", "application/x-www-form-urlencoded")
        r.onreadystatechange = function () {
            if (r.readyState !== 4) { return }
            if (r.status === 200) {

                hasChangedSinceLastRun = false
                var response = r.responseText
                if (response.startsWith("[") || response.startsWith("{")) {
                    showResult(response, true)
                } else if (response === "no document found") {
                    showResult(response, false)
                } else {
                    showError(response)
                }
            }
        }
        r.send(encodePlayground(false))
    }
}

function save() {

    formatAll(hasChangedSinceLastRun)

    var r = new XMLHttpRequest()
    r.open("POST", "/save")
    r.setRequestHeader("Content-Type", "application/x-www-form-urlencoded")
    r.onreadystatechange = function () {
        if (r.readyState !== 4) { return }
        if (r.status === 200) {

            hasChangedSinceLastSave = false
            var response = r.responseText
            if (response.startsWith("http")) {
                redirect(response, true)
            } else {
                showError(response)
            }
        }
    }
    r.send(encodePlayground(true))
}

function encodePlayground(keepComment) {
    var result = "mode=" + comboMode.getValue()
    if (keepComment) {
        result += "&config=" + encodeURIComponent(parser.compact(configEditor.getValue(), "config", comboMode.getValue()))
            + "&query=" + encodeURIComponent(parser.compact(queryEditor.getValue(), "query", comboMode.getValue()))
    } else {
        result += "&config=" + encodeURIComponent(parser.compactAndRemoveComment(configEditor.getValue(), "config", comboMode.getValue(), 0))
            + "&query=" + encodeURIComponent(parser.compactAndRemoveComment(queryEditor.getValue(), "query", comboMode.getValue(), comboStages.getSelectedIndex() + 1))
    }
    return result
}

function formatAll(clearResult) {

    if (clearResult) {
        resultEditor.setValue("", -1)
    }

    showDoc(false)

    var errors = configEditor.getSession().getAnnotations()
    if (errors.length > 0) {
        showError("Invalid configuration:\n\nLine " + (errors[0].row + 1) + ": " + errors[0].text)
        return false
    }
    errors = queryEditor.getSession().getAnnotations()
    if (errors.length > 0) {
        showError("Invalid query:\n\nLine " + (errors[0].row + 1) + ": " + errors[0].text)
        return false
    }

    if (hasChangedSinceLastRun) {
        configEditor.setValue(parser.indent(configEditor.getValue(), "config", comboMode.getValue()), 1)
        queryEditor.setValue(parser.indent(queryEditor.getValue(), "query", comboMode.getValue()), 1)
    }
    return true
}

function showError(errMsg) {
    document.getElementById("result").classList.add("text_red")
    resultEditor.setOption("wrap", true)
    resultEditor.setValue(errMsg, -1)
}

function showResult(result, doIndent) {
    document.getElementById("result").classList.remove("text_red")
    if (doIndent) {
        result = parser.indent(result, "result", comboMode.getValue())
    }
    resultEditor.setOption("wrap", false)
    resultEditor.setValue(result, -1)
}
