package main

import (
	"testing"

	"time"

	"github.com/antonholmquist/jason"
	"github.com/stretchr/testify/assert"
)

func NewSafeMetric(name string) *Metric {
	m, _ := NewMetric(name)
	return m
}

func TestCrawler_ExtractUpdates(t *testing.T) {
	Now = func() time.Time {
		t, _ := time.Parse("06/11/05, 02:04PM", "01/24/13, 11:27PM")
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
						Time: -62135596800,
						Y:    123,
					},
					{
						Time: -62135596800,
						Y:    456,
					},
					{
						Time: -62135596800,
						Y:    0,
					},
				},
			},
			{
				ID: "lc2",
				Points: []LinePoint{
					{
						Time: -62135596800,
						Y:    0,
					},
				},
			},
			{
				ID: "lc3",
				Points: []LinePoint{
					{
						Time: -62135596800,
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
