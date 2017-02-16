package main

import (
	"testing"

	"time"

	"errors"
	"net/url"

	"github.com/antonholmquist/jason"
	"github.com/stretchr/testify/assert"
)

func NewSafeMetric(name string) *Metric {
	m, _ := NewMetric(name)
	return m
}

func WaitTime(ch chan bool, timeout time.Duration) error {
	select {
	case <-ch:
		return nil
	case <-time.After(timeout):
	}
	return errors.New("timeout")
}

type mockFetcher struct {
	timeout time.Duration
	err     error
	vars    *Expvars
}

func (f *mockFetcher) Fetch(url url.URL) (*Expvars, error) {
	if f.timeout.Nanoseconds() > 0 {
		time.Sleep(f.timeout)
	}
	return f.vars, f.err
}

func TestCrawler_Start_Success(t *testing.T) {
	Now = func() time.Time {
		t, _ := time.Parse("2006-Jan-02", "2013-Feb-03")
		return t
	}

	defer func() {
		Now = time.Now
	}()

	o, err := jason.NewObjectFromBytes([]byte(`{"gauge": {"metric": 800}, "process": {"text": "text 1"}, "memstats": {"alloc": 123}}`))
	assert.NoError(t, err)

	crawler := &Crawler{
		interval: 200 * time.Millisecond,
		fetcher: &mockFetcher{
			vars: &Expvars{Object: o},
		},
		hub: &Hub{
			dataCh: make(chan []byte, 1),
		},
		services: []*Service{
			{
				Name: "service1",
			},
		},
		widgets: &Widgets{
			Gauges: []*Gauge{
				{
					cid:      "g1",
					Metric:   NewSafeMetric("gauge.metric"),
					MaxValue: 1000,
					Service:  "service1",
				},
			},
			LineCharts: []*LineChart{
				{
					cid:    "lc1",
					Metric: NewSafeMetric("memstats.alloc"),
				},
			},
			Texts: []*Text{
				{
					cid:     "t1",
					Metric:  NewSafeMetric("process.text"),
					Service: "service1",
				},
			},
		},
		done: make(chan struct{}, 1),
	}

	go crawler.Start()
	defer crawler.Stop()

	ch := make(chan bool)

	go func() {
		assert.Equal(t, `{"g":[{"i":"g1","v":0.8}],"lc":[{"i":"lc1","p":[{"time":1359849600,"y":123}]}],"t":[{"i":"t1","v":"text 1"}]}`, string(<-crawler.hub.dataCh))

		ch <- true
	}()

	if err := WaitTime(ch, time.Second); err != nil {
		t.Fatal("Did not get response in time")
	}
}

func TestCrawler_Start_FetcherError(t *testing.T) {
	Now = func() time.Time {
		t, _ := time.Parse("2006-Jan-02", "2013-Feb-03")
		return t
	}

	defer func() {
		Now = time.Now
	}()

	crawler := &Crawler{
		interval: 200 * time.Millisecond,
		fetcher: &mockFetcher{
			err: assert.AnError,
		},
		hub: &Hub{
			dataCh: make(chan []byte, 1),
		},
		services: []*Service{
			{
				Name: "service1",
			},
		},
		widgets: &Widgets{
			Gauges: []*Gauge{
				{
					cid:      "g1",
					Metric:   NewSafeMetric("gauge.metric"),
					MaxValue: 1000,
					Service:  "service1",
				},
			},
			LineCharts: []*LineChart{},
			Texts:      []*Text{},
		},
		done: make(chan struct{}, 1),
	}

	go crawler.Start()
	defer crawler.Stop()

	ch := make(chan bool)

	go func() {
		assert.Equal(t, `{"g":[{"i":"g1","v":0}],"lc":[],"t":[]}`, string(<-crawler.hub.dataCh))

		ch <- true
	}()

	if err := WaitTime(ch, time.Second); err != nil {
		t.Fatal("Did not get response in time")
	}
}

