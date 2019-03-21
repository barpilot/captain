workflow "Deploy" {
  on = "push"
  resolves = [
    "goreleaser",
    "is-tag",
    "Test on Travis CI",
  ]
}

action "is-tag" {
  uses = "actions/bin/filter@master"
  args = "tag"
}

action "goreleaser" {
  uses = "docker://goreleaser/goreleaser"
  secrets = [
    "GORELEASER_GITHUB_TOKEN",
  ]
  args = "release"
  needs = ["is-tag","Test on Travis CI"]
}

action "Test on Travis CI" {
  uses = "travis-ci/actions@master"
  secrets = ["TRAVIS_TOKEN"]
}
