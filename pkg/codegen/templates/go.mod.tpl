module github.com/tritonmedia/{{ .manifest.Name }}

go 1.14

replace github.com/tritonmedia/pkg => ../pkg

require (
	github.com/nats-io/stan.go v0.7.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/tritonmedia/pkg v0.0.0-20200629192003-d725332d5398
	github.com/urfave/cli/v2 v2.2.0
)
