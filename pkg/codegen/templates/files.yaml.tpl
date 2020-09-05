files:
    .ci/docker-builder.sh:
        templatePath: .ci/docker-builder.sh.tpl
    .editorconfig:
        templatePath: .editorconfig.tpl
    .gitattributes:
        templatePath: .gitattributes.tpl
    .github/workflows/build.yaml:
        templatePath: .github/workflows/build.yaml.tpl
    .gitignore:
        templatePath: .gitignore.tpl
    .golangci.yml:
        templatePath: .golangci.yml.tpl
    .tool-versions:
        templatePath: .tool-versions.tpl
    .vscode/extensions.json:
        templatePath: .vscode/extensions.json.tpl
    .vscode/settings.json:
        templatePath: .vscode/settings.json.tpl
    Dockerfile:
        templatePath: Dockerfile.tpl
    Makefile:
        templatePath: Makefile.tpl
    cmd/{{ .manifest.Name }}/{{ .manifest.Name }}.go:
        templatePath: cmd/main/main.go.tpl
    docker-compose.yaml:
        templatePath: docker-compose.yaml.tpl
    go.mod:
        templatePath: go.mod.tpl
        static: true
    go.sum:
        templatePath: go.sum.tpl
        static: true
    {{- if eq .manifest.Type "JobProcessor" }}
    internal/{{ .manifest.Name }}/consumer.go:
        templatePath: internal/consumer/consumer.go.tpl
    {{- end }}
    {{- if eq .manifest.Type "GRPC" }}
    internal/{{ .manifest.Name }}/grpc.go:
        templatePath: internal/grpc/grpc.go.tpl
    internal/{{ .manifest.Name }}/grpc_server.go:
        templatePath: internal/grpc/grpc_server.go.tpl
    {{- end }}
    renovate.json:
        templatePath: renovate.json.tpl
    scripts/gobin.sh:
        templatePath: scripts/gobin.sh.tpl
    scripts/goimports.sh:
        templatePath: scripts/goimports.sh.tpl
    scripts/golangci-lint.sh:
        templatePath: scripts/golangci-lint.sh.tpl
    scripts/lib/logging.sh:
        templatePath: scripts/lib/logging.sh.tpl
    scripts/make-log-wrapper.sh:
        templatePath: scripts/make-log-wrapper.sh.tpl
    scripts/protoc-gen-go-grpc.sh:
        templatePath: scripts/protoc-gen-go-grpc.sh.tpl
    scripts/protoc-gen-go.sh:
        templatePath: scripts/protoc-gen-go.sh.tpl
    scripts/protoc.sh:
        templatePath: scripts/protoc.sh.tpl
    scripts/shellcheck.sh:
        templatePath: scripts/shellcheck.sh.tpl
    scripts/shfmt.sh:
        templatePath: scripts/shfmt.sh.tpl
    scripts/test.sh:
        templatePath: scripts/test.sh.tpl
