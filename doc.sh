#!/bin/bash 
if [ $1 == "all" ]; then 
  echo '<div class="markdown-body">' > static/docs.html
  curl https://api.github.com/markdown/raw -X "POST" -H "Content-Type: text/plain" -d "$(cat web/DOCS.md)" >> static/docs.html
  echo '</div>' >> static/docs.html
fi


  purifycss web/playground.css web/github.css web/playground.js static/docs.html playground.html --min --info --out "static/playground-min.css"  

  # to install uglify-js, run 'npm install -g uglify-js'
  uglifyjs  web/playground.js --compress  --verbose --mangle --output  static/playground-min.js