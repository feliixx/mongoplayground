define("ace/mode/mongo_highlight_rules", ["require", "exports", "module", "ace/lib/oop", "ace/mode/text_highlight_rules"], function (require, exports, module) {
    "use strict";

    var oop = require("../lib/oop")
    var TextHighlightRules = require("./text_highlight_rules").TextHighlightRules

    var MongoHighlightRules = function () {

        var keywordMapper = function (value) {
            switch (value) {
                case "ObjectId":
                case "ISODate":
                case "Timestamp":
                case "NumberInt":
                case "NumberLong":
                case "NumberDecimal":
                case "BinData":
                case "true":
                case "false":
                case "null":
                case "undefined":
                    return "constant.language"
                case "db":
                case "find":
                case "aggregate":
                case "update":
                case "explain":
                    return "storage.function"
                default:
                    return "identifier"
            }
        }
        this.$rules = {
            "start": [
                {token: "comment", regex: "//.*"},
                {token: "string", regex: '["](?:(?:\\\\.)|(?:[^"\\\\]))*?["]'},
                {token: "constant.numeric", regex: "0[xX][0-9a-fA-F]+\\b"},
                {token: "constant.numeric", regex: "[+-]?\\d+(?:(?:\\.\\d*)?(?:[eE][+-]?\\d+)?)?\\b"},
                {token: "paren.lparen", regex: "[[({]"},
                {token: "paren.rparen", regex: "[\\])}]"},
                {token: keywordMapper, regex: "[a-zA-Z_$][a-zA-Z0-9_$]*\\b"}
            ]
        }
    }

    oop.inherits(MongoHighlightRules, TextHighlightRules)
    exports.MongoHighlightRules = MongoHighlightRules
});

define("ace/mode/matching_brace_outdent", ["require", "exports", "module", "ace/range"], function (require, exports, module) {
    "use strict";

    var Range = require("../range").Range

    var MatchingBraceOutdent = function () {

        this.checkOutdent = function (line, input) {
            if (!/^\s+$/.test(line)) {
                return false
            }
            return /^\s*\}/.test(input)
        }

        this.autoOutdent = function (doc, row) {
            var line = doc.getLine(row)
            var match = line.match(/^(\s*\})/)

            if (!match) {
                return 0
            }

            var column = match[1].length
            var openBracePos = doc.findMatchingBracket({row: row, column: column})

            if (!openBracePos || openBracePos.row == row) {
                return 0
            }

            var indent = this.$getIndent(doc.getLine(openBracePos.row))
            doc.replace(new Range(row, 0, row, column - 1), indent)
        }

        this.$getIndent = function (line) {
            return line.match(/^\s*/)[0];
        }
    }

    exports.MatchingBraceOutdent = MatchingBraceOutdent
});

define("ace/mode/folding/cstyle", ["require", "exports", "module", "ace/lib/oop", "ace/mode/folding/fold_mode"], function (require, exports, module) {
    "use strict";

    var oop = require("../../lib/oop")
    var BaseFoldMode = require("./fold_mode").FoldMode

    var FoldMode = function () {

        this.foldingStartMarker = /([\{\[\(])[^\}\]\)]*$/
        this.foldingStopMarker = /^[^\[\{\(]*([\}\]\)])/
        this._getFoldWidgetBase = this.getFoldWidget
        this.getFoldWidget = function (session, foldStyle, row) {
            return this._getFoldWidgetBase(session, foldStyle, row)
        }

        this.getFoldWidgetRange = function (session, foldStyle, row, forceMultiline) {
            var line = session.getLine(row)

            var match = line.match(this.foldingStartMarker)
            if (match.length >= 2) {
                return this.openingBracketBlock(session, match[1], row, match.index)
            }

            if (foldStyle === "markbegin") {
                return
            }

            match = line.match(this.foldingStopMarker)
            if (match.length >= 2) {
                return this.closingBracketBlock(session, match[1], row, match.index + match[0].length)
            }
        }
    }

    oop.inherits(FoldMode, BaseFoldMode)
    exports.FoldMode = FoldMode
});

define("ace/mode/mongo", ["require", "exports", "module", "ace/lib/oop", "ace/mode/text", "ace/mode/mongo_highlight_rules", "ace/mode/matching_brace_outdent", "ace/mode/behaviour/cstyle", "ace/mode/folding/cstyle"], function (require, exports, module) {
    "use strict";

    var oop = require("../lib/oop")
    var TextMode = require("./text").Mode
    var HighlightRules = require("./mongo_highlight_rules").MongoHighlightRules
    var MatchingBraceOutdent = require("./matching_brace_outdent").MatchingBraceOutdent
    var CstyleBehaviour = require("./behaviour/cstyle").CstyleBehaviour
    var CStyleFoldMode = require("./folding/cstyle").FoldMode

    var Mode = function () {

        this.HighlightRules = HighlightRules
        this.$outdent = new MatchingBraceOutdent()
        this.$behaviour = new CstyleBehaviour()
        this.foldingRules = new CStyleFoldMode()

        this.getNextLineIndent = function (state, line, tab) {
            var indent = this.$getIndent(line)

            if (state == "start") {
                var match = line.match(/^.*[\{\(\[]\s*$/)
                if (match) {
                    indent += tab
                }
            }
            return indent
        }

        this.checkOutdent = function (state, line, input) {
            return this.$outdent.checkOutdent(line, input)
        }

        this.autoOutdent = function (state, doc, row) {
            this.$outdent.autoOutdent(doc, row)
        }

        this.$id = "ace/mode/mongo"
    }

    oop.inherits(Mode, TextMode)
    exports.Mode = Mode
});

(function () {
    window.require(["ace/mode/mongo"], function (m) {
        if (typeof module == "object" && typeof exports == "object" && module) {
            module.exports = m
        }
    })
})();
            