files:
    .gitignore:
        templatePath: .gitignore.tpl
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
    docker-compose.yaml:
        templatePath: docker-compose.yaml.tpl
        static: false
    go.mod:
        templatePath: go.mod.tpl
        static: true
    go.sum:
        templatePath: go.sum.tpl
        static: true
    internal/converter/consumer.go:
        templatePath: internal/converter/consumer.go.tpl
        static: false
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
