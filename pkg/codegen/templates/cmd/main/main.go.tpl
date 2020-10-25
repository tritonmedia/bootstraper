package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/tritonmedia/pkg/app"
	"github.com/tritonmedia/pkg/service"
	"github.com/urfave/cli/v2"

	"github.com/tritonmedia/{{ .manifest.Name }}/internal/converter"
	"github.com/tritonmedia/{{ .manifest.Name }}/internal/api"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	log := logrus.New().WithContext(ctx)

	app := cli.App{
		Name:    "{{ .manifest.Name }}",
		Version: app.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "log-format",
				Usage: "set the formatter for logs",
				EnvVars: []string{"LOG_FORMAT"},
				Value: "JSON",
			},
		},
	}
	app.Action = func(c *cli.Context) error {
		logFormat := c.String("log-format")
		switch strings.ToLower(logFormat) {
		case "json":
			logrus.SetFormatter(&logrus.JSONFormatter{})
		case "text":
			logrus.SetFormatter(&logrus.TextFormatter{})
		default:
			return fmt.Errorf("unknown log format: %s", logFormat)
		}

		// update our logger's formatter, since we created before
		// we set it
		log.Logger.Formatter = logrus.StandardLogger().Formatter

		// start the service runner, which handles context cancellation
		// and threading
		r := service.NewServiceRunner(ctx, []service.Service{
			{{- if eq .manifest.Type "JobProcessor" }}
			&{{ .manifest.Name }}.ConsumerService{},
			{{- else if eq .manifest.Type "GRPC" }}
			api.NewGRPCService(),
			{{- end }}
		})
		sigC := make(chan os.Signal)

		// listen for signals that we want to cancel on, and cancel
		// the context if one is passed
		signal.Notify(sigC, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigC
			cancel()
		}()

		return r.Run(ctx, log)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("failed to run: %v", err)
	}
}
