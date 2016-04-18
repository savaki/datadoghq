package main

import (
	"os"
	"time"

	"github.com/savaki/datadoghq"
)

func main() {
	apiKey := os.Getenv("DATADOG_API_KEY")
	client := datadoghq.New(apiKey,
		datadoghq.Output(os.Stderr),
		datadoghq.ErrorOutput(os.Stderr),
		datadoghq.Interval(time.Minute),
	)
	client.Publish(datadoghq.Metric{
		Metric: "sampler.metric",
		Points: []datadoghq.Point{
			{
				Timestamp: time.Now(),
				Value:     12.34,
			},
		},
		Tags: []string{
			"environment:local",
		},
	})
	client.Publish(datadoghq.Metric{
		Metric: "sampler.metric",
		Points: []datadoghq.Point{
			{
				Timestamp: time.Now(),
				Value:     56.78,
			},
		},
		Tags: []string{
			"environment:local",
		},
	})
	time.Sleep(time.Millisecond * 50)
	client.Flush()
	client.Close()
}