func TestCrawler_Start_FetcherTimeout(t *testing.T) {
	crawler := &Crawler{
		interval: 200 * time.Millisecond,
		fetcher: &mockFetcher{
			timeout: 1200 * time.Millisecond,
		},
		hub: &Hub{
			dataCh: make(chan []byte, 1),
		},
		services: []*Service{
			{
				Name: "service1",
			},
		},
		widgets: &Widgets{
			Gauges: []*Gauge{
				{
					cid:      "g1",
					Metric:   NewSafeMetric("gauge.metric"),
					MaxValue: 1000,
					Service:  "service1",
				},
			},
			LineCharts: []*LineChart{},
			Texts:      []*Text{},
		},
		done: make(chan struct{}, 1),
	}

	go crawler.Start()
	defer crawler.Stop()

	ch := make(chan bool)

	go func() {
		assert.Equal(t, `{"g":[{"i":"g1","v":0}],"lc":[],"t":[]}`, string(<-crawler.hub.dataCh))

		ch <- true
	}()

	if err := WaitTime(ch, 1500*time.Millisecond); err != nil {
		t.Fatal("Did not get response in time")
	}
}

func TestCrawler_ExtractUpdates(t *testing.T) {
	Now = func() time.Time {
		t, _ := time.Parse("2006-Jan-02", "2013-Feb-03")
		return t
	}

	defer func() {
		Now = time.Now
	}()

	o1, err := jason.NewObjectFromBytes([]byte(`{"gauge": {"metric": 800}, "process": {"text": "text 1"}, "memstats": {"alloc": 123}}`))
	assert.NoError(t, err)
	o2, err := jason.NewObjectFromBytes([]byte(`{"gauge": {"metric": 600}, "process": {"text": "text 2"}, "memstats": {"alloc": 456}}`))
	assert.NoError(t, err)

	crawler := &Crawler{
		services: []*Service{
			{
				Name: "service1",
			},
			{
				Name: "service2",
			},
			{
				Name: "service3",
			},
		},
		widgets: &Widgets{
			Gauges: []*Gauge{
				{
					cid:      "g1",
					Metric:   NewSafeMetric("gauge.metric"),
					MaxValue: 1000,
					Service:  "service1",
				},
				{
					cid:      "g2",
					Metric:   NewSafeMetric("gauge.metric"),
					MaxValue: 1000,
					Service:  "service2",
				},
				{
					cid:      "g3",
					Metric:   NewSafeMetric("gauge.metric"),
					MaxValue: 1000,
					Service:  "service3",
				},
			},
			LineCharts: []*LineChart{
				{
					cid:    "lc1",
					Metric: NewSafeMetric("memstats.alloc"),
				},
				{
					cid:      "lc2",
					Metric:   NewSafeMetric("memstats.alloc"),
					Services: []string{"service3"},
				},
				{
					cid:      "lc3",
					Metric:   NewSafeMetric("memstats.alloc"),
					Services: []string{"service2"},
				},
			},
			Texts: []*Text{
				{
					cid:     "t1",
					Metric:  NewSafeMetric("process.text"),
					Service: "service1",
				},
				{
					cid:     "t2",
					Metric:  NewSafeMetric("process.text"),
					Service: "service2",
				},
				{
					cid:     "t3",
					Metric:  NewSafeMetric("process.text"),
					Service: "service3",
				},
			},
		},
	}

	updates := crawler.ExtractUpdates(map[string]*Expvars{
		"service1": {o1},
		"service2": {o2},
	})

	assert.Equal(t, &WidgetsUpdates{
		Gauges: []*GaugeUpdate{
			{
				ID:    "g1",
				Value: 0.8,
			},
			{
				ID:    "g2",
				Value: 0.6,
			},
			{
				ID:    "g3",
				Value: 0.0,
			},
		},
		LineCharts: []*LineChartUpdate{
			{
				ID: "lc1",
				Points: []LinePoint{
					{
						Time: 1359849600,
						Y:    123,
					},
					{
						Time: 1359849600,
						Y:    456,
					},
					{
						Time: 1359849600,
						Y:    0,
					},
				},
			},
			{
				ID: "lc2",
				Points: []LinePoint{
					{
						Time: 1359849600,
						Y:    0,
					},
				},
			},
			{
				ID: "lc3",
				Points: []LinePoint{
					{
						Time: 1359849600,
						Y:    456,
					},
				},
			},
		},
		Texts: []*TextUpdate{
			{
				ID:    "t1",
				Value: "text 1",
			},
			{
				ID:    "t2",
				Value: "text 2",
			},
			{
				ID:    "t3",
				Value: "N/A",
			},
		},
	}, updates)
}

