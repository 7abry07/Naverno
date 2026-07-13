alias r := run
alias b := build
alias t := test
alias d := debug
alias dt := debugtest

@run args="":
    go run cmd/app/main.go {{ args }}

@build:
    go build cmd/app/main.go

@test flags="":
    go test {{ flags }} ./...

@debug:
    dlv debug cmd/app/main.go

@debugtest package="":
    dlv test ./internal/{{ package }}
