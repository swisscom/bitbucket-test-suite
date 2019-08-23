package acceptance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/DATA-DOG/godog"
	"net/http"
	"os"
	"testing"
	"time"
	"errors"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"io/ioutil"
	"path/filepath"
	"os/exec"
)

const httpUrlKey = "BB_HTTP_URL"
const sshUrlKey = "BB_SSH_URL"
const projectKey = "BB_PROJECT"
const userKey = "BB_CREDENTIALS_USR"
const passwordKey = "BB_CREDENTIALS_PWD"
const repositoryKey = "BB_REPOSITORY"
const cloneDir = "/tmp/bitbucket_test"

var httpUrl string
var sshUrl string
var project string
var user string
var password string
var repository string
var urlRepos string

func TestMain(m *testing.M) {
	exitIfNotSet := func(env string, name string) {
		if len(env) == 0 {
			fmt.Println(name + " not set")
			os.Exit(1)
		}
	}
	httpUrl = os.Getenv(httpUrlKey)
	exitIfNotSet(httpUrl, httpUrlKey)
	sshUrl = os.Getenv(sshUrlKey)
	exitIfNotSet(sshUrl, sshUrlKey)
	project = os.Getenv(projectKey)
	exitIfNotSet(project, projectKey)
	user = os.Getenv(userKey)
	exitIfNotSet(user, userKey)
	password = os.Getenv(passwordKey)
	exitIfNotSet(password, passwordKey)
	repository = os.Getenv(repositoryKey)
	exitIfNotSet(repository, repositoryKey)
	urlRepos = httpUrl + "/rest/api/1.0/projects/" + project + "/repos"

	status := godog.RunWithOptions("repositories", func(s *godog.Suite) {
		FeatureContext(s)
	}, godog.Options{
		Format:    "pretty",
		Paths:     []string{"features"},
		Randomize: time.Now().UTC().UnixNano(), // randomize scenario execution order
	})

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

type RepositoryCreation struct {
	Name     string `json:"name"`
	ScmId    string `json:"scmId"`
	Forkable bool   `json:"forkable"`
}

func createRepository(repositoryName string) error {
	repositoryCreation := RepositoryCreation{
		Name:     repositoryName,
		ScmId:    "git",
		Forkable: true,
	}
	repositoryCreationJson, _ := json.Marshal(repositoryCreation)
	req, err := http.NewRequest("POST", urlRepos, bytes.NewBuffer(repositoryCreationJson))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "golang")
	req.SetBasicAuth(user, password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	Info("[createRepository] created repository [%s] with status [%s]", repositoryName, resp.Status)
	return nil
}

type Repository struct {
	Slug  string `json:"slug"`
	ScmId string `json:"scmId"`
}

func checkRepository(repositoryName string) error {
	urlRepoTest := urlRepos + "/" + repositoryName
	req, err := http.NewRequest("GET", urlRepoTest, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "golang")
	req.SetBasicAuth(user, password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	repository := Repository{}
	err = json.NewDecoder(resp.Body).Decode(&repository)
	if err != nil {
		panic(err)
	}
	Info("[checkRepository] got the repository [%s]", repository.Slug)
	return nil
}

func deleteRepository(repositoryName string) error {
	urlRepoTest := urlRepos + "/" + repositoryName
	req, err := http.NewRequest("DELETE", urlRepoTest, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "golang")
	req.SetBasicAuth(user, password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	Info("[deleteRepository] deleted repo [%s] with response status [%s]", repositoryName, resp.Status)
	return nil
}

func cloneRepository(repositoryName string) error {
	os.RemoveAll(cloneDir)
	sshUrlRepository := sshUrl + "/" + project + "/" + repositoryName + ".git"
	git.PlainClone(cloneDir, false, &git.CloneOptions{
		URL:      sshUrlRepository,
		Progress: os.Stdout,
	})

	// we don't chech the error here, because an empty repository returns an empty repository error
	// CheckIfError(err)

	Info("[cloneRepository] successfully cloned repo [%s]", repositoryName)

	return nil

}

func commitFile() error {

	r, err := git.PlainOpen(cloneDir)
	CheckIfError(err)

	w, err := r.Worktree()
	CheckIfError(err)

	filename := filepath.Join(cloneDir, "example-git-file")
	err = ioutil.WriteFile(filename, []byte("this is test"), 0644)
	CheckIfError(err)

	// Adds the new file to the staging area.
	Info("[commitFile] git add example-git-file")
	_, err = w.Add("example-git-file")
	CheckIfError(err)

	// Commits the current staging are to the repository, with the new file
	// just created. We should provide the object.Signature of Author of the
	// commit.
	Info("[commitFile] git commit -m \"example go-git commit\"")
	commit, err := w.Commit("example go-git commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "BitBucket Test",
			Email: "bitbuckettest@swisscom.com",
			When:  time.Now(),
		},
	})

	CheckIfError(err)

	// Prints the current HEAD to verify that all worked well.
	obj, err := r.CommitObject(commit)
	CheckIfError(err)

	Info("[commitFile] got commit [%s]", obj)

	return nil
}

// fallback to git command because of a bug in go-git
// https://github.com/src-d/go-git/issues/637
func pushRepository() error {
	cmd := exec.Command("git", "push", "--set-upstream", "origin", "master")
	cmd.Dir = cloneDir
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if len(out.String()) > 0 {
		Info("[pushRepository] " + out.String())
	}
	if len(stderr.String()) > 0 {
		Warning("[pushRepository] " + stderr.String())
	}
	CheckIfError(err)
	return nil
}

type Commit struct {
	Id string `json:"id"`
}

func compareCommit(repositoryName string) error {

	r, err := git.PlainOpen(cloneDir)
	CheckIfError(err)

	commitItr, err := r.CommitObjects()
	CheckIfError(err)

	commitLocal, err := commitItr.Next()
	CheckIfError(err)

	commitId := commitLocal.ID()
	commitUrl := urlRepos + "/" + repositoryName + "/commits/" + commitId.String()

	req, err := http.NewRequest("GET", commitUrl, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "golang")
	req.SetBasicAuth(user, password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	commitRemote := Commit{}
	err = json.NewDecoder(resp.Body).Decode(&commitRemote)
	if err != nil {
		panic(err)
	}
	Info("[compareCommit] local commit id [%s], remote commit id [%s] ", commitId.String(), commitRemote.Id)

	if commitId.String() != commitRemote.Id {
		Warning("[compareCommit] commit id's don't match")
		err := errors.New("commit id's don't match")
		panic(err)
	}

	return nil
}

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

// Info should be used to describe the example commands that are about to run.
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// Warning should be used to display a warning
func Warning(format string, args ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

func FeatureContext(s *godog.Suite) {
	s.Step(`^the repository ([A-Za-z_-]+) doesnt exist$`, deleteRepository)
	s.Step(`^I create repository ([A-Za-z_-]+)$`, createRepository)
	s.Step(`^repository ([A-Za-z_-]+) should be accessible$`, checkRepository)
	s.Step(`^clone the ([A-Za-z_-]+)$`, cloneRepository)
	s.Step(`^commit a file$`, commitFile)
	s.Step(`^push to remote`, pushRepository)
	s.Step(`^the commit should be visible in repository ([A-Za-z_-]+)$`, compareCommit)
}
