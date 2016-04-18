package datadoghq

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"io"

	"fmt"

	"golang.org/x/net/context"
)

const (
	TimeSeriesEndpoint = "https://app.datadoghq.com/api/v1/series?api_key={api_key}"
)

type Point struct {
	Timestamp time.Time
	Value     float32
}

func (p Point) MarshalJSON() ([]byte, error) {
	data := []interface{}{
		p.Timestamp.Unix(),
		p.Value,
	}
	return json.Marshal(data)
}

type Metric struct {
	Metric string   `json:"metric"`
	Points []Point  `json:"points,omitempty"`
	Type   string   `json:"type,omitempty"`
	Host   string   `json:"host,omitempty"`
	Tags   []string `json:"tags,omitempty"`
}

type Series struct {
	Series []Metric `json:"series,omitempty"`
}

type Client struct {
	cancel     func()
	ctx        context.Context
	apiKey     string
	ch         chan Metric
	flush      chan chan struct{}
	wg         *sync.WaitGroup
	interval   time.Duration
	bufferSize int
	endpoint   string
	out        io.Writer
	errOut     io.Writer
}

type Option func(*Client)

func New(apiKey string, options ...Option) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	endpoint := strings.Replace(TimeSeriesEndpoint, "{api_key}", apiKey, -1)
	client := &Client{
		cancel:     cancel,
		ctx:        ctx,
		apiKey:     apiKey,
		ch:         make(chan Metric, 256),
		flush:      make(chan chan struct{}),
		wg:         &sync.WaitGroup{},
		interval:   time.Second * 5,
		bufferSize: 256,
		endpoint:   endpoint,
		out:        io.MultiWriter(),
		errOut:     io.MultiWriter(),
	}

	for _, opt := range options {
		opt(client)
	}

	client.wg.Add(1)

	go client.start()

	return client
}

func (c *Client) start() {
	defer c.wg.Done()

	buffer := make([]Metric, c.bufferSize)
	index := 0
	timer := time.NewTimer(c.interval)

	for {
		timer.Reset(c.interval)

		select {
		case <-c.ctx.Done():
			return

		case v := <-c.ch:
			buffer[index] = v
			index++
			if index == c.bufferSize {
				c.post(buffer[0:index])
				index = 0
			}

		case v := <-c.flush:
			if index > 0 {
				c.post(buffer[0:index])
				index = 0
			}
			v <- struct{}{}

		case <-timer.C:
			if index > 0 {
				c.post(buffer[0:index])
				index = 0
			}
		}
	}
}

func (c *Client) post(metrics []Metric) error {
	data, err := json.Marshal(Series{metrics})
	if err != nil {
		fmt.Fprintf(c.errOut, "Unable to marshal time series, %v", err)
		return err
	}

	resp, err := http.Post(c.endpoint, "application/json", bytes.NewReader(data))
	if err != nil {
		fmt.Fprintf(c.errOut, "Unable to POST series to datadoghq, %v", err)
		return err
	}
	defer resp.Body.Close()
	io.Copy(c.out, resp.Body)
	return err
}

func (c *Client) Publish(metric Metric) {
	metric.Type = "gauge"
	c.ch <- metric
}

func (c *Client) Flush() {
	ch := make(chan struct{})
	defer close(ch)

	c.flush <- ch
	<-ch
}

func (c *Client) Close() {
	c.Flush()
	c.cancel()
	c.wg.Wait()
}

func Interval(v time.Duration) Option {
	return func(c *Client) {
		c.interval = v
	}
}

func BufferSize(v int) Option {
	return func(c *Client) {
		c.bufferSize = v
	}
}

func Output(w io.Writer) Option {
	return func(c *Client) {
		c.out = w
	}
}

func ErrorOutput(w io.Writer) Option {
	return func(c *Client) {
		c.errOut = w
	}
}
