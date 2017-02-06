package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

type LinePoint struct {
	Time int64 `json:"time"`
	Y    int64 `json:"y"`
}

type LineChartUpdate struct {
	ID     string      `json:"i"`
	Points []LinePoint `json:"p"`
}

type GaugeUpdate struct {
	ID    string  `json:"i"`
	Value float64 `json:"v"`
}

type TextUpdate struct {
	ID    string `json:"i"`
	Value string `json:"v"`
}

type WidgetsUpdates struct {
	Gauges     []*GaugeUpdate     `json:"g"`
	LineCharts []*LineChartUpdate `json:"lc"`
	Texts      []*TextUpdate      `json:"t"`
}

type Crawler struct {
	interval time.Duration
	fetcher  *Fetcher
	hub      *Hub
	services []*Service
	widgets  *Widgets
}

type result struct {
	service string
	vars    *Expvars
}

func (c *Crawler) Start() {
	tick := time.NewTicker(c.interval)
	for {
		<-tick.C

		updates := c.ExtractUpdates(c.fetchAll())
		data, err := json.Marshal(updates)
		if err != nil {
			fmt.Println("Error serializing response:", err)
			continue
		}

		c.hub.dataCh <- data
	}
}

func (c *Crawler) fetchAll() map[string]*Expvars {
	vars := map[string]*Expvars{}

	resCh := make(chan result, len(c.services))

	for _, service := range c.services {
		service := service
		go func() {
			vars, err := c.fetcher.Fetch(service.URL)
			if err != nil {
				fmt.Printf("Failed to crawl '%s': %s\n", service.Name, err)
			}
			resCh <- result{service: service.Name, vars: vars}
		}()
	}

	timeout := time.After(time.Second)

	for i := 0; i < len(c.services); i++ {
		select {
		case <-timeout:
			fmt.Println("Timed out waiting for all crawling results")
			return vars
		case r := <-resCh:
			if r.vars != nil {
				vars[r.service] = r.vars
			}
		}
	}

	return vars
}

func (c *Crawler) ExtractUpdates(vars map[string]*Expvars) *WidgetsUpdates {
	u := &WidgetsUpdates{
		Gauges:     []*GaugeUpdate{},
		LineCharts: []*LineChartUpdate{},
		Texts:      []*TextUpdate{},
	}

	now := Now().Unix()

	for _, g := range c.widgets.Gauges {
		u.Gauges = append(u.Gauges, &GaugeUpdate{
			ID:    g.ID(),
			Value: GaugeValue(g.Metric, g.MaxValue, vars[g.Service]),
		})
	}

	for _, ch := range c.widgets.LineCharts {
		lu := &LineChartUpdate{
			ID:     ch.ID(),
			Points: []LinePoint{},
		}
		if len(ch.Services) > 0 {
			for _, s := range ch.Services {
				lu.Points = append(lu.Points, LinePoint{
					Time: now,
					Y:    LineChartValue(ch.Metric, vars[s]),
				})
			}
		} else {
			for _, s := range c.services {
				lu.Points = append(lu.Points, LinePoint{
					Time: now,
					Y:    LineChartValue(ch.Metric, vars[s.Name]),
				})
			}
		}
		u.LineCharts = append(u.LineCharts, lu)
	}

	for _, t := range c.widgets.Texts {
		u.Texts = append(u.Texts, &TextUpdate{
			ID:    t.ID(),
			Value: TextValue(t.Metric, vars[t.Service]),
		})
	}

	return u
}

func GaugeValue(m *Metric, max int64, vars *Expvars) float64 {
	if vars == nil {
		return 0.0
	}

	v := ReadMetric(m, vars)
	if value, ok := v.(int64); ok {
		return float64(value) / float64(max)
	}
	if value, ok := v.(float64); ok {
		return value / float64(max)
	}

	fmt.Printf("%s: usage of %s with gauge is not supported\n", m, reflect.TypeOf(v))

	return 0.0
}

func LineChartValue(m *Metric, vars *Expvars) int64 {
	if vars == nil {
		return 0
	}

	v := ReadMetric(m, vars)
	if value, ok := v.(int64); ok {
		return value
	}

	fmt.Printf("%s: usage of %s with line-chart is not supported\n", m, reflect.TypeOf(v))

	return 0
}

func TextValue(m *Metric, vars *Expvars) string {
	if vars == nil {
		return "N/A"
	}

	v := ReadMetric(m, vars)
	if value, ok := v.(int64); ok {
		return fmt.Sprintf("%d", value)
	} else if value, ok := v.(float64); ok {
		return fmt.Sprintf("%.2f", value)
	} else if value, ok := v.(bool); ok {
		return fmt.Sprintf("%t", value)
	} else if value, ok := v.(string); ok {
		return value
	}

	fmt.Printf("%s: usage of %s with gauge is not supported\n", m, reflect.TypeOf(v))

	return "N/A"
}

func ReadMetric(m *Metric, vars *Expvars) interface{} {
	value, err := vars.GetValue(m.Path...)
	if err != nil {
		return nil
	}

	if v, err := value.Int64(); err == nil {
		return v
	} else if v, err := value.Float64(); err == nil {
		return v
	} else if v, err := value.Boolean(); err == nil {
		return v
	} else if v, err := value.String(); err == nil {
		return v
	} else if v, err := value.Array(); err == nil {
		return v
	}

	return nil
}
