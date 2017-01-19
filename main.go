package main

import (
	"flag"
	"os"
	"fmt"
)

var (
	dashboard = flag.String("d", "", "Dashboard configuration faile")
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

	layout, err := conf.ReadLayout()
	if err != nil {
		fmt.Println("Could not read dashboard layout:", err)
		os.Exit(1)
	}

	fmt.Println(layout.Render())
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
