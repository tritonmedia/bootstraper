// package codegen handles rendering bootstraper templates
package codegen

import (
	"bufio"
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/go-git/go-billy/v5"
	"golang.org/x/tools/imports"
)

var (
	blockRX = regexp.MustCompile(`[\w]*(///|###)\s*([a-zA-Z]+)\(([a-zA-Z]+)\)`)
)

type Renderer struct {
	branch string
	dir    string
	m      *ServiceManifest
	gfs    billy.Filesystem

	log *logrus.Entry
}

// NewRenderer creates a new template renderer that is the heart of bootstraper.
func NewRenderer(log *logrus.Entry, branch, dir string, m *ServiceManifest, gfs billy.Filesystem) *Renderer {
	return &Renderer{
		branch: branch,
		dir:    dir,
		m:      m,
		gfs:    gfs,
		log:    log,
	}
}

// Render starts the code generation process
func (r *Renderer) Render(ctx context.Context, log *logrus.Entry) error {
	list := &TemplateList{
		Templates: map[string]*Template{
			"files.yaml": {
				Source: "files.yaml.tpl",
			},
		},
	}

	// run the files list through the generator so we can allow
	// conditional file addition
	if err := r.GenerateFiles(ctx, list); err != nil {
		return err
	}

	f, err := os.Open(filepath.Join(r.dir, "files.yaml"))
	if err != nil {
		return err
	}

	// cleanup the file handle and then remove the files
	// we generated earlier
	defer func() {
		f.Close()
		os.Remove(filepath.Join(r.dir, "files.yaml"))
	}()

	var entries *TemplateList
	if err := yaml.NewDecoder(f).Decode(&entries); err != nil {
		return err
	}

	return r.GenerateFiles(ctx, entries)
}

/// GenerateFiles generates files based on a TemplateList being provided.
func (r *Renderer) GenerateFiles(ctx context.Context, list *TemplateList) error {
	// Build the default set of parameters
	args := map[string]interface{}{
		"manifest": r.m,
	}

	for writePath, tmpl := range list.Templates {
		if tmpl.Static {
			if _, err := os.Stat(writePath); err == nil {
				// skip templates we've already written to, when they are marked static
				r.log.Infof(" -> Skipping static file '%s'\n", writePath)
				continue
			}
		}

		data, err := r.FetchTemplate(ctx, filepath.Join("pkg/codegen/templates/", tmpl.Source))
		if err != nil {
			return errors.Wrap(err, "failed to fetch template")
		}

		if err := r.WriteTemplate(ctx, writePath, data, args); err != nil {
			return errors.Wrap(err, "failed to write template")
		}
	}
	return nil
}

// FetchTemplate fetches a template from a git repository, or if Branch is set to ""
// it will attempt to read a template from the ascertained local environment
func (r *Renderer) FetchTemplate(ctx context.Context, filePath string) ([]byte, error) {
	if r.branch == "" {
		// We can use the locally built executable to determine where the templates are stored.
		// This doesn't work for executables that are built outside of our developer environment.
		_, filename, _, _ := runtime.Caller(1)
		fullPath := filepath.Join(path.Dir(filename), filePath[strings.Index(filePath, "templates"):])
		return ioutil.ReadFile(fullPath)
	}

	f, err := r.gfs.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ioutil.ReadAll(f)
}

// WriteTemplate handles the processing, and writing of a template to disk.
func (r *Renderer) WriteTemplate(ctx context.Context, filePath string, contents []byte, args map[string]interface{}) error {
	// Search for any commands that are inscribed in the file.
	// Currently we use StartBlock and EndBlock to allow for
	// arbitrary data payloads to be saved across runs of bootstraper.
	// Eventually we might want to support 3 way merge instead
	f, err := os.Open(filePath)
	if err == nil {
		defer f.Close()

		var curBlockName string
		var i = 1

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			matches := blockRX.FindStringSubmatch(line)
			isCommand := false

			// 1: Comment (###|///)
			// 2: Command
			// 3: Argument to the command
			if len(matches) == 4 {
				cmd := matches[2]
				isCommand = true

				switch cmd {
				case "StartBlock":
					blockName := matches[3]

					if curBlockName != "" {
						r.log.Fatal("Invalid StartBlock when already inside of a block, at %s:%d", filePath, i)
					}
					curBlockName = blockName
				case "EndBlock":
					blockName := matches[3]

					if blockName != curBlockName {
						r.log.Fatal(
							"Invalid EndBlock, found EndBlock with name '%s' while inside of block with name '%s', at %s:%d",
							blockName, curBlockName, filePath, i,
						)
					}

					if curBlockName == "" {
						r.log.Fatal("Invalid EndBlock when not inside of a block, at %s:%d", filePath, i)
					}

					curBlockName = ""
				default:
					isCommand = false
				}
			}

			// we skip lines that had a recognized command in them, or that
			// aren't in a block
			if isCommand || curBlockName == "" {
				continue
			}

			// add the line we processed to the current block we're in
			// and account for having an existing curVal or not. If we
			// don't then we assign curVal to start with the line we
			// just found.
			curVal, ok := args[curBlockName]
			if ok {
				args[curBlockName] = curVal.(string) + "\n" + line
			} else {
				args[curBlockName] = line
			}

		}
	}

	data, err := r.execTemplate(filePath, contents, args)
	if err != nil {
		return err
	}

	absFilePath := filepath.Join(r.dir, filePath)

	action := "Updated"
	if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
		action = "Created"
	}

	ext := filepath.Ext(filePath)
	switch ext {
	case ".sh":
		// post-process shell files by making them executable here
		// TODO(jaredallard): run shfmt on them
		err = r.writeFile(filePath, data, 0744)
	case ".go":
		err = r.postProcessGoFile(filePath, data)
	default:
		err = r.writeFile(filePath, data, 0644)
	}

	r.log.Infof(" -> %s file '%s'", action, filePath)
	if err != nil {
		return errors.Wrapf(err, "error creating file '%s'", absFilePath)
	}

	return err
}

func (r *Renderer) execTemplate(fileName string, body []byte, args map[string]interface{}) ([]byte, error) {
	tmpl, err := template.New(fileName).Funcs(sprig.TxtFuncMap()).Parse(string(body))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, args); err != nil {
		return []byte{}, errors.Wrap(err, "failed to render template")
	}

	return buf.Bytes(), err
}

func (r *Renderer) postProcessGoFile(fileName string, data []byte) error {
	result, err := imports.Process(fileName, data, nil)
	if err != nil {
		// set result to data to allow us to be DRY in writing, we only want a
		// warning here
		result = data

		r.log.Warnf("goimports failed on file '%s': %v", fileName, err)
	}

	return r.writeFile(fileName, result, 0644)
}

func (r *Renderer) writeFile(fileName string, data []byte, perm os.FileMode) error {
	fileName = filepath.Join(r.dir, fileName)
	if err := os.MkdirAll(filepath.Dir(fileName), os.ModePerm); err != nil {
		return err
	}
	return ioutil.WriteFile(fileName, data, perm)
}
