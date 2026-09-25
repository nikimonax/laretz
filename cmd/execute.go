package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/nikimonax/laretz/internal/server"
)

func Execute() {
	var opts server.RunOptions
	flag.IntVar(&opts.Port, "port", 0, "TCP port to listen on (0 selects a free port)")
	flag.BoolVar(&opts.OpenBrowser, "open", false, "open the app URL in the system browser")
	flag.BoolVar(&opts.OpenBrowser, "o", false, "shorthand for -open")
	flag.Parse()

	if err := server.Run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "laretz:", err)
		os.Exit(1)
	}
}
