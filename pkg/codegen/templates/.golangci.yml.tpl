linters-settings:
  dupl:
    threshold: 100
  exhaustive:
    default-signifies-exhaustive: false
  funlen:
    lines: 100
    statements: 50
  goconst:
    min-len: 2
    min-occurrences: 2
  gocritic:
    enabled-tags:
      - diagnostic
      - experimental
      - opinionated
      - performance
      - style
    disabled-checks:
      - dupImport # https://github.com/go-critic/go-critic/issues/845
      - ifElseChain
      - octalLiteral
      - whyNoLint
      - wrapperFunc
      - commentFormatting # Breaks bootstraper
      - filepathJoin # why?
  gocyclo:
    min-complexity: 15
  goimports:
    local-prefixes: github.com/tritonmedia/{{ .manifest.Name }}
  golint:
    min-confidence: 0
  govet:
    check-shadowing: true
  lll:
    line-length: 140
  maligned:
    suggest-new: true
  misspell:
    locale: US
  nolintlint:
    allow-leading-space: true # don't require machine-readable nolint directives (i.e. with no leading space)
    allow-unused: false # report any unused nolint directives
    require-explanation: false # don't require an explanation for nolint directives
    require-specific: false # don't require nolint directives to be specific about which linter is being skipped
linters:
  disable-all: true
  enable:
    - bodyclose
    - deadcode
    - depguard
    - dogsled
    - dupl
    - errcheck
    - exhaustive
    - funlen
    - gochecknoinits
    - goconst
    - gocritic
    - gocyclo
    - gofmt
    - goimports
    - golint
    - goprintffuncname
    - gosec
    - gosimple
    - govet
    - ineffassign
    - interfacer
    - lll
    - misspell
    - nakedret
    - noctx
    - nolintlint
    - rowserrcheck
    - scopelint
    - staticcheck
    - structcheck
    - stylecheck
    - typecheck
    - unconvert
    - unparam
    - unused
    - varcheck
    - whitespace