# prom2bq
[![CI](https://github.com/ymyzk/prom2bq/actions/workflows/ci.yaml/badge.svg)](https://github.com/ymyzk/prom2bq/actions/workflows/ci.yaml)

## Usage
```console
$ go build
$ ./prom2bq \
    -start 2019-07-01T00:00:00Z \
    -end 2019-07-02T00:00:00Z \
    -prometheus http://<address of prometheus> \
    -bigquery <project>:<dataset>.<table> \
    -credential <path to GCP credential in JSON format> \
    <metric1> <metric2> ...
```
