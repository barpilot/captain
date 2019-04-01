package captain // import "github.com/harbur/captain"

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

var basedir, _ = os.Getwd()

func TestConfigFiles(t *testing.T) {
	c := configFile("captain.yml")
	sl := "captain.yml"
	assert.Equal(t, sl, c, "Should return possible config files")
}

func TestReadConfig(t *testing.T) {
	c := readConfig(configFile(path.Join(basedir, "/test/Simple/captain.yml")))
	assert.NotNil(t, c, "Should return configuration")
}

func TestNewConfig(t *testing.T) {
	c := NewConfig("", basedir+"/test/Simple/captain.yml", false)
	assert.NotNil(t, c, "Should return captain.yml configuration")
}

func TestNewConfigInferringValues(t *testing.T) {
	c := NewConfig("", basedir+"/test/noCaptainYML/captain.yml", false)
	assert.NotNil(t, c, "Should return infered configuration")
}

func TestFilterConfigEmpty(t *testing.T) {
	c := NewConfig("", basedir+"/test/Simple/captain.yml", false)
	assert.Equal(t, 2, len(c.GetApps()), "Should return 2 apps")

	res := c.FilterConfig([]string{})
	assert.True(t, res, "Should return true")
	assert.Equal(t, 2, len(c.GetApps()), "Should return 2 apps")
}

func TestFilterConfigNonExistent(t *testing.T) {
	c := NewConfig("", basedir+"/test/Simple/captain.yml", false)
	assert.Equal(t, 2, len(c.GetApps()), "Should return 2 apps")

	res := c.FilterConfig([]string{"nonexistent"})
	assert.False(t, res, "Should return false")
	assert.Equal(t, 0, len(c.GetApps()), "Should return 0 apps")
}

func TestFilterConfigWeb(t *testing.T) {
	c := NewConfig("", basedir+"/test/Simple/captain.yml", false)
	assert.Equal(t, 2, len(c.GetApps()), "Should return 2 apps")

	c.FilterConfig([]string{"web"})
	assert.Equal(t, 1, len(c.GetApps()), "Should return 1 app")
	assert.Equal(t, "Dockerfile", c.GetApp("web").Build, "Should return web Build field")
}

func TestGetApp(t *testing.T) {
	c := NewConfig("", basedir+"/test/Simple/captain.yml", false)
	app := c.GetApp("web")
	assert.Equal(t, "harbur/test_web", app.Image, "Should return web image")
}
