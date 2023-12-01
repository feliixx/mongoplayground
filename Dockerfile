FROM node:19-alpine as builder

WORKDIR /

COPY ./internal ./

COPY ./internal/web ./

WORKDIR /internal/web

COPY ./internal/web/bundle.sh ./

RUN npm i 

RUN chmod +x bundle.sh

FROM golang:1.21-alpine

WORKDIR /app

COPY --from=builder ./internal ./

COPY --from=builder ./internal/web ./

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY . .

COPY *.go ./

# Build
RUN go build -v -o /mongoplayground

EXPOSE 8080

# Run
CMD ["/mongoplayground"]