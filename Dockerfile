FROM golang:latest as esbuild

WORKDIR /tools

RUN go install github.com/evanw/esbuild/cmd/esbuild@v0.13.8
RUN mv ${GOPATH}/bin/esbuild . 

FROM golang:latest as builder

COPY --from=esbuild /tools/esbuild /usr/bin/

WORKDIR /app 

COPY go.mod .
COPY go.sum . 

RUN go mod download

COPY internal/web ./internal/web

RUN touch bundle.js
RUN cat internal/web/ace.js > bundle.js
RUN cat internal/web/ext-language_tools.js >> bundle.js
RUN cat internal/web/mode-mongo.js >> bundle.js
RUN cat internal/web/playground.js >> bundle.js

RUN esbuild --minify internal/web/playground.css > internal/web/static/playground-min.css
RUN esbuild --minify bundle.js > internal/web/static/playground-min.js

COPY main.go . 
COPY internal/*.go ./internal/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" .

FROM busybox

WORKDIR /app

COPY --from=builder /app/mongoplayground /usr/bin/

RUN mkdir storage
RUN mkdir backups 

ENTRYPOINT ["mongoplayground"]
