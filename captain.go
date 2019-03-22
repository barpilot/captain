package captain // import "github.com/harbur/captain"

import (
	"fmt"
	"os"
	"strings"
)

// Debug can be turned on to enable debug mode.
var Debug bool

// StatusError provides error code and id
type StatusError struct {
	err    error
	status int
}

// Pre function executes commands on pre section before build
func Pre(app App) error {
	for _, value := range app.Pre {
		pInfo("Running pre command: %s", value)
		res := execute("bash", "-c", value)
		if res != nil {
			return res
		}
	}
	return nil
}

// Post function executes commands on pre section after build
func Post(app App) error {
	for _, value := range app.Post {
		pInfo("Running post command: %s", value)
		res := execute("bash", "-c", value)
		if res != nil {
			return res
		}
	}
	return nil
}

type BuildOptions struct {
	Config       Config
	Tag          string
	Force        bool
	All_branches bool
	Long_sha     bool
	Branch_tags  bool
	Commit_tags  bool
}

// Build function compiles the Containers of the project
func Build(opts BuildOptions) {
	config := opts.Config

	rev, err := getRevision(opts.Long_sha)
	if err != nil {
		fmt.Println(err)
		return
	}

	// For each App
	for _, app := range config.GetApps() {
		// If no Git repo exist
		if !isGit() {
			// Perfoming [build latest]
			pDebug("No local git repository found, just building latest")

			// Execute Pre commands
			if res := Pre(app); res != nil {
				pError("Pre execution returned non-zero status")
				return
			}

			// Build latest image
			res := buildImage(app, "latest", opts.Force)
			if res != nil {
				os.Exit(BuildFailed)
			}

			// Add additional user-defined Tag
			if opts.Tag != "" {
				tagImage(app, "latest", opts.Tag)
			}
		} else {
			// Skip build if there are no local changes and the commit is already built
			if !isDirty() && imageExist(app, rev) && !opts.Force {
				// Performing [skip rev|tag rev@latest|tag rev@branch]
				pInfo("Skipping build of %s:%s - image is already built", app.Image, rev)

				// Tag commit image
				tagImage(app, rev, "latest")

				// Tag branch image
				branches, err := getBranches(opts.All_branches)
				if err != nil {
					pError(err.Error())
					return
				}
				for _, branch := range branches {
					res := tagImage(app, rev, branch)
					if res != nil {
						os.Exit(TagFailed)
					}
					res = tagImage(app, rev, branch+"-"+rev)
					if res != nil {
						os.Exit(TagFailed)
					}
				}

				// Add additional user-defined Tag
				if opts.Tag != "" {
					tagImage(app, rev, opts.Tag)
				}
			} else {
				// Performing [build latest|tag latest@rev|tag latest@branch]

				// Execute Pre commands
				if res := Pre(app); res != nil {
					pError("Pre execution returned non-zero status")
				}

				// Build latest image
				res := buildImage(app, "latest", opts.Force)
				if res != nil {
					os.Exit(BuildFailed)
				}
				if isDirty() {
					pDebug("Skipping tag of %s:%s - local changes exist", app.Image, rev)
				} else {
					// Tag commit image
					tagImage(app, "latest", rev)

					// Tag branch image
					branches, err := getBranches(opts.All_branches)
					if err != nil {
						pError(err.Error())
						return
					}
					for _, branch := range branches {
						res := tagImage(app, "latest", branch)
						if res != nil {
							os.Exit(TagFailed)
						}
						res = tagImage(app, rev, branch+"-"+rev)
						if res != nil {
							os.Exit(TagFailed)
						}
					}

					// Add additional user-defined Tag
					if opts.Tag != "" {
						tagImage(app, rev, opts.Tag)
					}
				}
			}
		}

		// Execute Post commands
		if res := Post(app); res != nil {
			pError("Post execution returned non-zero status")
		}
	}
}

// Test function executes the tests of the project
func Test(opts BuildOptions) {
	config := opts.Config

	for _, app := range config.GetApps() {
		for _, value := range app.Test {
			pInfo("Running test command: %s", value)
			res := execute("bash", "-c", value)
			if res != nil {
				pError("Test execution returned non-zero status")
				os.Exit(ExecuteFailed)
			}
		}
	}
}

