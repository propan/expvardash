package main

import (
	"errors"
	"fmt"
)

const (
	LineChartType = "LineChart"
)

type Chart interface {
	ID() string
	SetID(string)
}

type LineChart struct {
	cid      string   `json:"-"`
	Metric   string   `json:"metric"`
	Services []string `json:"services"`
}

func (c *LineChart) ID() string {
	return c.cid
}

func (c *LineChart) SetID(id string) {
	c.cid = id
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
