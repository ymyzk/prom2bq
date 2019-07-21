package main

import (
	"errors"
	"flag"
	"strings"
	"time"
)

type Config struct {
	Start             time.Time
	End               time.Time
	PrometheusAddress string
	CredentialFile    string
	BigQueryProject   string
	BigQueryDataset   string
	BigQueryTable     string
	Metrics           []string
}

func ParseOptions() (*Config, error) {
	start := flag.String("start", "", "start time")
	end := flag.String("end", "", "end time")
	prometheus := flag.String("prometheus", "", "URL of Prometheus")
	credential := flag.String("credential", "", "Credential file")
	bigquery := flag.String("bigquery", "", "BigQuery (project:dataset.table)")
	flag.Parse()

	if *start == "" {
		return nil, errors.New("start option is required")
	}

	if *end == "" {
		return nil, errors.New("end option is required")
	}

	if *prometheus == "" {
		return nil, errors.New("prometheus option is required")
	}

	if *credential == "" {
		return nil, errors.New("credential option is required")
	}

	if *bigquery == "" {
		return nil, errors.New("bigquery option is required")
	}

	parsedStart, err := time.Parse(time.RFC3339, *start)
	if err != nil {
		return nil, err
	}

	parsedEnd, err := time.Parse(time.RFC3339, *end)
	if err != nil {
		return nil, err
	}

	bigqueryComponents := strings.FieldsFunc(*bigquery, func(r rune) bool {
		return r == ':' || r == '.'
	})
	if len(bigqueryComponents) != 3 {
		return nil, errors.New("invalid bigquery argument format")
	}

	metrics := flag.Args()

	return &Config{
		Start:             parsedStart,
		End:               parsedEnd,
		PrometheusAddress: *prometheus,
		CredentialFile:    *credential,
		BigQueryProject:   bigqueryComponents[0],
		BigQueryDataset:   bigqueryComponents[1],
		BigQueryTable:     bigqueryComponents[2],
		Metrics:           metrics,
	}, nil
}
