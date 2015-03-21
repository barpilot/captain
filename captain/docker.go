package captain // import "github.com/harbur/captain/captain"

import (
	"os"

	"github.com/fsouza/go-dockerclient"
)

var endpoint = "unix:///var/run/docker.sock"
var client, _ = docker.NewClient(endpoint)

func buildImage(dockerfile string, image string) {
	info("Building image %s", image)

	opts := docker.BuildImageOptions{
		Name:                image,
		NoCache:             false,
		SuppressOutput:      false,
		RmTmpContainer:      true,
		ForceRmTmpContainer: true,
		OutputStream:        os.Stdout,
		ContextDir:          ".",
	}
	client.BuildImage(opts)
}

func tagImage(target string, repo string, tag string) {
	info("Tagging image as %s", target)
	opts := docker.TagImageOptions{Repo: repo, Tag: tag, Force: true}
	client.TagImage(target, opts)
}