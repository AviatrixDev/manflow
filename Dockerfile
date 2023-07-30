FROM golang:alpine as build
COPY . /src
WORKDIR /src
RUN go build -v .

FROM alpine:latest

WORKDIR /app

COPY --from=build /src/manflow /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/manflow"]
