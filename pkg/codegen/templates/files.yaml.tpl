files:
    cmd/{{ .manifest.Name }}/{{ .manifest.Name }}.go:
        templatePath: cmd/main/main.go.tpl
        static: false
