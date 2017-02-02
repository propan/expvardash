package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func LoadConf(path string) (*Config, error) {
	raw, err := ReadConf(path)
	if err != nil {
		return nil, err
	}

	return raw.ParseConf()
}

func ReadConf(path string) (*RawConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var conf RawConfig
	err = json.Unmarshal(data, &conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

type RawConfig struct {
	Services []RawService `json:"services"`
	Rows     []RawRow     `json:"rows"`
}

type RawService struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type RawRow struct {
	Items []RawItem `json:"items"`
}

type RawItem struct {
	Type  string           `json:"type"`
	Title string           `json:"title"`
	Size  int              `json:"size"`
	Conf  *json.RawMessage `json:"conf"`
}

type Config struct {
	Services []*Service
	Layout   *Layout
	Widgets  *Widgets
}

type Layout struct {
	Rows []*Row
}

type Row struct {
	Cols []*Col
}

type Col struct {
	ID     string
	Title  string
	Size   int
	Legend bool
	Series []string
}

func (c *RawConfig) ParseConf() (*Config, error) {
	config := &Config{
		Services: []*Service{},
		Layout: &Layout{
			Rows: []*Row{},
		},
		Widgets: &Widgets{},
	}

	defaultSeries := []string{}

	for _, raw := range c.Services {
		s, err := ReadService(raw)
		if err != nil {
			return nil, err
		}
		config.Services = append(config.Services, s)
		defaultSeries = append(defaultSeries, s.Name)
	}

	for _, row := range c.Rows {
		cols := []*Col{}

		for _, item := range row.Items {
			c, err := ReadChart(item)
			if err != nil {
				return nil, err
			}

			c.SetID(config.Widgets.NextID())

			err = config.Widgets.Append(c)
			if err != nil {
				return nil, err
			}

			title := item.Title
			if len(item.Title) == 0 {
				title = c.Title()
			}

			var series []string
			if c.HasLegend() {
				series = c.Series()
				if len(series) == 0 {
					series = defaultSeries
				}
			}

			cols = append(cols, &Col{
				ID:     c.ID(),
				Title:  title,
				Size:   item.Size,
				Legend: c.HasLegend(),
				Series: series,
			})
		}

		config.Layout.Rows = append(config.Layout.Rows, &Row{
			Cols: cols,
		})
	}

	return config, nil
}

func ReadService(raw RawService) (*Service, error) {
	url, err := ParseURL(raw.URL)
	if err != nil {
		return nil, err
	}

	return &Service{
		Name: raw.Name,
		URL:  *url,
	}, nil
}

func ReadChart(item RawItem) (Widget, error) {
	if item.Conf == nil {
		return nil, fmt.Errorf("Missing configuration for: %s", item.Type)
	}

	switch item.Type {
	case GaugeType:
		return ReadGauge(item.Conf)
	case LineChartType:
		return ReadLineChart(item.Conf)
	case TextType:
		return ReadText(item.Conf)
	default:
		return nil, fmt.Errorf("Unknown widget type: %s", item.Type)
	}
}

func ReadLineChart(data *json.RawMessage) (*LineChart, error) {
	var widget LineChart
	err := json.Unmarshal(*data, &widget)
	if err != nil {
		return nil, err
	}

	metric, err := NewMetric(widget.MetricName)
	if err != nil {
		return nil, err
	}
	widget.Metric = metric

	return &widget, nil
}

func ReadGauge(data *json.RawMessage) (*Gauge, error) {
	var widget Gauge
	err := json.Unmarshal(*data, &widget)
	if err != nil {
		return nil, err
	}

	metric, err := NewMetric(widget.MetricName)
	if err != nil {
		return nil, err
	}
	widget.Metric = metric

	return &widget, nil
}

func ReadText(data *json.RawMessage) (*Text, error) {
	var widget Text
	err := json.Unmarshal(*data, &widget)
	if err != nil {
		return nil, err
	}

	metric, err := NewMetric(widget.MetricName)
	if err != nil {
		return nil, err
	}
	widget.Metric = metric

	return &widget, nil
}
