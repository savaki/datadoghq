# datadoghq
datadoghq http api

``` golang
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
```