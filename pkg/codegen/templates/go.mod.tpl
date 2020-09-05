module github.com/tritonmedia/{{ .manifest.Name }}

go 1.14

replace github.com/tritonmedia/pkg => ../pkg

require (
	{{- if eq .manifest.Type "JobProcessor" -}}
	github.com/nats-io/stan.go v0.7.0
	{{- end }}
	github.com/golang/protobuf v1.4.2
	github.com/google/uuid v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/tritonmedia/pkg master
	github.com/urfave/cli/v2 v2.2.0
	google.golang.org/grpc v1.33.0-dev
	google.golang.org/protobuf v1.25.0
)