func TestGaugeValue(t *testing.T) {
	m := NewSafeMetric("test.metric")

	tests := []struct {
		name string
		vars string
		want float64
	}{
		{
			name: "read non existing value",
			vars: `{}`,
			want: 0,
		},
		{
			name: "read int64 value",
			vars: `{"test": {"metric": 747}}`,
			want: 0.747,
		},
		{
			name: "read float64 value",
			vars: `{"test": {"metric": 74.7}}`,
			want: 0.0747,
		},
		{
			name: "read bool value",
			vars: `{"test": {"metric": true}}`,
			want: 0,
		},
		{
			name: "read string value",
			vars: `{"test": {"metric": "hello"}}`,
			want: 0,
		},
		{
			name: "read array value",
			vars: `{"test": {"metric": [1,2,3]}}`,
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, err := jason.NewObjectFromBytes([]byte(tt.vars))
			assert.NoError(t, err)

			vars := &Expvars{o}

			if got := GaugeValue(m, 1000, vars); got != tt.want {
				t.Errorf("GaugeValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLineChartValue(t *testing.T) {
	m := NewSafeMetric("test.metric")

	tests := []struct {
		name string
		vars string
		want int64
	}{
		{
			name: "read non existing value",
			vars: `{}`,
			want: 0,
		},
		{
			name: "read int64 value",
			vars: `{"test": {"metric": 747}}`,
			want: 747,
		},
		{
			name: "read float64 value",
			vars: `{"test": {"metric": 36.6}}`,
			want: 0,
		},
		{
			name: "read bool value",
			vars: `{"test": {"metric": true}}`,
			want: 0,
		},
		{
			name: "read string value",
			vars: `{"test": {"metric": "hello"}}`,
			want: 0,
		},
		{
			name: "read array value",
			vars: `{"test": {"metric": [1,2,3]}}`,
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, err := jason.NewObjectFromBytes([]byte(tt.vars))
			assert.NoError(t, err)

			vars := &Expvars{o}

			if got := LineChartValue(m, vars); got != tt.want {
				t.Errorf("GaugeValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTextValue(t *testing.T) {
	m := NewSafeMetric("test.metric")

	tests := []struct {
		name string
		vars string
		want string
	}{
		{
			name: "read non existing value",
			vars: `{}`,
			want: "N/A",
		},
		{
			name: "read int64 value",
			vars: `{"test": {"metric": 747}}`,
			want: "747",
		},
		{
			name: "read float64 value",
			vars: `{"test": {"metric": 36.6}}`,
			want: "36.60",
		},
		{
			name: "read bool value",
			vars: `{"test": {"metric": true}}`,
			want: "true",
		},
		{
			name: "read string value",
			vars: `{"test": {"metric": "hello"}}`,
			want: "hello",
		},
		{
			name: "read array value",
			vars: `{"test": {"metric": [1,2,3]}}`,
			want: "N/A",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, err := jason.NewObjectFromBytes([]byte(tt.vars))
			assert.NoError(t, err)

			vars := &Expvars{o}

			if got := TextValue(m, vars); got != tt.want {
				t.Errorf("GaugeValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadMetric(t *testing.T) {
	m := NewSafeMetric("test.metric")

	tests := []struct {
		name string
		vars string
		want interface{}
	}{
		{
			name: "read non existing value",
			vars: `{}`,
			want: nil,
		},
		{
			name: "read int64 value",
			vars: `{"test": {"metric": 747}}`,
			want: int64(747),
		},
		{
			name: "read float64 value",
			vars: `{"test": {"metric": 36.6}}`,
			want: float64(36.6),
		},
		{
			name: "read bool value",
			vars: `{"test": {"metric": true}}`,
			want: true,
		},
		{
			name: "read string value",
			vars: `{"test": {"metric": "hello"}}`,
			want: "hello",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, err := jason.NewObjectFromBytes([]byte(tt.vars))
			assert.NoError(t, err)

			vars := &Expvars{o}

			if got := ReadMetric(m, vars); got != tt.want {
				t.Errorf("GaugeValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
