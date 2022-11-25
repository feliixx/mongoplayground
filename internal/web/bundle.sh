#!/bin/sh

HTML_HASH=$(md5sum static/docs.html | cut -d " " -f1)
sed -i "s/docs-.*\.html/docs-$HTML_HASH.html/" src/playground.js

# order of js files is important 
{
  cat vendored/ace.js 
  echo
  cat vendored/ext-language_tools.js
  echo
  cat src/mode-mongo.js
  echo
  cat src/completer.js 
  echo 
  cat src/custom_select.js 
  echo
  cat src/parser.js 
  echo
  cat src/playground.js
} > bundle.js


node_modules/esbuild/bin/esbuild --minify bundle.js > static/playground-min.js
JS_HASH=$(md5sum static/playground-min.js | cut -d " " -f1)
sed -i "s/playground-min-.*\.js/playground-min-$JS_HASH.js/" src/playground.html

node_modules/esbuild/bin/esbuild --minify src/playground.css > static/playground-min.css
CSS_HASH=$(md5sum static/playground-min.css | cut -d " " -f1)
sed -i "s/playground-min-.*\.css/playground-min-$CSS_HASH.css/" src/playground.html

rm -f bundle.js