FROM golang:1.16

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install

ENTRYPOINT ["/go/bin/prom2bq"]
