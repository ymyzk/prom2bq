FROM golang:1.16 AS build
WORKDIR /go/src/app
COPY . /go/src/app
RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install

FROM gcr.io/distroless/static
COPY --from=build /go/bin/prom2bq /
ENTRYPOINT ["/prom2bq"]
