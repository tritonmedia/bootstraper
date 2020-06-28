package codegen

// ServiceType is a general tritonmedia type of service. It's an
// enum for type safety, but in reality it could be set to anything.
type ServiceType string

var (
	GRPC         ServiceType = "GRPC"
	JobProcessor ServiceType = "JobProcessor"
)

// TemplateList is a list of templates in a directory that should be
// processed
type TemplateList struct {
	// Templates contains a list of templates to write to. The string here is the
	// path that a template should be written to.
	Templates map[string]*Template `yaml:"files"`
}

// ServiceManifest is a manifest used to describe a service and impact
// what files are included
type ServiceManifest struct {
	Name string `yaml:"name"`

	// Type is the general type of application this is
	Type ServiceType `yaml:"type"`

	// Arguments is a map of arbitrary arguments to pass to the generator
	Arguments map[string]interface{}
}

type Template struct {
	// Source is where the template is in relation to `pkg/codegen/templates`
	Source string `yaml:"templatePath"`

	// Static determines if this template should be written only once.
	Static bool `yaml:"static"`
}
