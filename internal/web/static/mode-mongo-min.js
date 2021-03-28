define("ace/mode/mongo_highlight_rules",["require","exports","module","ace/lib/oop","ace/mode/text_highlight_rules"],function(n,c,g){"use strict";var d=n("../lib/oop"),s=n("./text_highlight_rules").TextHighlightRules,o=function(){var e=function(r){switch(r){case"ObjectId":case"ISODate":case"Timestamp":case"NumberInt":case"NumberLong":case"NumberDecimal":case"BinData":case"true":case"false":case"null":case"undefined":return"constant.language";case"db":case"find":case"aggregate":case"update":case"explain":return"storage.function";default:return"identifier"}};this.$rules={start:[{token:"comment",regex:"//.*"},{token:"string",regex:'["](?:(?:\\\\.)|(?:[^"\\\\]))*?["]'},{token:"constant.numeric",regex:"0[xX][0-9a-fA-F]+\\b"},{token:"constant.numeric",regex:"[+-]?\\d+(?:(?:\\.\\d*)?(?:[eE][+-]?\\d+)?)?\\b"},{token:"paren.lparen",regex:"[[({]"},{token:"paren.rparen",regex:"[\\])}]"},{token:e,regex:"[a-zA-Z_$][a-zA-Z0-9_$]*\\b"}]}};d.inherits(o,s),c.MongoHighlightRules=o}),define("ace/mode/matching_brace_outdent",["require","exports","module","ace/range"],function(n,c,g){"use strict";var d=n("../range").Range,s=function(){this.checkOutdent=function(o,e){return/^\s+$/.test(o)?/^\s*\}/.test(e):!1},this.autoOutdent=function(o,e){var r=o.getLine(e),i=r.match(/^(\s*\})/);if(!i)return 0;var u=i[1].length,a=o.findMatchingBracket({row:e,column:u});if(!a||a.row==e)return 0;var t=this.$getIndent(o.getLine(a.row));o.replace(new d(e,0,e,u-1),t)},this.$getIndent=function(o){return o.match(/^\s*/)[0]}};c.MatchingBraceOutdent=s}),define("ace/mode/folding/cstyle",["require","exports","module","ace/lib/oop","ace/mode/folding/fold_mode"],function(n,c,g){"use strict";var d=n("../../lib/oop"),s=n("./fold_mode").FoldMode,o=function(){this.foldingStartMarker=/([\{\[\(])[^\}\]\)]*$/,this.foldingStopMarker=/^[^\[\{\(]*([\}\]\)])/,this._getFoldWidgetBase=this.getFoldWidget,this.getFoldWidget=function(e,r,i){return this._getFoldWidgetBase(e,r,i)},this.getFoldWidgetRange=function(e,r,i,u){var a=e.getLine(i),t=a.match(this.foldingStartMarker);if(t.length>=2)return this.openingBracketBlock(e,t[1],i,t.index);if(r!=="markbegin"&&(t=a.match(this.foldingStopMarker),t.length>=2))return this.closingBracketBlock(e,t[1],i,t.index+t[0].length)}};d.inherits(o,s),c.FoldMode=o}),define("ace/mode/mongo",["require","exports","module","ace/lib/oop","ace/mode/text","ace/mode/mongo_highlight_rules","ace/mode/matching_brace_outdent","ace/mode/behaviour/cstyle","ace/mode/folding/cstyle"],function(n,c,g){"use strict";var d=n("../lib/oop"),s=n("./text").Mode,o=n("./mongo_highlight_rules").MongoHighlightRules,e=n("./matching_brace_outdent").MatchingBraceOutdent,r=n("./behaviour/cstyle").CstyleBehaviour,i=n("./folding/cstyle").FoldMode,u=function(){this.HighlightRules=o,this.$outdent=new e,this.$behaviour=new r,this.foldingRules=new i,this.getNextLineIndent=function(a,t,l){var h=this.$getIndent(t);if(a=="start"){var f=t.match(/^.*[\{\(\[]\s*$/);f&&(h+=l)}return h},this.checkOutdent=function(a,t,l){return this.$outdent.checkOutdent(t,l)},this.autoOutdent=function(a,t,l){this.$outdent.autoOutdent(t,l)},this.$id="ace/mode/mongo"};d.inherits(u,s),c.Mode=u}),function(){window.require(["ace/mode/mongo"],function(n){typeof module=="object"&&typeof exports=="object"&&module&&(module.exports=n)})}();