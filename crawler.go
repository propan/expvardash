package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

type LinePoint struct {
	Time int64 `json:"time"`
	Y    int   `json:"y"`
}

type LineChartUpdate struct {
	ID     string      `json:"i"`
	Points []LinePoint `json:"p"`
}

type ChartsUpdates struct {
	LineCharts []*LineChartUpdate `json:"lc"`
}

type Crawler struct {
	interval time.Duration
	fetcher  *Fetcher
	hub      *Hub
	services []*Service
	charts   *Charts
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
				return
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
			vars[r.service] = r.vars
		}
	}

	return vars
}

func (c *Crawler) ExtractUpdates(vars map[string]*Expvars) *ChartsUpdates {
	u := &ChartsUpdates{
		LineCharts: []*LineChartUpdate{},
	}

	now := time.Now().Unix()

	for _, ch := range c.charts.LineCharts {
		lu := &LineChartUpdate{
			ID:     ch.ID(),
			Points: []LinePoint{},
		}
		if len(ch.Services) > 0 {
			for range ch.Services {
				lu.Points = append(lu.Points, LinePoint{Time: now, Y: rand.Intn(10)})
			}
		} else {
			for range c.services {
				lu.Points = append(lu.Points, LinePoint{Time: now, Y: rand.Intn(10)})
			}
		}
		u.LineCharts = append(u.LineCharts, lu)
	}

	return u
}
