package captain // import "github.com/harbur/captain"

import (
	"os"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func getRepository() (*git.Repository, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	opt := &git.PlainOpenOptions{DetectDotGit: true}

	return git.PlainOpenWithOptions(dir, opt)
}

func getRevision(longSha bool) (string, error) {
	r, err := getRepository()
	if err != nil {
		return "", err
	}

	h, err := getCurrentCommitFromRepository(r)
	if longSha || err != nil {
		return h, err
	}
	return h[:7], err
}

func getBranches(all_branches bool) ([]string, error) {
	// Labels (branches + tags)
	var labels = []string{}

	r, err := getRepository()
	if err != nil {
		return labels, err
	}

	branches, err := getCurrentBranchesFromRepository(r)
	if err != nil {
		return labels, err
	}

	if all_branches {
		for _, branch := range branches {
			labels = append(labels, branch)
		}
	} else {
		labels = append(labels, branches[0])
	}

	tags, err := getCurrentTagsFromRepository(r)
	for _, tag := range tags {
		labels = append(labels, tag)
	}

	return labels, err
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

// Thanks King'ori Maina @itskingori
// https://github.com/src-d/go-git/issues/1030#issuecomment-443679681

func getCurrentBranchesFromRepository(repository *git.Repository) ([]string, error) {
	var currentBranchesNames []string

	branchRefs, err := repository.Branches()
	if err != nil {
		return currentBranchesNames, err
	}

	headRef, err := repository.Head()
	if err != nil {
		return currentBranchesNames, err
	}

	err = branchRefs.ForEach(func(branchRef *plumbing.Reference) error {
		if branchRef.Hash() == headRef.Hash() {
			currentBranchesNames = append(currentBranchesNames, branchRef.Name().Short())

			return nil
		}

		return nil
	})
	if err != nil {
		return currentBranchesNames, err
	}

	return currentBranchesNames, nil
}

func getCurrentCommitFromRepository(repository *git.Repository) (string, error) {
	headRef, err := repository.Head()
	if err != nil {
		return "", err
	}
	headSha := headRef.Hash().String()

	return headSha, nil
}

func getCurrentTagsFromRepository(repository *git.Repository) ([]string, error) {
	var currentTagsNames []string

	tagRefs, err := repository.Branches()
	if err != nil {
		return currentTagsNames, err
	}

	headRef, err := repository.Head()
	if err != nil {
		return currentTagsNames, err
	}

	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		if tagRef.Hash() == headRef.Hash() {
			currentTagsNames = append(currentTagsNames, tagRef.Name().Short())

			return nil
		}

		return nil
	})
	if err != nil {
		return currentTagsNames, err
	}

	return currentTagsNames, nil
}
