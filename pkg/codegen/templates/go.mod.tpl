module github.com/tritonmedia/{{ .manifest.Name }}

go 1.15

replace github.com/tritonmedia/pkg => ../pkg

require (
	{{- if eq .manifest.Type "JobProcessor" -}}
	github.com/nats-io/stan.go v0.7.0
	{{- end }}
	github.com/google/uuid v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/tritonmedia/pkg master
	github.com/urfave/cli/v2 v2.2.0
	{{- if eq .manfiest.Type "GRPC" -}}
	github.com/golang/protobuf v1.4.3
	google.golang.org/grpc v1.33.1
	google.golang.org/protobuf v1.25.0
	{{- end }}
)
