package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

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

	err = ListenAndServe(fmt.Sprintf(":%d", *port), conf.Layout)
	if err != nil {
		fmt.Println("Could not start HTTP server:", err)
		os.Exit(1)
	}
}

func ListenAndServe(addr string, layout *Layout) error {
	fmt.Printf("Starting HTTP server on localhost%s\n", addr)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := layout.RenderTo(w)
		if err != nil {
			fmt.Println("Error rendering response:", err)
		}
	})
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

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
