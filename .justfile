alias r := run
alias b := build
alias t := test
alias d := debug

@run args="":
    go run cmd/app/main.go {{ args }}

@build:
    go build cmd/app/main.go

@test:
    go test ./...

@debug:
    dlv debug cmd/app/main.go
