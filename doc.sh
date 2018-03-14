#!/bin/bash 
if [ $1 == "all" ]; then 
  echo '<div class="markdown-body">' >> static/docs.html
  curl https://api.github.com/markdown/raw -X "POST" -H "Content-Type: text/plain" -d "$(cat web/DOCS.md)" >> static/docs.html
  echo '</div>' >> static/docs.html
fi


  purifycss web/playground.css web/github.css web/playground.js static/docs.html playground.html --whitelist ["ignoreWarnings", "ace_gutter","ace_layer","ace_warning", "ace_string", "ace_numeric", "ace_function", "ace_editor", "ace_error"] --min --info --out "static/playground-min.css"  
  gzip --best --verbose --force static/playground-min.css

  uglifyjs  web/playground.js --compress  --verbose --mangle --output  static/playground-min.js
  gzip --best --verbose --force static/playground-min.js

if [ $1 == "all" ]; then 
  gzip --best --force --verbose static/docs.html
fi
