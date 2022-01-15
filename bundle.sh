#!/bin/sh

# order of js files is important 
{
  cat internal/web/snippets.js 
  echo 
  cat internal/web/custom_select.js 
  echo
  cat internal/web/parser.js 
  echo
  cat internal/web/ace.js 
  echo
  cat internal/web/ext-language_tools.js
  echo
  cat internal/web/mode-mongo.js
  echo
  cat internal/web/playground.js
} > bundle.js


esbuild --minify bundle.js > internal/web/static/playground-min.js
esbuild --minify internal/web/playground.css > internal/web/static/playground-min.css

rm -f bundle.js