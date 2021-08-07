package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"google.golang.org/api/option"
)

const ChunkSize = 10000

type Item struct {
	Time   time.Time
	Name   string
	Value  float64
	Labels []string
}

func NewAPI(address string) (v1.API, error) {
	client, err := api.NewClient(api.Config{
		Address: address,
	})

	if err != nil {
		return nil, err
	}

	return v1.NewAPI(client), nil
}

func Query(ctx context.Context, api v1.API, metric string, start time.Time, end time.Time) (model.Matrix, error) {
	seconds := int(end.Sub(start).Seconds())
	if seconds < 0 {
		return nil, errors.New("start > end")
	}
	value, warnings, err := api.Query(
		ctx,
		fmt.Sprintf("%s[%ds]", metric, seconds),
		end,
	)
	if err != nil {
		return nil, err
	}
	if warnings != nil {
		return nil, errors.New(fmt.Sprintf("warnings: %+v", warnings))
	}
	return value.(model.Matrix), nil
}

func MsEpochToTime(ms int64) time.Time {
	return time.Unix(ms/int64(1000), (ms%int64(1000))*int64(1000000))
}

func ConvertToItems(data model.Matrix) []*Item {
	var items []*Item
	for _, sampleStream := range data {
		var labels []string
		for k, v := range sampleStream.Metric {
			if k == "__name__" {
				continue
			}
			labels = append(labels, fmt.Sprintf("%s=\"%s\"", k, v))
		}

		metricName := string(sampleStream.Metric["__name__"])
		for _, sample := range sampleStream.Values {
			items = append(items,
				&Item{
					Time:   MsEpochToTime(int64(sample.Timestamp)),
					Name:   metricName,
					Labels: labels,
					Value:  float64(sample.Value),
				},
			)
		}
	}
	return items
}

func CreateChunks(items []*Item, chunkSize int) (chunks [][]*Item) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}

func main() {
	config, err := ParseOptions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Configuration: %+v\n", config)

	api, err := NewAPI(config.PrometheusAddress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, config.BigQueryProject, option.WithCredentialsFile(config.CredentialFile))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	defer client.Close()

	uploader := client.Dataset(config.BigQueryDataset).Table(config.BigQueryTable).Uploader()

	for _, metric := range config.Metrics {
		fmt.Printf("Processing %s from %s to %s\n", metric, config.Start, config.End)
		data, err := Query(ctx, api, metric, config.Start, config.End)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
		items := ConvertToItems(data)
		fmt.Printf("Obtained %d records\n", len(items))
		for _, chunk := range CreateChunks(items, ChunkSize) {
			err = uploader.Put(ctx, chunk)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %s\n", err)
				os.Exit(1)
			}
		}
	}
}
