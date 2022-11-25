#!/bin/sh

HTML_HASH=$(md5sum internal/web/static/docs.html | cut -d " " -f1)
sed -i "s/docs-.*\.html/docs-$HTML_HASH.html/" internal/web/playground.js

# order of js files is important 
{
  cat internal/web/ace.js 
  echo
  cat internal/web/ext-language_tools.js
  echo
  cat internal/web/mode-mongo.js
  echo
  cat internal/web/completer.js 
  echo 
  cat internal/web/custom_select.js 
  echo
  cat internal/web/parser.js 
  echo
  cat internal/web/playground.js
} > bundle.js


esbuild --minify bundle.js > internal/web/static/playground-min.js
JS_HASH=$(md5sum internal/web/static/playground-min.js | cut -d " " -f1)
sed -i "s/playground-min-.*\.js/playground-min-$JS_HASH.js/" internal/web/playground.html

esbuild --minify internal/web/playground.css > internal/web/static/playground-min.css
CSS_HASH=$(md5sum internal/web/static/playground-min.css | cut -d " " -f1)
sed -i "s/playground-min-.*\.css/playground-min-$CSS_HASH.css/" internal/web/playground.html

rm -f bundle.js