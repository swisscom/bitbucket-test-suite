package acceptance

import (
    //"errors"
    "net/http"
    "encoding/json"
    "bytes"
    //"io/ioutil"
	"fmt"
	"os"
	"testing"
	"time"
	"github.com/DATA-DOG/godog"
)

const url_key = "BB_URL"
const project_key = "BB_PROJECT"
const user_key = "BB_USER"
const password_key = "BB_PASSWORD"
const repository_key = "BB_REPOSITORY"
var url string
var project string
var user string
var password string
var repository string
var urlRepos string
var urlTestRepo string

func TestMain(m *testing.M) {
    exitIfNotSet := func(env string, name string) {
        if len(env) == 0 {
            fmt.Println(name + " not set")
            os.Exit(1)
        }
    }
    url = os.Getenv(url_key)
    exitIfNotSet(url, url_key)
    project = os.Getenv(project_key)
    exitIfNotSet(project, project_key)
    user = os.Getenv(user_key)
    exitIfNotSet(user, user_key)
    password = os.Getenv(password_key)
    exitIfNotSet(password, password_key)
    repository = os.Getenv(repository_key)
    exitIfNotSet(repository, repository_key)
    urlRepos = url + "/rest/api/1.0/projects/PLAYG/repos"
    urlTestRepo = url + "/rest/api/1.0/projects/PLAYG/repos/" + repository
    if len(url) == 0 {
        fmt.Println("BB_URL not set")
        os.Exit(1)
    }
	status := godog.RunWithOptions("repositories", func(s *godog.Suite) {
		FeatureContext(s)
	}, godog.Options{
		Format:    "progress",
		Paths:     []string{"features"},
		Randomize: time.Now().UTC().UnixNano(), // randomize scenario execution order
	})

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

func createRepository() error {
    var jsonStr = []byte(`{ "name": "test_repo", "scmId": "git", "forkable": true }`)
    req, err := http.NewRequest("POST", urlRepos, bytes.NewBuffer(jsonStr))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("User-Agent", "golang")
    req.SetBasicAuth(user, password)
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    fmt.Println("created repository with status: ", resp.Status)
	return nil
}

type Repository struct {
	Slug    string `json:"slug"`
    ScmId   string `json:"scmId"`
}

func checkRepository() error {
    req, err := http.NewRequest("GET", urlTestRepo, nil)
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

func deleteRepository() error {
    req, err := http.NewRequest("DELETE", urlTestRepo, nil)
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


func FeatureContext(s *godog.Suite) {
	s.Step(`^the repository test_repo doesnt exist$`, deleteRepository)
	s.Step(`^I create repository test_repo$`, createRepository)
	s.Step(`^repository test_repo should be accessible$`, checkRepository)
}
