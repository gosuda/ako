package main

const golangcilintConfig = `version: "2"
linters:
  default: all
  enable:
    - asasalint
    - asciicheck
    - bidichk
    - bodyclose
    - canonicalheader
    - containedctx
    - contextcheck
    - copyloopvar
    - cyclop
    - decorder
    - dogsled
    - durationcheck
    - err113
    - errcheck
    - errchkjson
    - errname
    - errorlint
    - exhaustive
    - exptostd
    - fatcontext
    - forbidigo
    - forcetypeassert
    - funcorder
    - funlen
    - ginkgolinter
    - gocheckcompilerdirectives
    - gochecknoinits
    - gochecksumtype
    - gocognit
    - goconst
    - gocritic
    - gocyclo
    - godot
    - godox
    - goheader
    - gomoddirectives
    - gomodguard
    - goprintffuncname
    - gosec
    - gosmopolitan
    - govet
    - grouper
    - iface
    - importas
    - inamedparam
    - ineffassign
    - interfacebloat
    - intrange
    - ireturn
    - lll
    - loggercheck
    - maintidx
    - makezero
    - mirror
    - misspell
    - mnd
    - musttag
    - nakedret
    - nestif
    - nilerr
    - nilnesserr
    - nilnil
    - nlreturn
    - noctx
    - nolintlint
    - nonamedreturns
    - nosprintfhostport
    - paralleltest
    - perfsprint
    - prealloc
    - predeclared
    - promlinter
    - protogetter
    - reassign
    - recvcheck
    - revive
    - rowserrcheck
    - sloglint
    - spancheck
    - sqlclosecheck
    - staticcheck
    - tagalign
    - tagliatelle
    - testableexamples
    - testifylint
    - testpackage
    - thelper
    - tparallel
    - unconvert
    - unparam
    - unused
    - usestdlibvars
    - usetesting
    - varnamelen
    - wastedassign
    - whitespace
    - wrapcheck
    - wsl
    - zerologlint
  disable:
    - dupl
    - dupword
	- exhaustruct
	- depguard
	- gochecknoglobals
  exclusions:
    generated: lax
    warn-unused: true
    presets:
      - comments
      - std-error-handling
      - common-false-positives
      - legacy
    rules:
      - path: _test\.go
        linters:
          - gocyclo
          - errcheck
          - dupl
          - gosec
      - path-except: _test\.go
        linters:
          - forbidigo
    paths:
      - ".*/lib/adapter/gen/*.go$"
formatters:
  enable:
    - gci
    - gofmt
    - gofumpt
    - goimports
    - golines
  exclusions:
    generated: lax
    paths:
      - ".*/lib/adapter/gen/*.go$"
issues:
  max-issues-per-linter: 50
  max-same-issues: 3
  uniq-by-line: false
  fix: true
output:
  sort-order:
    - linter
    - severity
    - file # filepath, line, and column.
run:
  timeout: 5m
  relative-path-mode: gomod
  concurrency: 8
  allow-parallel-runners: true
  allow-serial-runners: false
  modules-download-mode: mod
`
