package codegen

// TemplateList is a list of templates in a directory that should be
// processed
type TemplateList struct {
	// Templates contains a list of templates to write to. The string here is the
	// path that a template should be written to.
	Templates map[string]*Template `yaml:"files"`
}

type Template struct {
	// Source is where the template is in relation to `pkg/codegen/templates`
	Source string `yaml:"templatePath"`

	// Static determines if this template should be written only once.
	Static bool `yaml:"static,omitempty"`
}

// ServiceManifest is a manifest used to describe a service and impact
// what files are included
type ServiceManifest struct {
	// Name is the name of the service
	Name string `yaml:"name"`

	// Repositories are the template repositories that this service depends
	// on and utilizes
	Repositories []TemplateRepository `yaml:"repositories"`

	// Arguments is a map of arbitrary arguments to pass to the generator
	Arguments map[string]string `yaml:"arguments"`
}

// TemplateRepository is a repository of template files.
type TemplateRepository struct {
	// GitURL is the fully qualified Git URL that is able to access the templates
	// and manifest.
	GitURL string `yaml:"gitUrl"`

	// Version is a semantic version of the template repository that should be downloaded
	// if not set then the latest version is used.
	Version string `yaml:"version"`
}

// TemplateRepositoryManifest is a manifest of a template repository
type TemplateRepositoryManifest struct {
	// Name is the name of this template repository.
	// This is likely to be used in the future.
	Name string `yaml:"name"`

	// Dependencies are template repositories that should be applied before
	// this one is. If they are not required by the service that required us
	// then they will be brought in.
	Dependencies []TemplateRepository `yaml:"dependencies"`

	// Arguments are a declaration of arguments to the template generator
	Arguments map[string]Argument
}

type Argument struct {
	// Required denotes this argument as required.
	Required bool `yaml:"required"`

	// Type declares the type of the argument. This is not implemented
	// yet, so is likely to change in the future.
	Type string `yaml:"type"`

	// Values is a list of possible values for this, if empty all input is
	// considered valid.
	Values []string `yaml:"values"`

	// Description is a description of this argument. Optional.
	Description string `yaml:"description"`
}
