workflow "Test" {
  on = "push"
  resolves = [
    "golangci-lint",
    "cedrickring/golang-action/go1.12@1.2.0",
  ]
}

workflow "gorelease" {
  on = "release"
  resolves = "goreleaser"
}

action "Docker Registry" {
  uses = "actions/docker/login@8cdf801b322af5f369e00d85e9cf3a7122f49108"
  secrets = ["DOCKER_USERNAME", "DOCKER_PASSWORD"]
}

action "goreleaser" {
  uses = "docker://goreleaser/goreleaser"
  secrets = [
    "GORELEASER_GITHUB_TOKEN",
  ]
  args = "release"
  needs = ["Docker Registry"]
}

action "golangci-lint" {
  uses = "docker://golangci/golangci-lint"
  runs = "/usr/bin/golangci-lint"
  args = "run --deadline 1h"
  env = {
    GO111MODULE = "on"
  }
}

action "cedrickring/golang-action/go1.12@1.2.0" {
  uses = "cedrickring/golang-action/go1.12@1.2.0"
  args = "go build -o captain ./cmd/captain"
  env = {
    GO111MODULE = "on"
  }
}
