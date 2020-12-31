// package codegen handles rendering bootstraper templates
package codegen

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
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
	"github.com/tritonmedia/bootstraper/internal/vfs"

	"github.com/go-git/go-billy/v5"
	"golang.org/x/tools/imports"
)

var (
	blockRX = regexp.MustCompile(`\w*(///|###)\s*([a-zA-Z]+)\(([a-zA-Z]+)\)`)
)

type Renderer struct {
	branch string
	dir    string
	m      *ServiceManifest

	fetcher *Fetcher
	log     logrus.FieldLogger

	args map[string]Argument
}

// NewRenderer creates a new template renderer that is the heart of bootstraper.
func NewRenderer(log logrus.FieldLogger, branch, dir string, m *ServiceManifest) *Renderer {
	fetcher := NewFetcher(log, m)
	return &Renderer{
		fetcher: fetcher,
		branch:  branch,
		dir:     dir,
		m:       m,
		log:     log,
	}
}

// Render starts the code generation process
func (r *Renderer) Render(ctx context.Context, log *logrus.Entry) error {
	if len(r.m.Repositories) == 0 {
		return fmt.Errorf("missing template repositories, must specify at least one")
	}

	fs, args, err := r.fetcher.CreateVFS()
	if err != nil {
		return err
	}
	r.args = args

	for k, a := range args {
		v, isPresent := r.m.Arguments[k]

		if !isPresent && a.Required {
			return fmt.Errorf("missing required argument '%s'", k)
		}

		if v != "" && len(a.Values) > 0 {
			found := false
			for _, allowedV := range a.Values {
				if v == allowedV {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("invalid value for argument '%s', expected: %v, got: %v", k, a.Values, v)
			}
		}
	}

	return r.GenerateFiles(ctx, fs)
}

// GenerateFiles generates files based on a TemplateList being provided.
func (r *Renderer) GenerateFiles(ctx context.Context, fs billy.Filesystem) error {
	// Build the default set of parameters
	args := map[string]interface{}{
		"manifest": r.m,
	}

	return vfs.Walk(fs, "", func(path string, info os.FileInfo, err error) error {
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

		data, err := r.FetchTemplate(ctx, fs, path)
		if err != nil {
			return errors.Wrap(err, "failed to fetch template")
		}

		path = strings.TrimSuffix(path, ".tpl")

		if err := r.WriteTemplate(ctx, path, data, args); err != nil {
			return errors.Wrap(err, "failed to write template")
		}

		return nil
	})
}

// FetchTemplate fetches a template from a git repository, or if Branch is set to ""
// it will attempt to read a template from the ascertained local environment
func (r *Renderer) FetchTemplate(ctx context.Context, fs billy.Filesystem, filePath string) ([]byte, error) {
	if r.branch == "" {
		// We can use the locally built executable to determine where the templates are stored.
		// This doesn't work for executables that are built outside of our developer environment.
		_, filename, _, _ := runtime.Caller(1)
		fullPath := filepath.Join(path.Dir(filename), filePath[strings.Index(filePath, "templates"):])
		return ioutil.ReadFile(fullPath)
	}

	f, err := fs.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ioutil.ReadAll(f)
}

// WriteTemplate handles the processing, and writing of a template to disk.
func (r *Renderer) WriteTemplate(ctx context.Context, filePath string, contents []byte, args map[string]interface{}) error { //nolint:funlen,gocyclo,lll
	// Search for any commands that are inscribed in the file.
	// Currently we use StartBlock and EndBlock to allow for
	// arbitrary data payloads to be saved across runs of bootstraper.
	// Eventually we might want to support 3 way merge instead
	f, err := os.Open(filePath)
	if err == nil {
		defer f.Close()

		var curBlockName string
		var i = 0

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			i++

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
						r.log.Fatalf("Invalid StartBlock when already inside of a block, at %s:%d", filePath, i)
					}
					curBlockName = blockName
				case "EndBlock":
					blockName := matches[3]

					if blockName != curBlockName {
						r.log.Fatalf(
							"Invalid EndBlock, found EndBlock with name '%s' while inside of block with name '%s', at %s:%d",
							blockName, curBlockName, filePath, i,
						)
					}

					if curBlockName == "" {
						r.log.Fatalf("Invalid EndBlock when not inside of a block, at %s:%d", filePath, i)
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

	data, isStatic, shouldWriteFile, newFilePath, err := r.execTemplate(filePath, contents, args)
	if err != nil {
		return err
	}

	absFilePath := filepath.Join(r.dir, newFilePath)

	action := "Updated"

	// Why: We're fine shadowing err.
	//nolint:govet
	if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
		action = "Created"
	}

	if isStatic && action != "Created" {
		shouldWriteFile = false
	}

	if !shouldWriteFile {
		action = "Skipping"
	}

	if shouldWriteFile {
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
	}

	r.log.Infof(" -> %s file '%s'", action, filePath)
	if err != nil {
		return errors.Wrapf(err, "error creating file '%s'", absFilePath)
	}

	return err
}

// execTemplate executes a template and gets back metadata
// returns the byte contents, if static, if we should write the file, the potentially new file name
// and an error if it occurred
func (r *Renderer) execTemplate(fileName string, body []byte, args map[string]interface{}) ([]byte, bool, bool, string, error) {
	isStatic := false
	writeFile := true
	outputName := fileName

	funcs := sprig.TxtFuncMap()

	// argEq checks to see if an argument is equal to a given value
	funcs["argEq"] = func(argName, value string) bool {
		return r.m.Arguments[argName] == value
	}

	// Static marks this file as static and doesn't write it if it already exists
	funcs["static"] = func() bool {
		isStatic = true
		return false
	}

	// setOutputName sets the output of this file
	funcs["setOutputName"] = func(out string) bool {
		outputName = out
		return false
	}

	// writeIf writes this file only if a given argument is equal to
	// a specified value
	funcs["writeIf"] = func(argName, value string) bool {
		writeFile = false
		if r.m.Arguments[argName] == value {
			writeFile = true
		}
		return false
	}

	tmpl, err := template.New(fileName).Funcs(funcs).Parse(string(body))
	if err != nil {
		return nil, false, false, "", err
	}

	var buf bytes.Buffer
	// Why: We're fine shadowing err.
	//nolint:govet
	if err := tmpl.Execute(&buf, args); err != nil {
		return []byte{}, false, false, "", errors.Wrap(err, "failed to render template")
	}

	return buf.Bytes(), isStatic, writeFile, outputName, err
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
