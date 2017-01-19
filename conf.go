package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"io/ioutil"
)

var Template = template.Must(template.ParseFiles("templates/index.html"))

func LoadConf(path string) (*RawConfig, error) {
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
	Rows []RawRow `json:"rows"`
}

type RawRow struct {
	Items []RawItem `json:"items"`
}

type RawItem struct {
	Type string           `json:"type"`
	Size int              `json:"size"`
	Conf *json.RawMessage `json:"conf"`
}

type Layout struct {
	Rows []*Row
}

type Row struct {
	Cols []*Col
}

type Col struct {
	ID   string
	Size int
}

func (c *RawConfig) ReadLayout() (*Layout, error) {
	layout := &Layout{
		Rows: []*Row{},
	}

	for _, row := range c.Rows {
		cols := []*Col{}

		for _, item := range row.Items {
			cols = append(cols, &Col{
				Size: item.Size,
			})
		}

		layout.Rows = append(layout.Rows, &Row{
			Cols: cols,
		})
	}

	return layout, nil
}

func (l *Layout) Render() (string, error) {
	buf := new(bytes.Buffer)

	err := l.RenderTo(buf)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (l *Layout) RenderTo(w io.Writer) error {
	return Template.Execute(w, *l)
}
