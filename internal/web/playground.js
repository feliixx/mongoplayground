/* ***** BEGIN LICENSE BLOCK *****
* mongoplayground: a sandbox to test and share MongoDB queries
* Copyright (C) 2017 Adrien Petel
*
* This program is free software: you can redistribute it and/or modify
* it under the terms of the GNU Affero General Public License as published
* by the Free Software Foundation, either version 3 of the License, or
* (at your option) any later version.
*
* This program is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
* GNU Affero General Public License for more details.
*
* You should have received a copy of the GNU Affero General Public License
* along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * ***** END LICENSE BLOCK ***** */

var configEditor,
    queryEditor,
    resultEditor,

    comboMode,
    comboTemplate,
    comboStages,

    parser = new Parser(),

    configChangedSinceLastRun = true,
    queryChangedSinceLastRun = true,
    configOrQueryChangedSinceLastSave = true,

    isConfigHandlerDragging = false,
    isQueryHandlerDragging = false

function init() {

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
        "showPrintMargin": false
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
        "wrap": true,
        "showPrintMargin": false
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

    configChangedSinceLastRun = false
    queryChangedSinceLastRun = false
    configOrQueryChangedSinceLastSave = false

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
    document.getElementById("showDoc").addEventListener("click", function (e) { showDoc(true) })
}

/**
 * Check editor content for syntax error
 * 
 * @param {ace.Editor} editor - the ace editor to check content from
 * @param {string} type - type of editor, must be one of ["config", "query"] 
 */
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
        if (parser.getQueryType() === "aggregate") {
            document.getElementById("custom-aggregation_stages").style.visibility = "visible"
            document.getElementById("aggregation_stages_label").style.visibility = "visible"
            comboStages.setOptions(parser.getAggregationStages())
        } else {
            document.getElementById("custom-aggregation_stages").style.visibility = "hidden"
            document.getElementById("aggregation_stages_label").style.visibility = "hidden"
        }
    }

    if (!configChangedSinceLastRun || !queryChangedSinceLastRun || !configOrQueryChangedSinceLastSave) {
        if (type === "query") {
            queryChangedSinceLastRun = true
        } else {
            configChangedSinceLastRun = true
        }
        configOrQueryChangedSinceLastSave = true
        redirect("/", false)
    }
}

/**
 * Change the browser url 
 * 
 * @param {string} url - the url to display in the browser 
 * @param {boolean} showLink - wether to show the playground link in the toolbar
 */
function redirect(url, showLink) {
    window.history.replaceState({}, "MongoDB playground", url)
    document.getElementById("link").style.visibility = showLink ? "visible" : "hidden"
    document.getElementById("link").value = url
    document.getElementById("share").disabled = showLink
}

/**
 * Fill config and query editor with a specific template 
 * 
 * @param {int} index - index of the template to use  
 */
function setTemplate(index) {
    comboMode.setValue(templates[index].mode)
    configEditor.setValue(parser.indent(templates[index].config, "config", comboMode.getValue()), 1)
    queryEditor.setValue(parser.indent(templates[index].query, "query", comboMode.getValue()), 1)
    resultEditor.setValue("", 1)
}

/**
 * Show or hide the documentation panel
 * 
 * @param {boolean} doShow - display the doc panel if true, otherwise hide it 
 */
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
    if (!doShow && configOrQueryChangedSinceLastSave) {
        redirect("/", false)
    }
}

/**
 * load the documentation and add it to the doc panel
 */
function loadDocs() {
    var r = new XMLHttpRequest()
    r.open("GET", "/static/docs-13.html", true)
    r.onreadystatechange = function () {
        if (r.readyState !== 4) { return }
        if (r.status === 200) {
            document.getElementById("docPanel").innerHTML = r.responseText
        }
    }
    r.send(null)
}

/**
 * Format both editors and run the current playground
 */
function run() {
    if (formatAll(false)) {

        showResult("running query...", false)

        var r = new XMLHttpRequest()
        r.open("POST", "/run")
        r.setRequestHeader("Content-Type", "application/x-www-form-urlencoded")
        r.onreadystatechange = function () {
            if (r.readyState !== 4) { return }
            if (r.status === 200) {

                configChangedSinceLastRun = false
                queryChangedSinceLastRun = false

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

/**
 * Save the current playground
 */
function save() {

    formatAll(configChangedSinceLastRun || queryChangedSinceLastRun)

    var r = new XMLHttpRequest()
    r.open("POST", "/save")
    r.setRequestHeader("Content-Type", "application/x-www-form-urlencoded")
    r.onreadystatechange = function () {
        if (r.readyState !== 4) { return }
        if (r.status === 200) {

            configOrQueryChangedSinceLastSave = false
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

/**
 * Encode the content of a playground as an URI
 * 
 * @param {boolean} keepComment - wether to keep comment or not 
 * 
 * @returns {string} the content of the playground URI encoded
 */
function encodePlayground(keepComment) {
    var result = "mode=" + comboMode.getValue()
    if (keepComment) {
        result += "&config=" + encodeURIComponent(parser.compact(configEditor.getValue(), "config", comboMode.getValue()))
            + "&query=" + encodeURIComponent(parser.compact(queryEditor.getValue(), "query", comboMode.getValue()))
    } else {
        result += "&config=" + encodeURIComponent(parser.compactAndRemoveComment(configEditor.getValue(), "config", comboMode.getValue(), -1))
            + "&query=" + encodeURIComponent(parser.compactAndRemoveComment(queryEditor.getValue(), "query", comboMode.getValue(), comboStages.getSelectedIndex() + 1))
    }
    return result
}

/**
 * Format both editors and returns wether they have any syntax errors
 * 
 * @param {boolean} clearResult - wether to clear the result editor or not 
 * 
 * @returns {boolean} true if there's no syntax errors in both editors
 */
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

    if (configChangedSinceLastRun) {
        configEditor.setValue(parser.indent(configEditor.getValue(), "config", comboMode.getValue()), 1)
    }
    if (queryChangedSinceLastRun) {
        queryEditor.setValue(parser.indent(queryEditor.getValue(), "query", comboMode.getValue()), 1)
    }
    return true
}

/**
 * Display an error ( syntax error or server error) in the result editor
 * 
 * @param {string} errMsg - error message to display in result editor 
 */
function showError(errMsg) {
    document.getElementById("result").classList.add("text_red")
    resultEditor.setOption("wrap", true)
    resultEditor.setValue(errMsg, -1)
}

/**
 * Display a valid result in the result editor
 * 
 * @param {string} result - the text to display in the result editor 
 * @param {boolean} doIndent - wether to indent the result or not 
 */
function showResult(result, doIndent) {
    document.getElementById("result").classList.remove("text_red")
    if (doIndent) {
        result = parser.indent(result, "result", comboMode.getValue())
    }
    resultEditor.setOption("wrap", false)
    resultEditor.setValue(result, -1)
}


// code to execute when script is loaded
init()