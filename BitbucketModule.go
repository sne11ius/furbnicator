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

type BitbucketModule struct {
	gitUrl                 string
	httpUrl                string
	username               string
	password               string
	repositoriesWithReadme []RepositoryWithReadme
}

type RepositoryWithReadme struct {
	Repository bitclient.Repository `json:"repository"`
	Readme     string               `json:"readme"`
}

func NewBitbucketModule() *BitbucketModule {
	return new(BitbucketModule)
}

func (b *BitbucketModule) Name() string {
	return "Bitbucket"
}

func (b *BitbucketModule) Description() string {
	return "Provides access to bitbucket projects"
}

func (b *BitbucketModule) CanBeDisabled() bool {
	return true
}

func (b *BitbucketModule) UpdateSettings() {
	configKey := b.Name() + ".git-url"
	if !viper.IsSet(configKey) {
		log.Fatalf("Missing configuration key `%s` (eg. 'atlassian.example.com:7998')", configKey)
	}
	b.gitUrl = viper.GetString(configKey)

	configKey = b.Name() + ".http-url"
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

func (b *BitbucketModule) NeedsExternalData() bool {
	return true
}

func (b *BitbucketModule) UpdateExternalData() {
	allRepositories := b.LoadRepositories()
	b.repositoriesWithReadme = b.LoadReadmes(allRepositories)
	fmt.Printf("  - Updated %d projects\n", len(b.repositoriesWithReadme))
}

func (b BitbucketModule) LoadRepositories() []bitclient.Repository {
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

func (b *BitbucketModule) LoadReadmes(repositories []bitclient.Repository) []RepositoryWithReadme {
	numRepositories := len(repositories)
	var wg sync.WaitGroup
	wg.Add(numRepositories)
	queue := make(chan RepositoryWithReadme, 1)
	for i := 0; i < numRepositories; i++ {
		go func(i int) {
			repository := repositories[i]
			readme, err := b.GetReadmeText(repository, "develop")
			if err != nil {
				readme, _ = b.GetReadmeText(repository, "master")
			}
			queue <- RepositoryWithReadme{
				Repository: repository,
				Readme:     readme,
			}
		}(i)
	}
	var repositoriesWithReadmes []RepositoryWithReadme
	go func() {
		for t := range queue {
			repositoriesWithReadmes = append(repositoriesWithReadmes, t)
			wg.Done()
		}
	}()
	wg.Wait()
	return repositoriesWithReadmes
}

func (b *BitbucketModule) GetReadmeText(repository bitclient.Repository, branch string) (string, error) {
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

func (b *BitbucketModule) WriteExternalData(file *os.File) {
	bytes, err := json.Marshal(b.repositoriesWithReadme)
	if err != nil {
		log.Fatalf("Cannot serialize repository data: %s", err)
	}
	if _, err = file.Write(bytes); err != nil {
		log.Fatalf("Cannot write repository data to %v: %s", file, err)
	}
}

func (b *BitbucketModule) ReadExternalData(data []byte) {
	err := json.Unmarshal(data, &b.repositoriesWithReadme)
	if err != nil {
		log.Fatalf("Cannot read bitbucket project cache. Consider running with `-u` parameter.")
	}
}

type BitbucketBrowseAction struct {
	repo RepositoryWithReadme
}

func (b BitbucketBrowseAction) GetLabel() string {
	return "[bitbucket] BROWSE " + b.repo.Repository.Name
}

func (b BitbucketBrowseAction) Run() string {
	url := b.repo.Repository.Links["self"][0]["href"]
	if err := LaunchUrl(url); err != nil {
		log.Fatalf("Could not browse %s: %v", url, err)
	}
	return "Opened " + url
}

func (b *BitbucketModule) CreateActions(tags []string) []action {
	var actions []action
	for _, repo := range b.repositoriesWithReadme {
		for _, tag := range tags {
			if ContainsCaseInsensitive(repo.Repository.Name, tag) ||
				ContainsCaseInsensitive(repo.Readme, tag) ||
				ContainsCaseInsensitive(repo.Repository.Project.Name, tag) {
				actions = append(actions, BitbucketBrowseAction{repo: repo})
				break
			}
		}
	}
	return actions
}
