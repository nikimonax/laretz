package main

import (
	"log"

	"github.com/nikimonax/laretz/cmd/laretz/cli"
	"github.com/nikimonax/laretz/internal/impl/plugin"
	"github.com/nikimonax/laretz/plugin/coingecko"
)

func init() {
	if err := plugin.DefaultRegistry().Register(coingecko.NewPlugin()); err != nil {
		log.Fatal(err)
	}
}

func main() {
	cli.Run()
}
