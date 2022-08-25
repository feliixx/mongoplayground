FROM golang:alpine3.16 as esbuild

WORKDIR /tools

RUN go install github.com/evanw/esbuild/cmd/esbuild@v0.14.11
RUN mv ${GOPATH}/bin/esbuild . 

FROM golang:alpine3.16 as builder

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

RUN --mount=type=cache,target=/root/.cache/go-build go build

FROM alpine:3.16

WORKDIR /app

COPY --from=builder /app/mongoplayground /usr/bin/
COPY config.json . 

RUN mkdir storage
RUN mkdir backups 

ENTRYPOINT ["mongoplayground"]
