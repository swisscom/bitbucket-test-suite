package acceptance

import (
    "net/http"
    "encoding/json"
    "bytes"
	"fmt"
	"os"
	"testing"
	"time"
	"github.com/DATA-DOG/godog"

    "gopkg.in/src-d/go-git.v4"
	"path/filepath"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"io/ioutil"
)

const http_url_key = "BB_HTTP_URL"
const ssh_url_key = "BB_SSH_URL"
const project_key = "BB_PROJECT"
const user_key = "BB_USER"
const password_key = "BB_PASSWORD"
const repository_key = "BB_REPOSITORY"
var http_url string
var ssh_url string
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
    http_url = os.Getenv(http_url_key)
    exitIfNotSet(http_url, http_url_key)
	ssh_url = os.Getenv(ssh_url_key)
	exitIfNotSet(ssh_url, ssh_url_key)
    project = os.Getenv(project_key)
    exitIfNotSet(project, project_key)
    user = os.Getenv(user_key)
    exitIfNotSet(user, user_key)
    password = os.Getenv(password_key)
    exitIfNotSet(password, password_key)
    repository = os.Getenv(repository_key)
    exitIfNotSet(repository, repository_key)
    urlRepos = http_url + "/rest/api/1.0/projects/" + project + "/repos"

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
	Name    	string `json:"name"`
	ScmId   	string `json:"scmId"`
	Forkable   	bool `json:"forkable"`
}

func createRepository(repositoryName string) error {
	repositoryCreation := RepositoryCreation{
		Name: repositoryName,
		ScmId: "git",
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

    Info("created repository [%s] with status [%s]", repositoryName, resp.Status)
	return nil
}

type Repository struct {
	Slug    string `json:"slug"`
    ScmId   string `json:"scmId"`
}

func checkRepository(repositoryName string) error {
	urlRepoTest := urlRepos + "/" + repositoryName
	fmt.Println("http_url test repo: " + urlRepoTest)
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
    if err != nil{
        panic(err)
    }
    fmt.Println("got the repository: ", repository.Slug)
    return nil
}

func deleteRepository(repositoryName string) error {
	urlRepoTest := urlRepos + "/" + repositoryName
    fmt.Println("http_url test repo: " + urlRepoTest)
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
    fmt.Println("deleted repo with response status:", resp.Status)
    return nil
}

func cloneRepo(repositoryName string) error {
	os.RemoveAll("/tmp/bitbucket_test")
	sshUrlRepository := ssh_url + "/" + project + "/" + repositoryName + ".git"
    r, _ := git.PlainClone("/tmp/bitbucket_test", false, &git.CloneOptions{
        URL:      sshUrlRepository,
        Progress: os.Stdout,
    })

    fmt.Println(r.Remotes())

    return nil

}

func commitFile() error {

	r, err := git.PlainOpen("/tmp/bitbucket_test")
	CheckIfError(err)

	w, err := r.Worktree()
	CheckIfError(err)

	filename := filepath.Join("/tmp/bitbucket_test", "example-git-file")
	err = ioutil.WriteFile(filename, []byte("this is test"), 0644)
	CheckIfError(err)

	// Adds the new file to the staging area.
	Info("git add example-git-file")
	_, err = w.Add("example-git-file")
	CheckIfError(err)

	// We can verify the current status of the worktree using the method Status.
	Info("git status --porcelain")
	status, err := w.Status()
	CheckIfError(err)

	fmt.Println(status)

	// Commits the current staging are to the repository, with the new file
	// just created. We should provide the object.Signature of Author of the
	// commit.
	Info("git commit -m \"example go-git commit\"")
	commit, err := w.Commit("example go-git commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "BitBucket Test",
			Email: "bitbuckettest@swisscom.com",
			When:  time.Now(),
		},
	})

	CheckIfError(err)

	// Prints the current HEAD to verify that all worked well.
	Info("git show -s")
	obj, err := r.CommitObject(commit)
	CheckIfError(err)

	fmt.Println(obj)

	return nil
}

func push() error {

	r, err := git.PlainOpen("/tmp/bitbucket_test")
	CheckIfError(err)

	Info("git push")
	// push using default options
	err = r.Push(&git.PushOptions{})
	CheckIfError(err)

	return nil
}

type Commit struct {
	Id    string `json:"id"`
}

func compareCommit(repositoryName string) error {

	r, err := git.PlainOpen("/tmp/bitbucket_test")
	CheckIfError(err)

	commitItr, err := r.CommitObjects()
	CheckIfError(err)

	commitLocal, err := commitItr.Next()
	CheckIfError(err)

	commitId :=	commitLocal.ID()
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
    if err != nil{
        panic(err)
    }
    fmt.Println("got remote commit: ", commitRemote.Id)
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
	s.Step(`^the repository ([A-Za-z_-]+) exists$`, createRepository)
	s.Step(`^clone the ([A-Za-z_-]+)$`, cloneRepo)
	s.Step(`^commit a file$`, commitFile)
	s.Step(`^push to remote`, push)
	s.Step(`^the commit should be visible in repository ([A-Za-z_-]+)$`, compareCommit)
}
