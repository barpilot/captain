package captain // import "github.com/harbur/captain"

import (
	"fmt"
	"os"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func getRepository() (*git.Repository, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	pInfo(dir)

	return git.PlainOpen(dir)
}

func getRevision(longSha bool) (string, error) {
	r, err := getRepository()
	if err != nil {
		return "", err
	}

	h, err := r.ResolveRevision(plumbing.Revision("HEAD"))
	return h.String(), err

	// params := []string{"rev-parse"}
	// if !long_sha {
	// 	params = append(params, "--short")
	// }

	// params = append(params, "HEAD")
	// res, _ := oneliner("git", params...)
	// return "", res
}

func getBranches(all_branches bool) ([]string, error) {
	// Labels (branches + tags)
	var labels = []string{}

	r, err := getRepository()
	if err != nil {
		return labels, err
	}

	tagrefs, err := r.Tags()
	if err != nil {
		return labels, err
	}

	err = tagrefs.ForEach(func(t *plumbing.Reference) error {
		labels = append(labels, t.Name().Short())
		fmt.Println(t)
		return nil
	})
	if err != nil {
		return labels, err
	}

	brancherefs, err := r.Branches()

	err = brancherefs.ForEach(func(t *plumbing.Reference) error {
		labels = append(labels, t.Name().Short())
		fmt.Println(t)
		return nil
	})
	if err != nil {
		return labels, err
	}

	return labels, nil

	// branches_str, _ := oneliner("git", "rev-parse", "--symbolic-full-name", "--abbrev-ref", "HEAD")
	// if all_branches {
	// 	branches_str, _ = oneliner("git", "branch", "--no-column", "--contains", "HEAD")
	// }

	// var branches = make([]string, 5)
	// if branches_str != "" {
	// 	// Remove asterisk from branches list
	// 	r := regexp.MustCompile("[\\* ] ")
	// 	branches_str = r.ReplaceAllString(branches_str, "")
	// 	branches = strings.Split(branches_str, "\n")

	// 	// Branches list is separated by spaces. Let's put it in an array
	// 	labels = append(labels, branches...)
	// }

	// tags_str, _ := oneliner("git", "tag", "--points-at", "HEAD")

	// if tags_str != "" {
	// 	tags := strings.Split(tags_str, "\n")
	// 	pDebug("Active branches %s and tags %s", branches, tags)
	// 	// Git tag list is separated by multi-lines. Let's put it in an array
	// 	labels = append(labels, tags...)
	// }

	// for key := range labels {
	// 	// Remove start of "heads/origin" if exist
	// 	r := regexp.MustCompile("^heads\\/origin\\/")
	// 	labels[key] = r.ReplaceAllString(labels[key], "")

	// 	// Remove start of "remotes/origin" if exist
	// 	r = regexp.MustCompile("^remotes\\/origin\\/")
	// 	labels[key] = r.ReplaceAllString(labels[key], "")

	// 	// Replace all "/" with "."
	// 	labels[key] = strings.Replace(labels[key], "/", ".", -1)

	// 	// Replace all "~" with "."
	// 	labels[key] = strings.Replace(labels[key], "~", ".", -1)
	// }

	// return nil, labels
}

func isDirty() bool {

	r, err := getRepository()
	if err != nil {
		return true
	}

	w, err := r.Worktree()
	if err != nil {
		return true
	}

	status, err := w.Status()
	if err != nil {
		return true
	}

	return !status.IsClean()

	// res, _ := oneliner("git", "status", "--porcelain")
	// return len(res) > 0
}

func isGit() bool {
	_, err := getRepository()
	if err != nil {
		return false
	}
	return true
}
