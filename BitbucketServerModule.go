package main

import (
	"encoding/json"
	"fmt"
	"github.com/sne11ius/bitclient"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

type BitbucketServerModule struct {
	httpUrl                string
	username               string
	password               string
	repositoriesWithReadme []BitbucketServerRepositoryWithReadme
}

type BitbucketServerRepositoryWithReadme struct {
	Repository bitclient.Repository `json:"repository"`
	Readme     string               `json:"readme"`
}

func NewBitbucketServerModule() *BitbucketServerModule {
	return new(BitbucketServerModule)
}

func (b *BitbucketServerModule) Name() string {
	return "BitbucketServer"
}

func (b *BitbucketServerModule) Description() string {
	return "Provides access to bitbucket server projects"
}

func (b *BitbucketServerModule) CanBeDisabled() bool {
	return true
}

func (b *BitbucketServerModule) UpdateSettings() {
	configKey := b.Name() + ".http-url"
	if !viper.IsSet(configKey) {
		log.Fatalf("Missing configuration key `%s` (eg. 'https://atlassian.example.com/bitbucket')", configKey)
	}
	b.httpUrl = viper.GetString(configKey)

	configKey = b.Name() + ".username"
	if !viper.IsSet(configKey) {
		log.Fatalf("Missing configuration key `%s` (eg. 'myusername')", configKey)
	}
	b.username = viper.GetString(configKey)

	configKey = b.Name() + ".password"
	if !viper.IsSet(configKey) {
		log.Fatalf("Missing configuration key `%s` (eg. 'mypassword')", configKey)
	}
	b.password = viper.GetString(configKey)
}

func (b *BitbucketServerModule) NeedsExternalData() bool {
	return true
}

func (b *BitbucketServerModule) UpdateExternalData() {
	allRepositories := b.LoadRepositories()
	b.repositoriesWithReadme = b.LoadReadmes(allRepositories)
	fmt.Printf("  - Updated %d projects\n", len(b.repositoriesWithReadme))
}

func (b BitbucketServerModule) LoadRepositories() []bitclient.Repository {
	client := bitclient.NewBitClient(b.httpUrl, b.username, b.password)

	requestParams := bitclient.PagedRequest{
		Limit: 10000,
		Start: 0,
	}
	projectsResponse, err := client.GetProjects(requestParams)

	if err != nil {
		log.Fatalf("Cannot list projects: %v", err)
	}

	projects := projectsResponse.Values

	numProjects := len(projects)
	var wg sync.WaitGroup
	wg.Add(numProjects)
	queue := make(chan []bitclient.Repository, 1)
	for i := 0; i < numProjects; i++ {
		go func(i int) {
			project := projects[i]
			repositories, err := client.GetRepositories(project.Key, bitclient.PagedRequest{
				Limit: 1000,
				Start: 0,
			})
			if err != nil {
				log.Fatalf("Cannot get repositories for %s", project.Name)
			}
			queue <- repositories.Values
		}(i)
	}
	var allRepositories []bitclient.Repository
	go func() {
		for t := range queue {
			allRepositories = append(allRepositories, t...)
			wg.Done()
		}
	}()
	wg.Wait()
	return allRepositories
}

func (b *BitbucketServerModule) LoadReadmes(repositories []bitclient.Repository) []BitbucketServerRepositoryWithReadme {
	numRepositories := len(repositories)
	var wg sync.WaitGroup
	wg.Add(numRepositories)
	queue := make(chan BitbucketServerRepositoryWithReadme, 1)
	for i := 0; i < numRepositories; i++ {
		go func(i int) {
			repository := repositories[i]
			readme, err := b.GetReadmeText(repository, "develop")
			if err != nil {
				readme, _ = b.GetReadmeText(repository, "master")
			}
			queue <- BitbucketServerRepositoryWithReadme{
				Repository: repository,
				Readme:     readme,
			}
		}(i)
	}
	var repositoriesWithReadmes []BitbucketServerRepositoryWithReadme
	go func() {
		for t := range queue {
			repositoriesWithReadmes = append(repositoriesWithReadmes, t)
			wg.Done()
		}
	}()
	wg.Wait()
	return repositoriesWithReadmes
}

func (b *BitbucketServerModule) GetReadmeText(repository bitclient.Repository, branch string) (string, error) {
	var selfLink = repository.Links["self"][0]["href"]
	var baseLink = strings.TrimSuffix(selfLink, "/browse")
	var readmeLink = baseLink + "/raw/README.md?at=refs%2Fheads%2F" + branch
	client := &http.Client{}
	req, err := http.NewRequest("GET", readmeLink, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(b.username, b.password)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatalf("Could not close response body: %v", err)
		}
	}()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		} else {
			return string(bodyBytes), nil
		}
	} else {
		return "", fmt.Errorf("could not get README. Status: %v", resp.StatusCode)
	}
}

func (b *BitbucketServerModule) WriteExternalData(file *os.File) {
	bytes, err := json.Marshal(b.repositoriesWithReadme)
	if err != nil {
		log.Fatalf("Cannot serialize repository data: %s", err)
	}
	if _, err = file.Write(bytes); err != nil {
		log.Fatalf("Cannot write repository data to %v: %s", file, err)
	}
}

func (b *BitbucketServerModule) ReadExternalData(data []byte) {
	err := json.Unmarshal(data, &b.repositoriesWithReadme)
	if err != nil {
		log.Fatalf("Cannot read bitbucket project cache. Consider running with `-u` parameter.")
	}
}

type BitbucketServerBrowseAction struct {
	repo BitbucketServerRepositoryWithReadme
}

func (b BitbucketServerBrowseAction) GetLabel() string {
	return "[bitbucket[] BROWSE " + b.repo.Repository.Name
}

func (b BitbucketServerBrowseAction) Run() string {
	url := b.repo.Repository.Links["self"][0]["href"]
	if err := launchUrl(url); err != nil {
		log.Fatalf("Could not browse %s: %v", url, err)
	}
	return "Opened " + url
}

type BitbucketServerCloneAction struct {
	repo BitbucketServerRepositoryWithReadme
}

func (b BitbucketServerCloneAction) GetLabel() string {
	return "[bitbucket[] CLONE " + b.repo.Repository.Name
}

func (b BitbucketServerCloneAction) Run() string {
	var cloneUrl string
	found := false
	for _, link := range b.repo.Repository.Links["clone"] {
		if strings.HasPrefix(link["href"], "ssh://") {
			cloneUrl = link["href"]
			found = true
			break
		}
	}
	if !found {
		cloneUrl = b.repo.Repository.Links["clone"][0]["href"]
	}
	if err := runGitClone(cloneUrl); err != nil {
		log.Fatalf("Could not clone %s: %v", cloneUrl, err)
	}
	return "Cloned " + cloneUrl
}

func (b *BitbucketServerModule) CreateActions(tags []Tag) []action {
	var actions []action
	for _, repo := range b.repositoriesWithReadme {
		strs := []string{"bitbucket", "browse", repo.Repository.Name, repo.Readme, repo.Repository.Project.Name}
		if DoMatch(strs, tags) {
			actions = append(actions, BitbucketServerBrowseAction{repo: repo})
		}
		strs = []string{"bitbucket", "clone", repo.Repository.Name, repo.Readme, repo.Repository.Project.Name}
		if DoMatch(strs, tags) {
			actions = append(actions, BitbucketServerCloneAction{repo: repo})
		}
	}
	return actions
}
