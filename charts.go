package main

import (
	"errors"
	"fmt"
	"strings"
)

const (
	LineChartType = "LineChart"
)

type Chart interface {
	ID() string
	SetID(string)
	Title() string
}

type LineChart struct {
	cid        string   `json:"-"`
	MetricName string   `json:"metric"`
	Metric     *Metric  `json:"-"`
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

type Charts struct {
	nextID     int
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
	default:
		return errors.New("Unknown chart type")
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
