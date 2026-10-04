package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v3"

	"github.com/nikimonax/laretz/internal/server"
	"github.com/nikimonax/laretz/internal/server/config"
)

func Run() {
	if err := Build().Run(context.Background(), os.Args); err != nil {
		HandleError(err)
	}
}

func Build() *cli.Command {
	return &cli.Command{
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "open",
				Aliases: []string{"o"},
				Value:   false,
				Usage:   "open browser after server starts",
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "",
				Usage:   "custom config file path",
			},
		},
		Action: actionRunApp,
		ExitErrHandler: func(_ context.Context, _ *cli.Command, err error) {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		},
	}
}

func HandleError(err error) {
	var exitCoder cli.ExitCoder

	if errors.As(err, &exitCoder) {
		os.Exit(exitCoder.ExitCode())
	}

	os.Exit(1)
}

func actionRunApp(_ context.Context, cmd *cli.Command) error {
	// load .env
	if err := godotenv.Load(); err != nil {
		return err
	}

	// determine config path
	configPath := cmd.String("config")

	if configPath == "" {
		configPath = os.Getenv("LARETZ_CONFIG")
	}

	if configPath == "" {
		configPath = config.DefaultConfigPath()
	}

	// read config
	cfg, err := config.EnsureConfig(configPath)

	if err != nil {
		return err
	}

	if cmd.Bool("open") {
		cfg.Server.OpenBrowser = true
	}

	app, err := server.New(cfg)

	if err != nil {
		return err
	}

	if err := app.Err(); err != nil {
		return err
	}

	app.Run()
	return nil
}
