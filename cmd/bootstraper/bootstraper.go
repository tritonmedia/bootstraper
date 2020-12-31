package main

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/tritonmedia/bootstraper/pkg/codegen"
	"github.com/tritonmedia/pkg/app"
)

// Why: Refactor later.
//nolint:funlen,gocyclo
func main() {
	ctx := context.Background()
	log := logrus.New().WithContext(ctx)

	//nolint:gocritic importShadow
	app := cli.App{
		Version: app.Version,
		Name:    "bootstraper",
		Action: func(c *cli.Context) error {
			dev := c.Bool("dev")

			cwd, err := os.Getwd()
			if err != nil {
				return errors.Wrap(err, "failed to get the current working directory")
			}

			b, err := ioutil.ReadFile(filepath.Join(cwd, "service.yaml"))
			if err != nil {
				log.Info("A service.yaml can be generated with 'bootstraper generate'")
				return errors.Wrap(err, "failed to read service.yaml")
			}

			var m *codegen.ServiceManifest
			err = yaml.Unmarshal(b, &m)
			if err != nil {
				return errors.Wrap(err, "failed to parse service.yaml")
			}

			firstInit := false
			_, err = git.PlainOpen(cwd)
			if err != nil {
				log.Info("running 'git init'")
				_, err = git.PlainInit(cwd, false)
				if err != nil {
					return errors.Wrap(err, "failed to initialize git repository")
				}

				firstInit = true
			}

			branch := "master"
			if dev {
				// Setting branch to "" causes us to use local templates instead
				branch = ""
			}

			r := codegen.NewRenderer(log, branch, cwd, m)
			err = r.Render(ctx, log)
			if err != nil {
				return errors.Wrap(err, "failed to run bootstraper")
			}

			if !firstInit {
				return nil
			}

			// TODO(jaredallard): create an initial commit
			return nil
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "dev",
				Usage: "Use local manifests instead of remote ones, useful for development",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Errorf("failed to run: %v", err)
		os.Exit(1)
	}
}
