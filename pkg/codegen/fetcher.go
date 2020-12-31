package codegen

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tritonmedia/bootstraper/internal/vfs"
	"gopkg.in/yaml.v3"
)

type Fetcher struct {
	log logrus.FieldLogger
	m   *ServiceManifest
}

func NewFetcher(log logrus.FieldLogger, m *ServiceManifest) *Fetcher {
	return &Fetcher{log, m}
}

func (f *Fetcher) DownloadRepository(r TemplateRepository) (billy.Filesystem, error) {
	fs := memfs.New()

	auth, err := ssh.NewSSHAgentAuth("git")
	if err != nil {
		return nil, err
	}

	opts := &git.CloneOptions{
		URL:               r.GitURL,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Depth:             1,
		Auth:              auth,
	}

	if r.Version != "" {
		opts.ReferenceName = plumbing.NewTagReferenceName(r.Version)
		opts.SingleBranch = true
	}

	f.log.Infof("Downloading repository '%s'", r.GitURL)
	_, err = git.Clone(memory.NewStorage(), fs, opts)
	if err != nil {
		return nil, err
	}

	return fs, nil
}

func (f *Fetcher) ParseRepositoryManifest(r TemplateRepository, fs billy.Filesystem) (*TemplateRepositoryManifest, error) {
	mf, err := fs.Open("manifest.yaml")
	if err != nil {
		return nil, err
	}

	var manifest *TemplateRepositoryManifest
	dec := yaml.NewDecoder(mf)
	err = dec.Decode(&manifest)
	return manifest, err
}

// ResolveDependencies resolved the dependencies of a given template repository.
// It currently only supports one level dependency resolution and doesn't do any
// smart logic for ordering other than first wins.
func (f *Fetcher) ResolveDependencies(filesystems map[string]bool, r *TemplateRepositoryManifest) ([]billy.Filesystem, map[string]Argument, error) {
	depFilesystems := make([]billy.Filesystem, 0)
	args := make(map[string]Argument)
	for _, d := range r.Dependencies {
		// If the filesystem already exists, then we can just skip it
		// since something already required it.
		if _, ok := filesystems[d.GitURL]; ok {
			continue
		}

		fs, err := f.DownloadRepository(d)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to download repository '%s'", d.GitURL)
		}
		filesystems[d.GitURL] = true

		mf, err := f.ParseRepositoryManifest(d, fs)
		if err != nil {
			return nil, nil, err
		}
		for k, v := range mf.Arguments {
			args[k] = v
		}

		subDepFilesystems, args, err := f.ResolveDependencies(filesystems, mf)
		if err != nil {
			return nil, nil, err
		}
		for k, v := range args {
			args[k] = v
		}

		// append the resolved dependencies of the sub-dependencies to the array of dependencies
		// of the manifest we're operating on. Be sure to put the filesystem of this dependency after
		// it's sub-dependencies.
		depFilesystems = append(depFilesystems, append(subDepFilesystems, fs)...)
	}

	return depFilesystems, args, nil
}

func (f *Fetcher) CreateVFS() (billy.Filesystem, map[string]Argument, error) {
	// Create a shim template manifest from our service dependencies
	layers, args, err := f.ResolveDependencies(make(map[string]bool), &TemplateRepositoryManifest{
		Dependencies: f.m.Repositories,
	})
	if err != nil {
		return nil, nil, err
	}

	return vfs.NewMergedFilesystem(layers...), args, nil
}
