/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	githubOutputLog = ctrl.Log.WithName("outputs").WithName("github")
	once            = new(sync.Once)

	personalAccessToken string
	mu                  sync.Mutex
)

func init() {
	once.Do(func() {
		personalAccessToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	})
}

// GitHubOutput defines the spec for integrating with GitHub
type GitHubOutput struct {
	// +kubebuilder:validation:Required

	Config GitHubConfig `json:"config"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format:=string

	LocalFilePath string `json:"localFilePath"`
}

type GitHubConfig struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format:=string

	RepositoryURL string `json:"repositoryUrl"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format:=string

	BaseBranch string `json:"baseBranch"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format:=string

	ManifestPath string `json:"manifestPath"`

	// +kubebuilder:validation:Required

	Author Author `json:"author"`
}

type Author struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format:=string

	Name string `json:"name"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format:=string

	Email string `json:"email"`
}

func (o *GitHubOutput) Setup() error {
	if err := o.clone(); err != nil {
		return err
	}

	r, err := o.open()
	if err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()

	bb := o.Config.BaseBranch
	return o.checkout(r, bb)
}

func (o *GitHubOutput) Publish(name string, manifest []byte) (err error) {
	r, err := o.open()
	if err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()

	if err = o.pull(r); err != nil {
		return err
	}

	nb := fmt.Sprintf("manifest-capturer-%s", generateTimestamp())
	bb := o.Config.BaseBranch
	if err = o.branch(r, nb); err != nil {
		return err
	}

	if err = o.checkout(r, nb); err != nil {
		return err
	}
	defer func() {
		err = o.checkout(r, bb)
	}()

	if err = o.commit(r, name, manifest); err != nil {
		return err
	}

	if err = o.push(r); err != nil {
		return err
	}

	return nil
}

func (o *GitHubOutput) clone() error {
	url := o.Config.RepositoryURL
	directory := o.LocalFilePath

	_, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL:               url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		if err.Error() == git.ErrRepositoryAlreadyExists.Error() {
			return nil
		}

		githubOutputLog.Error(err, "failed `git clone %s %s --recursive`", url, directory)
		return err
	}

	return nil
}

func (o *GitHubOutput) open() (*git.Repository, error) {
	directory := o.LocalFilePath
	r, err := git.PlainOpen(directory)
	if err != nil {
		githubOutputLog.Error(err, "failed to open local repository on %s", "directory", directory)
		return nil, err
	}

	return r, nil
}

func (o *GitHubOutput) branch(r *git.Repository, branch string) error {
	headRef, err := r.Head()
	if err != nil {
		githubOutputLog.Error(err, "failed to fetch HEAD ref")
		return err
	}

	ref := plumbing.NewHashReference(
		plumbing.NewBranchReferenceName(branch),
		headRef.Hash(),
	)

	if err = r.Storer.SetReference(ref); err != nil {
		githubOutputLog.Error(err, "failed `git branch <branch>`")
		return err
	}

	return nil
}

func (o *GitHubOutput) checkout(r *git.Repository, branch string) error {
	w, err := r.Worktree()
	if err != nil {
		githubOutputLog.Error(err, "failed to open worktree")
		return err
	}

	if err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Force:  true,
	}); err != nil {
		githubOutputLog.Error(err, "failed `git checkout <branch>`")
		return err
	}

	return nil
}

func (o *GitHubOutput) pull(r *git.Repository) error {
	w, err := r.Worktree()
	if err != nil {
		githubOutputLog.Error(err, "failed to open worktree")
		return err
	}

	if err = w.Pull(&git.PullOptions{RemoteName: "origin"}); err != nil {
		if err != git.NoErrAlreadyUpToDate {
			githubOutputLog.Error(err, "failed `git pull origin`")
			return err
		}
	}

	return nil
}

func (o *GitHubOutput) commit(r *git.Repository, name string, manifest []byte) error {
	w, err := r.Worktree()
	if err != nil {
		githubOutputLog.Error(err, "failed to open worktree")
		return err
	}

	directory := o.LocalFilePath
	manifestPath := o.Config.ManifestPath
	filename := filepath.Join(directory, manifestPath)

	header := []byte(fmt.Sprintf("# this file is generated by manifest-capturer by %s\n\n", name))
	content := append(header[:], manifest[:]...)
	if err = ioutil.WriteFile(filename, content, 0644); err != nil {
		githubOutputLog.Error(err, "failed to write file", filename)
		return err
	}

	_, err = w.Add(manifestPath)
	if err != nil {
		githubOutputLog.Error(err, "failed `git add`", "filename", filename)
		return err

	}

	author := o.Config.Author
	msg := "update manifest"
	_, err = w.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  author.Name,
			Email: author.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		githubOutputLog.Error(err, "failed `git commit -m`", "messsage", msg)
		return err
	}

	return nil
}

func (o *GitHubOutput) push(r *git.Repository) error {
	author := o.Config.Author
	if err := r.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: author.Name,
			Password: personalAccessToken,
		},
	}); err != nil {
		githubOutputLog.Error(err, "failed `git push`")
		return err
	}

	return nil
}

func generateTimestamp() string {
	t := time.Now()
	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	return fmt.Sprintf("%04d%02d%02d%02d%02d%02d", year, int(month), day, hour, min, sec)
}
