package {{ .manifest.Name }}

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	apiv1 "github.com/tritonmedia/{{ .manifest.Name }}/api/v1"
	"github.com/tritonmedia/{{ .manifest.Name }}/internal/ent"

	///StartBlock(imports)
	{{- if .imports }}
{{ .imports }}
	{{- end }}
	///EndBlock(imports)	
)

type GRPCServiceHandler struct {
	log logrus.FieldLogger

	///StartBlock(grpcConfig)
	{{- if .grpcConfig }}
{{ .grpcConfig }}
	{{- end }}
	///EndBlock(grpcConfig)
}

///StartBlock(global)
{{- if .global }}
{{ .global }}
{{- end }}
///EndBlock(global)

func NewServiceHandler(ctx context.Context, log logrus.FieldLogger) (*GRPCServiceHandler, error) {
	///StartBlock(grpcInit)
	{{- if .grpcInit }}
{{ .grpcInit }}
	{{- end }}
	///EndBlock(grpcInit)

	return &GRPCServiceHandler{
		log,
		///StartBlock(grpcConfigInit)
		{{- if .grpcConfigInit }}
{{ .grpcConfigInit }}
		{{- end }}
		///EndBlock(grpcConfigInit)
	}, nil
}

///StartBlock(grpcHandlers)
{{- if .grpcHandlers }}
{{ .grpcHandlers }}
{{- end }}
///EndBlock(grpcHandlers)
