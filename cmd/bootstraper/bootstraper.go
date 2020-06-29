package main

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tritonmedia/bootstraper/pkg/codegen"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

func main() {
	ctx := context.Background()
	log := logrus.New().WithContext(ctx)

	app := cli.App{
		Version: "1.0.0",
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

			gfs := memfs.New()
			branch := "master"
			if dev {
				// Setting branch to "" causes us to use local templates instead
				branch = ""
			} else {
				log.Info("cloning bootstraper repo ...")
				_, err = git.Clone(memory.NewStorage(), gfs, &git.CloneOptions{
					URL:           "https://github.com/tritonmedia/bootstraper.git",
					ReferenceName: plumbing.ReferenceName("refs/heads/" + branch),
					SingleBranch:  true,
					Depth:         1,
				})
				if err != nil {
					return errors.Wrap(err, "failed to clone bootstraper repository")
				}
			}

			r := codegen.NewRenderer(log, branch, cwd, m, gfs)
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
	app.Commands = []*cli.Command{
		{
			Name: "generate",
			Action: func(c *cli.Context) error {
				cwd, err := os.Getwd()
				if err != nil {
					return errors.Wrap(err, "failed to get the current working directory")
				}

				m := codegen.ServiceManifest{
					Name:      filepath.Base(cwd),
					Type:      "",
					Arguments: make(map[string]interface{}),
				}
				b, err := yaml.Marshal(m)
				if err != nil {
					return err
				}

				err = ioutil.WriteFile(filepath.Join(cwd, "service.yaml"), b, 0644)
				if err != nil {
					return err
				}

				log.Info("generated service.yaml in current directory")
				return nil
			},
		},
		{
			Name: "generate-templatelist",
			Action: func(c *cli.Context) error {
				cwd, err := os.Getwd()
				if err != nil {
					return errors.Wrap(err, "failed to get the current working directory")
				}

				templatePath := filepath.Join(cwd, "pkg/codegen/templates")
				filesYaml := filepath.Join(templatePath, "files.yaml.tpl")

				if _, err := os.Stat(templatePath); os.IsNotExist(err) {
					return errors.Wrap(err, "must be run in root of bootstraper repository")
				} else if err != nil {
					return err
				}

				tl := codegen.TemplateList{
					Templates: make(map[string]*codegen.Template),
				}

				// attempt to read in the current, existing, files.yaml
				if b, err := ioutil.ReadFile(filesYaml); err == nil {
					if err := yaml.Unmarshal(b, &tl); err != nil {
						return errors.Wrap(err, "failed to parse existing files.yaml")
					}
				}

				// TODO(jaredallard): eventually support removing files.
				err = filepath.Walk(templatePath, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}

					// skip non .tpl files
					if !strings.HasSuffix(path, ".tpl") {
						return nil
					}

					// we don't care about directories
					if info.IsDir() {
						return nil
					}

					path, err = filepath.Rel(templatePath, path)
					if err != nil {
						return err
					}

					// skip second run
					if path == "files.yaml.tpl" {
						return nil
					}

					writePath := strings.TrimSuffix(path, ".tpl")
					if _, ok := tl.Templates[writePath]; !ok {
						tl.Templates[writePath] = &codegen.Template{
							Source: path,
						}
					}

					return nil
				})
				if err != nil {
					return err
				}

				b, err := yaml.Marshal(tl)
				if err != nil {
					return err
				}

				err = ioutil.WriteFile(filesYaml, b, 0644)
				if err != nil {
					return err
				}

				log.Infof("generated files.yaml.tpl at %s", filesYaml)
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Errorf("failed to run: %v", err)
		os.Exit(1)
	}
}
