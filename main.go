package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

var (
	port = flag.Int("p", 4444, "Dashboard HTTP port")
	dashboard = flag.String("d", "", "Dashboard configuration fail")
)

func main() {
	flag.Usage = Usage
	flag.Parse()

	// Load configuration file
	conf, err := LoadConf(*dashboard)
	if err != nil {
		fmt.Println("Could not read dashboard configuration:", err)
		os.Exit(1)
	}

	// Start handler for web-socket connections
	hub := NewHub()
	go hub.Start()

	err = ListenAndServe(fmt.Sprintf(":%d", *port), hub, conf.Layout)
	if err != nil {
		fmt.Println("Could not start HTTP server:", err)
		os.Exit(1)
	}
}

func ListenAndServe(addr string, hub *Hub, layout *Layout) error {
	fmt.Printf("Starting HTTP server on localhost%s\n", addr)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := layout.RenderTo(w)
		if err != nil {
			fmt.Println("Error rendering response:", err)
		}
	})
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

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
	%s -dashboard=dashboard.json

For more details and docs, see README: http://github.com/propan/expvardash
`, progname)
}
