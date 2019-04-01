package captain // import "github.com/harbur/captain"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// Config represents the information stored at captain.yml. It keeps information about images and unit tests.
type Config interface {
	FilterConfig(filter []string) bool
	GetApp(app string) App
	GetApps() []App
	GetPath() string
}

type configV1 struct {
	Build  build
	Test   map[string][]string
	Images []string
	Root   []string
}

type build struct {
	Images map[string]string
}

type config struct {
	Apps map[string]App `yaml:",inline"`
	Path string         `yaml:"-"`
}

//var configOrder *yaml.MapSlice

// App struct
type App struct {
	Build     string            `yaml:"build"`
	Image     string            `yaml:"image"`
	Context   string            `yaml:"context,omitempty"`
	Pre       []string          `yaml:"pre,omitempty"`
	Post      []string          `yaml:"post,omitempty"`
	Test      []string          `yaml:"test,omitempty"`
	Build_arg map[string]string `yaml:"build_arg,omitempty"`
}

func (a *App) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawApp App
	raw := rawApp{Build: "Dockerfile", Context: "."} // Put your defaults here
	if err := unmarshal(&raw); err != nil {
		return err
	}

	*a = App(raw)
	return nil
}

// configFile returns the file to read the config from.
// If the --config option was given,
// it will only use the given file.
func configFile(path string) string {
	if len(path) > 0 {
		return path
	}
	return "captain.yml"
}

// readConfig will read the config file
// and return the created config.
func readConfig(filename string) *config {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(StatusError{err, 74})
	}
	conf := unmarshal(data)
	conf.Path = filepath.Dir(filename)
	return conf
}

// displaySyntaxError will display more information
// such as line and error type given an error and
// the data that was unmarshalled.
// Thanks to https://github.com/markpeek/packer/commit/5bf33a0e91b2318a40c42e9bf855dcc8dd4cdec5
func displaySyntaxError(data []byte, syntaxError error) (err error) {
	syntax, ok := syntaxError.(*json.SyntaxError)
	if !ok {
		err = syntaxError
		return
	}
	newline := []byte{'\x0a'}
	space := []byte{' '}

	start, end := bytes.LastIndex(data[:syntax.Offset], newline)+1, len(data)
	if idx := bytes.Index(data[start:], newline); idx >= 0 {
		end = start + idx
	}

	line, pos := bytes.Count(data[:start], newline)+1, int(syntax.Offset)-start-1

	err = fmt.Errorf("\nError in line %d: %s \n%s\n%s^", line, syntaxError, data[start:end], bytes.Repeat(space, pos))
	return
}

// unmarshal converts either JSON
// or YAML into a config object.
func unmarshal(data []byte) *config {
	var configV1 *configV1
	_ = yaml.Unmarshal(data, &configV1)
	if len(configV1.Build.Images) > 0 {
		pError("Old %s format detected! Please check the https://github.com/harbur/captain how to upgrade", "captain.yml")
		os.Exit(-1)
	}

	var config *config
	err := yaml.Unmarshal(data, &config)

	if err != nil {
		err = displaySyntaxError(data, err)
		pError("%s", err)
		os.Exit(InvalidCaptainYML)
	}

	return config
}

// NewConfig returns a new Config instance based on the reading the captain.yml
// file at path.
// Containers will be ordered so that they can be
// brought up and down with Docker.
func NewConfig(namespace, path string, forceOrder bool) Config {
	var conf *config
	f := configFile(path)
	if _, err := os.Stat(f); err == nil {
		conf = readConfig(f)
	}

	if conf == nil {
		pInfo("No configuration found %v - inferring values", configFile(path))
		autoconf := config{Path: filepath.Dir(path)}
		autoconf.Apps = make(map[string]App)
		conf = &autoconf
		dockerfiles := getDockerfiles(namespace)
		for build, image := range dockerfiles {
			autoconf.Apps[image] = App{Build: build, Image: image}
		}
	}

	var err error
	if err != nil {
		panic(StatusError{err, 78})
	}
	return conf
}

// GetApps returns a list of Apps
func (c *config) GetApps() []App {
	var apps []App

	for _, app := range c.Apps {
		apps = append(apps, app)
	}

	return apps
}

func (c *config) FilterConfig(filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	untouched := true
	for name, _ := range c.Apps {
		filtered := true
		for _, filter := range filters {
			if name == filter {
				filtered = false
				break
			}
		}
		if filtered {
			untouched = false
			delete(c.Apps, name)
		}
	}
	return untouched
}

// GetApp returns App configuration
func (c *config) GetApp(app string) App {
	return c.Apps[app]
}

func (c *config) GetPath() string {
	return c.Path
}

// Global list, how can I pass it to the visitor pattern?
// var imagesMap = make(map[string]string)

func getDockerfiles(namespace string) map[string]string {
	var imagesMap = make(map[string]string)
	if err := filepath.Walk(".", visit(namespace, imagesMap)); err != nil {
		pError(err.Error())
	}
	return imagesMap
}

func visit(namespace string, images map[string]string) filepath.WalkFunc {
	return func(path string, f os.FileInfo, err error) error {
		// Filename is "Dockerfile" or has "Dockerfile." prefix and is not a directory
		if (f.Name() == "Dockerfile" || strings.HasPrefix(f.Name(), "Dockerfile.")) && !f.IsDir() {
			// Get Parent Dirname
			absolutePath, _ := filepath.Abs(path)
			var image = strings.ToLower(filepath.Base(filepath.Dir(absolutePath)))
			images[path] = namespace + "/" + image + strings.ToLower(filepath.Ext(path))
			pInfo("Located %s will be used to create %s", path, images[path])
		}
		return nil
	}
}
