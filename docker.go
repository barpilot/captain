package captain // import "github.com/harbur/captain"

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	docker "github.com/fsouza/go-dockerclient"
)

var client *docker.Client

func init() {
	var err error
	client, err = docker.NewClientFromEnv()
	if err != nil {
		panic(err)
	}
}

type BuildArgSet struct {
	slice []docker.BuildArg
}

func buildImage(app App, tag string, pathConfig string, force bool) error {
	pInfo("Building image %s:%s", app.Image, tag)

	// Nasty issue with CircleCI https://github.com/docker/docker/issues/4897
	if os.Getenv("CIRCLECI") == "true" {
		pInfo("Running at %s environment...", "CIRCLECI")
		return execute("docker", "build", "-t", app.Image+":"+tag, filepath.Dir(app.Build))
	}

	// Create BuildArg set
	buildArgSet := BuildArgSet{(make([]docker.BuildArg, 0, 10))}
	if len(app.Build_arg) > 0 {
		for k, arg := range app.Build_arg {
			buildArgSet.slice = append(buildArgSet.slice, docker.BuildArg{Name: k, Value: arg})
		}
	}
	contextDir := path.Join(pathConfig, app.Context)
	pInfo(contextDir)
	pInfo(path.Join(pathConfig, app.Build))
	Dockerfile := app.Build

	opts := docker.BuildImageOptions{
		Name:                app.Image + ":" + tag,
		Dockerfile:          Dockerfile,
		NoCache:             force,
		SuppressOutput:      false,
		RmTmpContainer:      true,
		ForceRmTmpContainer: true,
		OutputStream:        os.Stdout,
		ContextDir:          contextDir,
		BuildArgs:           buildArgSet.slice,
	}

	// Use ~/.docker/ auth configuration if exists
	dockercfg, _ := docker.NewAuthConfigurationsFromDockerCfg()
	if dockercfg != nil {
		opts.AuthConfigs = *dockercfg
	}

	if err := client.BuildImage(opts); err != nil {
		pError("%s", err)
		return err
	}

	return nil
}

func pushImage(image string, version string) error {
	return execute("docker", "push", image+":"+version)
}

func pullImage(image string, version string) error {
	return execute("docker", "pull", image+":"+version)
}

func tagImage(app App, origin string, tag string) error {
	if tag != "" {
		pInfo("Tagging image %s:%s as %s:%s", app.Image, origin, app.Image, tag)
		opts := docker.TagImageOptions{Repo: app.Image, Tag: tag, Force: true}
		err := client.TagImage(app.Image+":"+origin, opts)
		if err != nil {
			fmt.Printf("%s", err)
		}
		return err
	}

	pDebug("Skipping tag of %s - no git repository", app.Image)

	return nil
}

func removeImage(name string) error {
	return client.RemoveImage(name)
}

/**
 * Retrieves a list of existing Images for the specific App.
 */
func getImages(app App) []docker.APIImages {
	pDebug("Getting images %s", app.Image)
	imgs, _ := client.ListImages(docker.ListImagesOptions{All: false, Filter: app.Image})
	return imgs
}

func imageExist(app App, tag string) bool {
	repo := app.Image + ":" + tag
	image, _ := client.InspectImage(repo)
	if image != nil {
		return true
	}
	return false
}
