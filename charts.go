package main

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	GaugeType     = "Gauge"
	LineChartType = "LineChart"
)

type Chart interface {
	ID() string
	SetID(string)
	Title() string
	HasLegend() bool
	Series() []string
}

type LineChart struct {
	cid        string   `json:"-"`
	Metric     *Metric  `json:"-"`
	MetricName string   `json:"metric"`
	ShowLegend *bool    `json:"show_legend"`
	Services   []string `json:"services"`
}

func (c *LineChart) ID() string {
	return c.cid
}

func (c *LineChart) SetID(id string) {
	c.cid = id
}

func (c *LineChart) Title() string {
	return c.Metric.String()
}

func (c *LineChart) HasLegend() bool {
	if c.ShowLegend == nil {
		return true
	}
	return *c.ShowLegend
}

func (c *LineChart) Series() []string {
	return c.Services
}

type Gauge struct {
	cid        string  `json:"-"`
	Metric     *Metric `json:"-"`
	MetricName string  `json:"metric"`
	Service    string  `json:"service"`
	MaxValue   int64   `json:"max"`
}

func (g *Gauge) ID() string {
	return g.cid
}

func (g *Gauge) SetID(id string) {
	g.cid = id
}

func (g *Gauge) Title() string {
	return g.Metric.String()
}

func (g *Gauge) HasLegend() bool {
	return false
}

func (g *Gauge) Series() []string {
	return []string{}
}

type Charts struct {
	nextID     int
	Gauges     []*Gauge
	LineCharts []*LineChart
}

func (cc *Charts) NextID() string {
	cc.nextID++
	return fmt.Sprintf("c%d", cc.nextID)
}

func (cc *Charts) Append(chart Chart) error {
	switch c := chart.(type) {
	case *LineChart:
		cc.LineCharts = append(cc.LineCharts, c)
		return nil
	case *Gauge:
		cc.Gauges = append(cc.Gauges, c)
		return nil
	default:
		return fmt.Errorf("Unknown chart type: %s", reflect.TypeOf(chart))
	}
}

type Metric struct {
	Path []string
}

func NewMetric(name string) (*Metric, error) {
	return &Metric{
		Path: strings.Split(name, "."),
	}, nil
}

func (m *Metric) String() string {
	return strings.Join(m.Path, ".")
}
