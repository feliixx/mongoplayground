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

var Playground = function () {

    let configChangedSinceLastRun = true
    let queryChangedSinceLastRun = true
    let configOrQueryChangedSinceLastSave = true

    let isConfigHandlerDragging = false
    let isQueryHandlerDragging = false

    const configPanel = document.getElementById("configPanel")
    const queryPanel = document.getElementById("queryPanel")
    const resultPanel = document.getElementById("resultPanel")
    const docPanel = document.getElementById("docPanel")

    const link = document.getElementById("link")
    const shareBtn = document.getElementById("share")

    const commonOpts = {
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

    const configEditor = ace.edit(document.getElementById("config"), commonOpts)
    const queryEditor = ace.edit(document.getElementById("query"), commonOpts)
    const resultEditor = ace.edit(document.getElementById("result"), {
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

    const comboStages = new CustomSelect({
        selectId: "aggregation_stages",
        onChange: run
    })
    const comboMode = new CustomSelect({
        selectId: "mode",
        onChange: checkEditorContent.bind(null, configEditor, "config")
    })
    const comboTemplate = new CustomSelect({
        selectId: "template",
        onChange: () => { setTemplate(comboTemplate.getSelectedIndex()) }
    })
    document.getElementById("labelTemplate").style.visibility = "visible"

    const customStages = document.getElementById("custom-aggregation_stages")
    const labelStages = document.getElementById("aggregation_stages_label")

    resultEditor.renderer.$cursorLayer.element.style.display = "none"

    const parser = new Parser()
    const completer = new Completer({
        parser: parser
    })

    configEditor.completers = [completer.configCompleter]
    queryEditor.completers = [completer.queryCompleter]

    configEditor.getSession().on("change", checkEditorContent.bind(null, configEditor, "config"))
    queryEditor.getSession().on("change", checkEditorContent.bind(null, queryEditor, "query"))

    configEditor.setValue(parser.indent(configEditor.getValue(), "config", comboMode.getValue()), -1)
    queryEditor.setValue(parser.indent(queryEditor.getValue(), "query", comboMode.getValue()), -1)

    document.querySelector("div.content").style.visibility = "visible"

    configChangedSinceLastRun = false
    queryChangedSinceLastRun = false
    configOrQueryChangedSinceLastSave = false

    document.addEventListener("keydown", event => {
        if ((event.ctrlKey || event.metaKey) && event.key === "Enter") {
            event.preventDefault()
            run()
        }
        if ((event.ctrlKey || event.metaKey) && event.key === "s") {
            event.preventDefault()
            formatAll()
        }
    })

    document.addEventListener("mousedown", event => {
        if (event.target.id === "configResizeHandler") {
            isConfigHandlerDragging = true
        }
        if (event.target.id === "queryResizeHandler") {
            isQueryHandlerDragging = true
        }
    })

    document.addEventListener("mousemove", event => {
        let box
        if (isConfigHandlerDragging) {
            box = configPanel
        } else if (isQueryHandlerDragging) {
            box = queryPanel
        } else {
            return false
        }
        let pointerRelativeXpos = event.clientX - box.offsetLeft
        let width = Math.max(60, pointerRelativeXpos + 2)

        box.style.width = `${width}px`
        box.style.flexGrow = "0"
    })

    document.addEventListener("mouseup", () => {
        isConfigHandlerDragging = false
        isQueryHandlerDragging = false
    })

    document.getElementById("run").addEventListener("click", run)
    document.getElementById("format").addEventListener("click", formatAll)
    document.getElementById("share").addEventListener("click", save)
    document.getElementById("showDoc").addEventListener("click", toggleDoc)

    document.querySelectorAll("[data-tooltip]").forEach(elem => {
        const container = document.createElement("div")
        container.className = "tooltip"
        elem.parentNode.insertBefore(container, elem)

        const span = document.createElement("span")
        span.innerHTML = elem.getAttribute("data-tooltip")
        span.className = "tooltiptext"
        span.classList.add('tooltip-hover')
        if (elem.id == "link") {
            span.id = "link_tooltip"
            span.classList.remove('tooltip-hover')
        }
        container.appendChild(span)
        container.appendChild(elem)
    })

    /**
     * Check editor content for syntax error
     * 
     * @param {ace.Editor} editor - the ace editor to check content from
     * @param {string} type - type of editor, must be one of ["config", "query"] 
     */
    function checkEditorContent(editor, type) {

        let errors = []

        const err = parser.parse(editor.getValue(), type, comboMode.getValue())
        if (err != null) {
            const pos = editor.getSession().getDocument().indexToPosition(err.at - 1)
            errors.push({
                row: pos.row,
                column: pos.column,
                text: err.message,
                type: "error"
            })
        }
        editor.getSession().setAnnotations(errors)

        if (type === "query") {
            if (parser.getQueryType() === "aggregate" && parser.getAggregationStages().length > 0) {
                comboStages.setOptions(parser.getAggregationStages())
                customStages.style.visibility = "visible"
                labelStages.style.visibility = "visible"
            } else {
                customStages.style.visibility = "hidden"
                labelStages.style.visibility = "hidden"
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
            document.getElementById("link_tooltip").classList.remove("tooltip-fadein-fadeout")
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
        link.style.visibility = showLink ? "visible" : "hidden"
        link.innerHTML = url
        shareBtn.disabled = showLink
    }

    const templates = [
        {
            config: '[{"key":1},{"key":2}]',
            query: 'db.collection.find()',
            mode: 'bson'
        },
        {
            config: 'db={"orders":[{"_id":1,"item":"almonds","price":12,"quantity":2},{"_id":2,"item":"pecans","price":20,"quantity":1},{"_id":3}],"inventory":[{"_id":1,"sku":"almonds","description":"product 1","instock":120},{"_id":2,"sku":"bread","description":"product 2","instock":80},{"_id":3,"sku":"cashews","description":"product 3","instock":60},{"_id":4,"sku":"pecans","description":"product 4","instock":70},{"_id":5,"sku":null,"description":"Incomplete"}]}',
            query: 'db.orders.aggregate([{"$lookup":{"from":"inventory","localField":"item","foreignField":"sku","as":"inventory_docs"}}])',
            mode: 'bson'
        },
        {
            config: '[{"collection":"collection","count":10,"content":{"key":{"type":"int","min":0,"max":10}}}]',
            query: 'db.collection.find()',
            mode: 'mgodatagen'
        },
        {
            config: '[{"key":1},{"key":2}]',
            query: 'db.collection.update({"key":2},{"$set":{"updated":true}},{"multi":false,"upsert":false})',
            mode: 'bson'
        },
        {
            config: '[{"collection":"collection","count":5,"content":{"description":{"type":"enum","values":["Coffee and cakes","Gourmet hamburgers","Just coffee","Discount clothing","Indonesian goods"]}},"indexes":[{"name":"description_text_idx","key":{"description":"text"}}]}]',
            query: 'db.collection.find({"$text":{"$search":"coffee"}})',
            mode: 'mgodatagen'
        },
        {
            config: '[{"_id":1,"item":"ABC","price":80,"sizes":["S","M","L"]},{"_id":2,"item":"EFG","price":120,"sizes":[]},{"_id":3,"item":"IJK","price":160,"sizes":"M"},{"_id":4,"item":"LMN","price":10},{"_id":5,"item":"XYZ","price":5.75,"sizes":null}]',
            query: 'db.collection.aggregate([{"$unwind":{"path":"$sizes","preserveNullAndEmptyArrays":true}},{"$group":{"_id":"$sizes","averagePrice":{"$avg":"$price"}}},{"$sort":{"averagePrice":-1}}]).explain("executionStats")',
            mode: 'bson'
        }
    ]

    /**
     * Fill config and query editor with a specific template 
     * 
     * @param {Number} index - index of the template to use  
     */
    function setTemplate(index) {
        comboMode.setValue(templates[index].mode)
        configEditor.setValue(parser.indent(templates[index].config, "config", comboMode.getValue()), 1)
        queryEditor.setValue(parser.indent(templates[index].query, "query", comboMode.getValue()), 1)
        resultEditor.setValue("", 1)
    }

    function toggleDoc() {
        if (docPanel.style.display === "inline") {
            hideDoc()
        } else {
            showDoc()
        }
    }

    function showDoc() {

        if (!docPanel.hasChildNodes()) {
            loadDocs()
        }

        docPanel.style.display = "inline"
        queryPanel.style.display = "none"
        resultPanel.style.display = "none"
    }

    function hideDoc() {
        docPanel.style.display = "none"
        queryPanel.style.display = "inline"
        resultPanel.style.display = "inline"
    }

    /**
     * load the documentation and add it to the doc panel
     */
    async function loadDocs() {
        const r = await fetch("/static/docs-c310647d0539a44970e85f228788385b.html", { method: "GET" })
        if (!r.ok) {
            return showError(`Failed to fetch doc: ${r.status} ${await r.text()}`)
        }
        docPanel.innerHTML = await r.text()
    }

    /**
     * Format both editors and run the current playground
     */
    async function run() {

        if (hasSyntaxError()) {
            return
        }
        formatAll()
        showResult("running query...", false)

        const r = await fetch("/run", { method: "POST", body: encodePlayground(false) })
        if (!r.ok) {
            return showError(`Failed to run playground: ${r.status} ${await r.text()}`)
        }

        configChangedSinceLastRun = false
        queryChangedSinceLastRun = false

        const result = await r.text()
        if (result.startsWith("[") || result.startsWith("{")) {
            return showResult(result, true)
        }
        if (result === "no document found") {
            return showResult(result, false)
        }
        showError(result)
    }

    /**
     * Save the current playground. The playground can be saved even if 
     * it contains syntax errors 
     */
    async function save() {

        formatAll()

        const r = await fetch("/save", { method: "POST", body: encodePlayground(true) })
        if (!r.ok) {
            return showError(`Failed to save playground: ${r.status} ${await r.text()}`)
        }

        configOrQueryChangedSinceLastSave = false

        const result = await r.text()
        if (!result.startsWith("http")) {
            return showError(result)
        }
        redirect(result, true)
        navigator.clipboard.writeText(result);
        document.getElementById("link_tooltip").classList.add("tooltip-fadein-fadeout")
    }

    /**
     * Encode the content of a playground as an URI
     * 
     * @param {boolean} keepComment - wether to keep comment or not 
     * 
     * @returns {FormData} a formData containing the mode, config and query 
     */
    function encodePlayground(keepComment) {

        let mode = comboMode.getValue()
        let compactFunc = keepComment ? parser.compact : parser.compactAndRemoveComment

        const formData = new FormData()
        formData.append("mode", mode)
        formData.append("config", compactFunc(configEditor.getValue(), "config", mode))
        formData.append("query", compactFunc(queryEditor.getValue(), "query", mode, comboStages.getSelectedIndex() + 1))

        return formData
    }

    /**
     * Check wether there is any syntax error in config or query editor 
     * 
     * @returns {boolean} true if there is at least one syntax error, in config or query editor
     */
    function hasSyntaxError() {

        let errors = configEditor.getSession().getAnnotations()
        if (errors.length > 0) {
            showError(`Invalid configuration:\n\nLine ${(errors[0].row + 1)}: ${errors[0].text}`)
            return true
        }
        errors = queryEditor.getSession().getAnnotations()
        if (errors.length > 0) {
            showError(`Invalid query:\n\nLine ${(errors[0].row + 1)}: ${errors[0].text}`)
            return true
        }
        return false
    }

    /**
     * Format both config and query editors 
     */
    function formatAll() {

        hideDoc()

        if (hasSyntaxError()) {
            return
        }

        if (configChangedSinceLastRun || queryChangedSinceLastRun) {
            resultEditor.setValue("", -1)
        }
        if (configChangedSinceLastRun) {
            configEditor.setValue(parser.indent(configEditor.getValue(), "config", comboMode.getValue()), 1)
        }
        if (queryChangedSinceLastRun) {
            queryEditor.setValue(parser.indent(queryEditor.getValue(), "query", comboMode.getValue()), 1)
        }
    }

    /**
     * Display an error ( syntax error or server error) in the result editor
     * 
     * @param {string} errMsg - error message to display in result editor 
     */
    function showError(errMsg) {

        hideDoc()

        resultPanel.classList.add("text_red")
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
        resultPanel.classList.remove("text_red")
        if (doIndent) {
            result = parser.indent(result, "result", comboMode.getValue())
        }
        resultEditor.setOption("wrap", false)
        resultEditor.setValue(result, -1)
    }
}

window.onload = () => {
    new Playground()
}