// Push function pushes the containers to the remote registry
func Push(opts BuildOptions) {
	config := opts.Config

	// If no Git repo exist
	if !isGit() {
		pError("No local git repository found, cannot push")
		os.Exit(NoGit)
	}

	if isDirty() {
		pError("Git repository has local changes, cannot push")
		os.Exit(GitDirty)
	}

	for _, app := range config.GetApps() {
		branches, err := getBranches(opts.All_branches)
		if err != nil {
			pError(err.Error())
			return
		}
		for _, branch := range branches {
			pInfo("Pushing image %s:%s", app.Image, "latest")
			if res := pushImage(app.Image, "latest"); res != nil {
				pError("Push returned non-zero status")
				os.Exit(ExecuteFailed)
			}
			if opts.Branch_tags {
				pInfo("Pushing image %s:%s", app.Image, branch)
				if res := pushImage(app.Image, branch); res != nil {
					pError("Push returned non-zero status")
					os.Exit(ExecuteFailed)
				}
			}
			if opts.Commit_tags {
				rev, err := getRevision(opts.Long_sha)
				if err != nil {
					pError(err.Error())
					return
				}
				pInfo("Pushing image %s:%s", app.Image, rev)
				if res := pushImage(app.Image, rev); res != nil {
					pError("Push returned non-zero status")
					os.Exit(ExecuteFailed)
				}
			}
			if opts.Branch_tags && opts.Commit_tags {
				rev, err := getRevision(opts.Long_sha)
				if err != nil {
					pError(err.Error())
					return
				}

				branchRevTag := branch + "-" + rev
				pInfo("Pushing image %s:%s", app.Image, branchRevTag)
				if res := pushImage(app.Image, branchRevTag); res != nil {
					pError("Push returned non-zero status")
					os.Exit(ExecuteFailed)
				}
			}

			// Add additional user-defined Tag
			if opts.Tag != "" {
				pInfo("Pushing image %s:%s", app.Image, opts.Tag)
				if res := pushImage(app.Image, opts.Tag); res != nil {
					pError("Push returned non-zero status")
					os.Exit(ExecuteFailed)
				}
			}
		}
	}
}

// Pull function pulls the containers from the remote registry
func Pull(opts BuildOptions) {
	config := opts.Config

	for _, app := range config.GetApps() {
		branches, err := getBranches(opts.All_branches)
		if err != nil {
			pError(err.Error())
			return
		}
		for _, branch := range branches {
			pInfo("Pulling image %s:%s", app.Image, "latest")
			if res := pullImage(app.Image, "latest"); res != nil {
				pError("Pull returned non-zero status")
				os.Exit(ExecuteFailed)
			}
			if opts.Branch_tags {
				pInfo("Pulling image %s:%s", app.Image, branch)
				if res := pullImage(app.Image, branch); res != nil {
					pError("Pull returned non-zero status")
					os.Exit(ExecuteFailed)
				}
			}
			if opts.Commit_tags {
				rev, err := getRevision(opts.Long_sha)
				if err != nil {
					pError(err.Error())
					return
				}

				pInfo("Pulling image %s:%s", app.Image, rev)
				if res := pullImage(app.Image, rev); res != nil {
					pError("Pull returned non-zero status")
					os.Exit(ExecuteFailed)
				}
			}
			if opts.Branch_tags && opts.Commit_tags {
				rev, err := getRevision(opts.Long_sha)
				if err != nil {
					pError(err.Error())
					return
				}

				branchRevTag := branch + "-" + rev
				pInfo("Pulling image %s:%s", app.Image, branchRevTag)
				if res := pullImage(app.Image, branchRevTag); res != nil {
					pError("Pull returned non-zero status")
					os.Exit(ExecuteFailed)
				}
			}

			// Add additional user-defined Tag
			if opts.Tag != "" {
				pInfo("Pulling image %s:%s", app.Image, opts.Tag)
				if res := pullImage(app.Image, opts.Tag); res != nil {
					pError("Pull returned non-zero status")
					os.Exit(ExecuteFailed)
				}
			}
		}
	}
}

// Purge function purges the stale images
func Purge(opts BuildOptions) {
	config := opts.Config

	// For each App
	for _, app := range config.GetApps() {
		var tags = []string{}

		// Retrieve the list of the existing Image tags
		for _, img := range getImages(app) {
			tags = append(tags, img.RepoTags...)
		}

		// Remove from the list: The latest image
		for key, tag := range tags {
			if tag == app.Image+":latest" {
				tags = append(tags[:key], tags[key+1:]...)
			}
		}

		// Remove from the list: The current commit-id
		for key, tag := range tags {
			rev, err := getRevision(opts.Long_sha)
			if err != nil {
				pError(err.Error())
				return
			}
			if tag == app.Image+":"+rev {
				tags = append(tags[:key], tags[key+1:]...)
			}
		}

		// Remove from the list: The working-dir git branches
		branches, err := getBranches(opts.All_branches)
		if err != nil {
			pError(err.Error())
			return
		}

		for _, branch := range branches {
			for key, tag := range tags {
				if tag == app.Image+":"+branch {
					tags = append(tags[:key], tags[key+1:]...)
				}
				if strings.Contains(tag, app.Image+":"+branch+"-") {
					tags = append(tags[:key], tags[key+1:]...)
				}
			}
		}

		// Proceed with deletion of Images
		for _, tag := range tags {
			pInfo("Deleting image %s", tag)
			res := removeImage(tag)
			if res != nil {
				pError("Deleting image failed: %s", res)
				os.Exit(DeleteImageFailed)
			}
		}
	}
}
