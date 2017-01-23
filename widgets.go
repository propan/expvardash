package main

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	GaugeType     = "Gauge"
	LineChartType = "LineChart"
	TextType      = "Text"
)

type Widget interface {
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

type Text struct {
	cid        string  `json:"-"`
	Metric     *Metric `json:"-"`
	MetricName string  `json:"metric"`
	Service    string  `json:"service"`
}

func (t *Text) ID() string {
	return t.cid
}

func (t *Text) SetID(id string) {
	t.cid = id
}

func (t *Text) Title() string {
	return t.Metric.String()
}

func (t *Text) HasLegend() bool {
	return false
}

func (t *Text) Series() []string {
	return []string{}
}

type Widgets struct {
	nextID     int
	Gauges     []*Gauge
	LineCharts []*LineChart
	Texts      []*Text
}

func (ww *Widgets) NextID() string {
	ww.nextID++
	return fmt.Sprintf("c%d", ww.nextID)
}

func (ww *Widgets) Append(widget Widget) error {
	switch c := widget.(type) {
	case *Gauge:
		ww.Gauges = append(ww.Gauges, c)
		return nil
	case *LineChart:
		ww.LineCharts = append(ww.LineCharts, c)
		return nil
	case *Text:
		ww.Texts = append(ww.Texts, c)
		return nil
	default:
		return fmt.Errorf("Unknown widget type: %s", reflect.TypeOf(widget))
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
