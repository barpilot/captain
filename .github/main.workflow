workflow "Deploy" {
  on = "push"
  resolves = [
    "goreleaser",
    "Test on Travis CI",
    "Docker Registry",
  ]
}

action "is-tag" {
  uses = "actions/bin/filter@master"
  args = "tag"
}



action "Test on Travis CI" {
  uses = "travis-ci/actions@master"
  secrets = ["TRAVIS_TOKEN"]
}

action "Docker Registry" {
  uses = "actions/docker/login@8cdf801b322af5f369e00d85e9cf3a7122f49108"
  needs = ["is-tag"]
  secrets = ["DOCKER_USERNAME", "DOCKER_PASSWORD"]
}

action "goreleaser" {
  uses = "docker://goreleaser/goreleaser"
  secrets = [
    "GORELEASER_GITHUB_TOKEN",
  ]
  args = "release"
  needs = ["is-tag", "Test on Travis CI", "Docker Registry"]
}
