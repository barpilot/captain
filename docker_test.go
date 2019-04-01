package captain // import "github.com/harbur/captain"

import (
	"path"
	"testing"

	"os"

	"github.com/stretchr/testify/assert"
)

func TestBuildImage(t *testing.T) {
	app := App{Build: "Dockerfile", Image: "captain_test"}
	res := buildImage(app, "latest", path.Join(basedir, "test/noCaptainYML"), false)
	assert.Nil(t, res, "Docker build should not return any error")
}

func TestBuildImageError(t *testing.T) {
	app := App{Build: "Dockerfile.error", Image: "captain_test"}
	res := buildImage(app, "latest", path.Join(basedir, "/test/noCaptainYML"), false)
	assert.NotNil(t, res, "Docker build should return an error")
}

func TestBuildImageCircleCI(t *testing.T) {
	os.Setenv("CIRCLECI", "true")
	app := App{Build: "test/noCaptainYML/Dockerfile", Image: "captain_test"}
	res := buildImage(app, "latest", path.Join(basedir, "/test/noCaptainYML"), false)
	assert.Nil(t, res, "Docker build should not return any error")
}

func TestTagImage(t *testing.T) {
	app := App{Image: "golang"}
	res := tagImage(app, "1.4.2", "testing")
	assert.Nil(t, res, "Docker tag should not return any error")
}

func TestTagNonexistingImage(t *testing.T) {
	app := App{Image: "golang"}
	res := tagImage(app, "nonexist", "testing")
	assert.NotNil(t, res, "Docker tag should return an error")
	println()
}

func TestImageExist(t *testing.T) {
	app := App{Image: "golang"}
	exist := imageExist(app, "1.4.2")
	assert.Equal(t, true, exist, "Docker image golang:1.4.2 should exist")
}

func TestImageDoesNotExist(t *testing.T) {
	app := App{Image: "golang"}
	exist := imageExist(app, "nonexist")
	assert.Equal(t, false, exist, "Docker image golang:nonexist should not exist")
}
