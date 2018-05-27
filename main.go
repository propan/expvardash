package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"html/template"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
)

var Now = time.Now

var upgrader = websocket.Upgrader{}

var (
	interval  = flag.Duration("i", 5*time.Second, "Polling interval: 5s, 1m")
	port      = flag.Int("p", 4444, "Dashboard HTTP port")
	dashboard = flag.String("d", "", "Dashboard configuration file")
	fs        = flag.Bool("fs", false, "Serve static files from file system")
)

func main() {
	flag.Usage = Usage
	flag.Parse()

	if *interval <= 0 {
		fmt.Fprintln(os.Stderr, "Invalid polling interval.")
		Usage()
		os.Exit(1)
	}

	// Load configuration file
	conf, err := LoadConf(*dashboard)
	if err != nil {
		fmt.Println("Could not read dashboard configuration:", err)
		os.Exit(1)
	}

	// Start handler for web-socket connections
	hub := NewHub()
	go hub.Start()

	fetcher := NewFetcher()

	crawler := &Crawler{
		interval: *interval,
		fetcher:  fetcher,
		hub:      hub,
		widgets:  conf.Widgets,
		services: conf.Services,
	}
	go crawler.Start()

	err = ListenAndServe(*port, hub, conf.Layout, *fs)
	if err != nil {
		fmt.Println("Could not start HTTP server:", err)
		os.Exit(1)
	}
}

func LoadTemplate(fsMode bool) (*template.Template, error) {
	if fsMode {
		return template.ParseFiles("templates/index.html")
	} else {
		data, err := Asset("templates/index.html")
		if err != nil {
			return nil, err
		}

		return template.New("templates/index.html").Parse(string(data))
	}
}
func ListenAndServe(port int, hub *Hub, layout *Layout, fsMode bool) error {
	t, err := LoadTemplate(fsMode)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf(":%d", port)

	fmt.Printf("Starting HTTP server on localhost%s\n", addr)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		err := t.Execute(w, map[string]interface{}{"Port": port, "Layout": *layout})
		if err != nil {
			fmt.Println("Error rendering response:", err)
		}
	})
	if fsMode {
		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	} else {
		http.Handle("/static/", http.FileServer(&assetfs.AssetFS{
			Asset:     Asset,
			AssetDir:  AssetDir,
			AssetInfo: AssetInfo,
			Prefix:    "",
		}))
	}

	http.HandleFunc("/updates", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Could not upgrade:", err)
			return
		}

		client := &Client{
			hub:    hub,
			conn:   conn,
			dataCh: make(chan []byte, 10),
		}

		defer func() {
			client.hub.leaveCh <- client
			client.conn.Close()
		}()

		hub.enterCh <- client

		for {
			err := client.conn.WriteMessage(websocket.TextMessage, <-client.dataCh)
			if err != nil {
				fmt.Println("Could not send:", err)
				break
			}
		}
	})

	return http.ListenAndServe(addr, nil)
}

func Usage() {
	progname := os.Args[0]
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", progname)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
Examples:
	%s -d=dashboard.json
	%s -d=dashboard.json -i=10s

For more details and docs, see README: http://github.com/propan/expvardash
`, progname, progname)
}
