#!/bin/bash 

nb=7
if [ $1 == "all" ]; then 
  echo '<div class="markdown-body">' > static/docs-$nb.html
  curl https://api.github.com/markdown/raw -X "POST" -H "Content-Type: text/plain" -d "$(cat web/DOCS.md)" >> static/docs-$nb.html
  echo '</div>' >> static/docs-$nb.html
fi


  purifycss web/playground.css web/github.css static/docs-$nb.html playground.html --whitelist ["ace_gutter","ace_layer","ace_warning", "ace_info", "ace_string", "ace_numeric", "ace_function", "ace_editor", "ace_error", "text_red", "ignore_warnings"] --min --info -r --out "static/playground-min-$nb.css"  

  uglifyjs  web/playground.js --compress  --verbose --mangle --output  static/playground-min-$nb.js
