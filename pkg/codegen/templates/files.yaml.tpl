files:
    .editorconfig:
        templatePath: .editorconfig.tpl
        static: false
    .gitattributes:
        templatePath: .gitattributes.tpl
        static: false
    .gitignore:
        templatePath: .gitignore.tpl
        static: false
    .tool-versions:
        templatePath: .tool-versions.tpl
        static: false
    Dockerfile:
        templatePath: Dockerfile.tpl
        static: false
    Makefile:
        templatePath: Makefile.tpl
        static: false
    cmd/{{ .manifest.Name }}/{{ .manifest.Name }}.go:
        templatePath: cmd/main/main.go.tpl
        static: false
    cmd/main/main.go:
        templatePath: cmd/main/main.go.tpl
        static: false
    docker-compose.yaml:
        templatePath: docker-compose.yaml.tpl
        static: false
    go.mod:
        templatePath: go.mod.tpl
        static: true
    go.sum:
        templatePath: go.sum.tpl
        static: true
    {{- if eq .manifest.Type "JobProcessor" }}
    internal/converter/consumer.go:
        templatePath: internal/converter/consumer.go.tpl
        static: false
    {{- end }}
    scripts/gobin.sh:
        templatePath: scripts/gobin.sh.tpl
        static: false
    scripts/lib/logging.sh:
        templatePath: scripts/lib/logging.sh.tpl
        static: false
    scripts/make-log-wrapper.sh:
        templatePath: scripts/make-log-wrapper.sh.tpl
        static: false
    scripts/test.sh:
        templatePath: scripts/test.sh.tpl
        static: false
