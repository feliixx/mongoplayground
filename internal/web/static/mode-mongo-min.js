define("ace/mode/mongo_highlight_rules",["require","exports","module","ace/lib/oop","ace/mode/text_highlight_rules"],function(r,d,v){"use strict";var u=r("../lib/oop"),h=r("./text_highlight_rules").TextHighlightRules,c=function(){var l=function(e){switch(e){case"ObjectId":case"ISODate":case"Timestamp":case"NumberInt":case"NumberLong":case"NumberDecimal":case"BinData":case"true":case"false":case"null":case"undefined":return"constant.language";case"db":case"find":case"aggregate":case"update":case"explain":return"storage.function";default:return"identifier"}};this.$rules={start:[{token:"comment",regex:"//.*"},{token:"string",regex:'["](?:(?:\\\\.)|(?:[^"\\\\]))*?["]'},{token:"constant.numeric",regex:"0[xX][0-9a-fA-F]+\\b"},{token:"constant.numeric",regex:"[+-]?\\d+(?:(?:\\.\\d*)?(?:[eE][+-]?\\d+)?)?\\b"},{token:"paren.lparen",regex:"[[({]"},{token:"paren.rparen",regex:"[\\])}]"},{token:l,regex:"[a-zA-Z_$][a-zA-Z0-9_$]*\\b"}]}};u.inherits(c,h),d.MongoHighlightRules=c}),define("ace/mode/matching_brace_outdent",["require","exports","module","ace/range"],function(r,d,v){"use strict";var u=r("../range").Range,h=function(){};(function(){this.checkOutdent=function(c,l){return/^\s+$/.test(c)?/^\s*\}/.test(l):!1},this.autoOutdent=function(c,l){var e=c.getLine(l),n=e.match(/^(\s*\})/);if(!n)return 0;var t=n[1].length,o=c.findMatchingBracket({row:l,column:t});if(!o||o.row==l)return 0;var a=this.$getIndent(c.getLine(o.row));c.replace(new u(l,0,l,t-1),a)},this.$getIndent=function(c){return c.match(/^\s*/)[0]}}).call(h.prototype),d.MatchingBraceOutdent=h}),define("ace/mode/folding/cstyle",["require","exports","module","ace/lib/oop","ace/range","ace/mode/folding/fold_mode"],function(r,d,v){"use strict";var u=r("../../lib/oop"),h=r("../../range").Range,c=r("./fold_mode").FoldMode,l=d.FoldMode=function(e){e&&(this.foldingStartMarker=new RegExp(this.foldingStartMarker.source.replace(/\|[^|]*?$/,"|"+e.start)),this.foldingStopMarker=new RegExp(this.foldingStopMarker.source.replace(/\|[^|]*?$/,"|"+e.end)))};u.inherits(l,c),function(){this.foldingStartMarker=/([\{\[\(])[^\}\]\)]*$|^\s*(\/\*)/,this.foldingStopMarker=/^[^\[\{\(]*([\}\]\)])|^[\s\*]*(\*\/)/,this.singleLineBlockCommentRe=/^\s*(\/\*).*\*\/\s*$/,this.tripleStarBlockCommentRe=/^\s*(\/\*\*\*).*\*\/\s*$/,this.startRegionRe=/^\s*(\/\*|\/\/)#?region\b/,this._getFoldWidgetBase=this.getFoldWidget,this.getFoldWidget=function(e,n,t){var o=e.getLine(t);if(this.singleLineBlockCommentRe.test(o)&&!this.startRegionRe.test(o)&&!this.tripleStarBlockCommentRe.test(o))return"";var a=this._getFoldWidgetBase(e,n,t);return!a&&this.startRegionRe.test(o)?"start":a},this.getFoldWidgetRange=function(e,n,t,o){var a=e.getLine(t);if(this.startRegionRe.test(a))return this.getCommentRegionBlock(e,a,t);var i=a.match(this.foldingStartMarker);if(i){var g=i.index;if(i[1])return this.openingBracketBlock(e,i[1],t,g);var s=e.getCommentFoldRange(t,g+i[0].length,1);return s&&!s.isMultiLine()&&(o?s=this.getSectionRange(e,t):n!="all"&&(s=null)),s}if(n!=="markbegin"){var i=a.match(this.foldingStopMarker);if(i){var g=i.index+i[0].length;return i[1]?this.closingBracketBlock(e,i[1],t,g):e.getCommentFoldRange(t,g,-1)}}},this.getSectionRange=function(e,n){var t=e.getLine(n),o=t.search(/\S/),a=n,i=t.length;n=n+1;for(var g=n,s=e.getLength();++n<s;){t=e.getLine(n);var m=t.search(/\S/);if(m!==-1){if(o>m)break;var f=this.getFoldWidgetRange(e,"all",n);if(f){if(f.start.row<=a)break;if(f.isMultiLine())n=f.end.row;else if(o==m)break}g=n}}return new h(a,i,g,e.getLine(g).length)},this.getCommentRegionBlock=function(e,n,t){for(var o=n.search(/\s*$/),a=e.getLength(),i=t,g=/^\s*(?:\/\*|\/\/|--)#?(end)?region\b/,s=1;++t<a;){n=e.getLine(t);var m=g.exec(n);if(!!m&&(m[1]?s--:s++,!s))break}var f=t;if(f>i)return new h(i,o,f,n.length)}}.call(l.prototype)}),define("ace/mode/mongo",["require","exports","module","ace/lib/oop","ace/mode/text","ace/mode/mongo_highlight_rules","ace/mode/matching_brace_outdent","ace/mode/behaviour/cstyle","ace/mode/folding/cstyle"],function(r,d,v){"use strict";var u=r("../lib/oop"),h=r("./text").Mode,c=r("./mongo_highlight_rules").MongoHighlightRules,l=r("./matching_brace_outdent").MatchingBraceOutdent,e=r("./behaviour/cstyle").CstyleBehaviour,n=r("./folding/cstyle").FoldMode,t=function(){this.HighlightRules=c,this.$outdent=new l,this.$behaviour=new e,this.foldingRules=new n};u.inherits(t,h),function(){this.lineCommentStart="//",this.getNextLineIndent=function(o,a,i){var g=this.$getIndent(a);if(o=="start"){var s=a.match(/^.*[\{\(\[]\s*$/);s&&(g+=i)}return g},this.checkOutdent=function(o,a,i){return this.$outdent.checkOutdent(a,i)},this.autoOutdent=function(o,a,i){this.$outdent.autoOutdent(a,i)},this.$id="ace/mode/mongo"}.call(t.prototype),d.Mode=t}),function(){window.require(["ace/mode/mongo"],function(r){typeof module=="object"&&typeof exports=="object"&&module&&(module.exports=r)})}();
