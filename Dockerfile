FROM golang:latest as esbuild

WORKDIR /tools

RUN go install github.com/evanw/esbuild/cmd/esbuild@v0.14.11
RUN mv ${GOPATH}/bin/esbuild . 

FROM golang:latest as builder

COPY --from=esbuild /tools/esbuild /usr/bin/

WORKDIR /app 

COPY go.mod .
COPY go.sum . 

RUN go mod download

COPY internal/web ./internal/web
COPY bundle.sh . 
RUN ./bundle.sh

COPY main.go . 
COPY internal/*.go ./internal/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" .

FROM busybox

WORKDIR /app

COPY --from=builder /app/mongoplayground /usr/bin/
COPY config.json . 

RUN mkdir storage
RUN mkdir backups 

ENTRYPOINT ["mongoplayground"]
