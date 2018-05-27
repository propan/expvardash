# ExpvarDash

a monitoring system for Go applications using [expvar](https://golang.org/pkg/expvar/) (/debug/vars)

This is not a solid monitoring solution, but rather a tool that allows you to see quickly the status of the services you develop.
# Example

![example dashboard](https://github.com/propan/expvardash/raw/master/screenshot.png)

# Installation

Build the application by running:

```bash
make install
```

# Usage

```bash
expvardash -d dashboard.json
```

## Getting Help

```bash
expvardash --help
```

## Configuration

Before you can start using the application, you have to configure the dashboard you want to see and the services you want to monitor.

The application configuration is a simple JSON file that consists of a list of services to monitor and a dashboard layout:

The services are defined using `services` block:

```json
{
  "services": [
    {
      "name": "service-1",
      "url": "localhost:4004"
    },
    {
      "name": "service-2",
      "url": "http://localhost:4005/debug/vars"
    }
  ]
}
```

- **name** - an identifier that is used when you refer to the service in the dashboard configuration
- **url** - a HTTP-endpoint that exposes service's [expvar](https://golang.org/pkg/expvar/)

The dashboard layout is configured with `rows` block. A dashboard consists of rows and each row consists of blocks. The following block types are supported:

- Text
- Gauge
- LineChart

Each block allows you to specify its `size` in units and `title`. A row has width of 12 units.

#### Text Block

Text block allows you to visualize current value of a variable as it is.

```json
{
    "type": "Text",
    "title": "Req/Sec",
    "size": 2,
    "conf": {
      "service": "service-1",
      "metric": "node.RequestPerSecond"
    }
}
```

Configuration:

- **service** - an identifier of the service
- **metric** - a metric to visualize

#### Gauge Block

Gauge block can be used to visualize a variable which value changes within a known range.

```json
{
    "type": "Gauge",
    "title": "In-Flight",
    "size": 2,
    "conf": {
      "service": "service-2",
      "metric": "node.RequestsInflight",
      "max": 200
    }
}
```

Configuration:

- **service** - an identifier of the service
- **metric** - a metric to visualize
- **max** - the maximum value of the metric

#### Line Chart Block

Line Chart can be used to visualize changes of the variable values over time.

```json
{
    "type": "LineChart",
    "title": "Service #1: Heap Usage",
    "size": 6,
    "conf": {
        "metric": "memstats.HeapAlloc",
        "show_legend": false,
        "services": [
          "service-1"
        ]
    }
}
```

Configuration:

- **metric** - a metric to visualize
- **show_legend** - a flag that controls whether the chart legend is visible or not
- **services** - identifiers of the services to be included on the chart. If omitted, all services are included. 

### Example

```json
{
   "services": [
    {
      "name": "service-1",
      "url": "localhost:4004"
    },
    {
      "name": "service-2",
      "url": "http://localhost:4005/debug/vars"
    }
   ],
   "rows": [
     {
       "items": [
         {
           "type": "Gauge",
           "title": "Service #1: In-Flight",
           "size": 2,
           "conf": {
             "service": "service-1",
             "metric": "node.RequestsInflight",
             "max": 200
           }
         },
         {
           "type": "Gauge",
           "title": "Service-2: In-Flight",
           "size": 2,
           "conf": {
             "service": "service-2",
             "metric": "node.RequestsInflight",
             "max": 200
           }
         },
         {
           "type": "LineChart",
           "title": "Goroutines",
           "size": 8,
           "conf": {
             "metric": "Goroutines"
           }
         }
       ]
     },
     {
       "items": [
         {
           "type": "LineChart",
           "title": "Memory Usage",
           "size": 6,
           "conf": {
             "metric": "memstats.Alloc"
           }
         },
         {
           "type": "LineChart",
           "title": "Service #1: Heap Usage",
           "size": 6,
           "conf": {
             "metric": "memstats.HeapAlloc",
             "show_legend": false,
             "services": [
               "service-1"
             ]
           }
         }
       ]
     },
     {
       "items": [
         {
           "type": "Text",
           "title": "Service #1: Req/Sec",
           "size": 2,
           "conf": {
             "service": "service-1",
             "metric": "node.RequestPerSecond"
           }
         },
         {
           "type": "LineChart",
           "title": "Request per Second",
           "size": 10,
           "conf": {
             "metric": "node.RequestPerSecond"
           }
         }
       ]
     }
   ]
}
```

# License

Copyright © 2017-2018 Pavel Prokopenko

Distributed under the Eclipse Public License either version 1.0 or (at your option) any later version.
