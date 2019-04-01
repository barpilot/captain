workflow "Test" {
  on = "push"
  resolves = ["build test", "golangci-lint"]
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

action "build test" {
  uses = "cedrickring/golang-action/go1.12@1.2.0"
}

action "golangci-lint" {
  needs = "build test"
  uses = "docker://golangci/golangci-lint"
  runs = "/usr/bin/golangci-lint"
  args = "run"
}
