#!/bin/bash 
if [ $1 == "all" ]; then 
  echo '<div class="markdown-body">' >> static/docs.html
  curl https://api.github.com/markdown/raw -X "POST" -H "Content-Type: text/plain" -d "$(cat web/DOCS.md)" >> static/docs.html
  echo '</div>' >> static/docs.html
fi


  purifycss web/playground.css web/github.css static/docs.html playground.html --whitelist ["ignoreWarnings", "ace_gutter","ace_layer","ace_warning", "ace_string", "ace_numeric", "ace_function", "ace_editor", "ace_error"] --min --info --out "static/playground-min-2.css"  
  gzip --best --verbose --force static/playground-min-2.css

if [ $1 == "all" ]; then 
  mv static/docs.html static/docs-2.html
  gzip --best --force --verbose static/docs-2.html
fi
