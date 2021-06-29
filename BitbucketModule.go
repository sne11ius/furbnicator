package main

import (
	"encoding/json"
	"fmt"
	"github.com/ktrysmt/go-bitbucket"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
)

type BitbucketModule struct {
	username               string
	password               string
	repositoriesWithReadme []BitbucketRepositoryWithReadme
	notificationModule     *DelegatingNotificationsModule
}

type BitbucketRepositoryWithReadme struct {
	Repository bitbucket.Repository `json:"repository"`
	Readme     string               `json:"readme"`
}

func NewBitbucketModule(notificationModule *DelegatingNotificationsModule) *BitbucketModule {
	b := new(BitbucketModule)
	b.notificationModule = notificationModule
	return b
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
	configKey := b.Name() + ".username"
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

func (b *BitbucketModule) LoadRepositories() []bitbucket.Repository {
	client := bitbucket.NewBasicAuth(b.username, b.password)
	workspaces, err := client.Workspaces.List()

	if err != nil {
		log.Fatalf("Cannot list workspaces: %v", err)
	}

	numWorkspaces := len(workspaces.Workspaces)
	var wg sync.WaitGroup
	wg.Add(numWorkspaces)
	queue := make(chan []bitbucket.Repository, 1)
	for i := 0; i < numWorkspaces; i++ {
		go func(i int) {
			workspace := workspaces.Workspaces[i]
			optr := &bitbucket.RepositoriesOptions{
				Owner: workspace.Slug,
				Role:  "",
			}

			repositories, err := client.Repositories.ListForAccount(optr)
			if err != nil {
				log.Fatalf("Cannot get repositories for %s; %v", workspace.UUID, err)
			}
			fmt.Printf("  - %d repositories in team %v\n", len(repositories.Items), workspace.Name)
			queue <- repositories.Items
		}(i)
	}
	var allRepositories []bitbucket.Repository
	go func() {
		for t := range queue {
			allRepositories = append(allRepositories, t...)
			wg.Done()
		}
	}()
	wg.Wait()
	return allRepositories
}

func (b *BitbucketModule) LoadReadmes(repositories []bitbucket.Repository) []BitbucketRepositoryWithReadme {
	numRepositories := len(repositories)
	var wg sync.WaitGroup
	wg.Add(numRepositories)
	queue := make(chan BitbucketRepositoryWithReadme, 1)
	for i := 0; i < numRepositories; i++ {
		go func(i int) {
			repository := repositories[i]
			readme, _ := b.GetReadmeText(repository)
			queue <- BitbucketRepositoryWithReadme{
				Repository: repository,
				Readme:     readme,
			}
		}(i)
	}
	var repositoriesWithReadmes []BitbucketRepositoryWithReadme
	go func() {
		for t := range queue {
			repositoriesWithReadmes = append(repositoriesWithReadmes, t)
			wg.Done()
		}
	}()
	wg.Wait()
	return repositoriesWithReadmes
}

func (b *BitbucketModule) GetReadmeText(repository bitbucket.Repository) (string, error) {
	// Haha, I'm sure its not me that messed up. Thats the beauty of golang :D
	iter := reflect.ValueOf(repository.Links["self"]).MapRange()
	var selfLink string
	for iter.Next() {
		selfLink = iter.Value().Interface().(string)
		break
	}
	var srcLink = selfLink + "/src"
	client := &http.Client{}
	req, err := http.NewRequest("GET", srcLink, nil)
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
	// Ok, I'll admit its not golangs fault from here on ;)
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		} else {
			var jsonMap map[string]interface{}
			err := json.Unmarshal(bodyBytes, &jsonMap)

			if err != nil {
				log.Fatalf("Cannot parse bitbucket api response.")
			}
			iter := reflect.ValueOf(jsonMap).MapRange()
			for iter.Next() {
				key := iter.Key()
				if key.String() == "values" {
					values := iter.Value()
					kind := values.Kind()
					if kind == reflect.Interface {
						items := values.Interface().([]interface{})
						for i := 0; i < len(items); i++ {
							item := items[i]
							file := item.(map[string]interface{})
							path := file["path"].(string)
							if strings.EqualFold(path, "README.md") {
								iter := reflect.ValueOf(file["links"]).MapRange()
								var selfLink interface{}
								for iter.Next() {
									selfLink = iter.Value().Interface().(map[string]interface{})
									readmeLink := selfLink.(map[string]interface{})["href"].(string)
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
							}
						}
					}
				}
			}
		}
	} else {
		return "", fmt.Errorf("could not get README. Status: %v", resp.StatusCode)
	}
	return "", nil
}

func (b *BitbucketModule) UpdateExternalData() {
	allRepositories := b.LoadRepositories()
	fmt.Printf("  - Analyzing %d repositories\n", len(allRepositories))
	updatedReposList := b.LoadReadmes(allRepositories)
	var newRepos []BitbucketRepositoryWithReadme
	for _, repo := range updatedReposList {
		slug := repo.Repository.Slug
		found := false
		for _, previewsRepo := range b.repositoriesWithReadme {
			if slug == previewsRepo.Repository.Slug {
				found = true
				break
			}
			if found {
				break
			}
		}
		if !found {
			newRepos = append(newRepos, repo)
		}
	}
	if len(newRepos) > 0 {
		b.notificationModule.AddNotification(prepareBitbucketNotification(newRepos))
	}
	b.repositoriesWithReadme = updatedReposList
}

func prepareBitbucketNotification(createdRepos []BitbucketRepositoryWithReadme) Notification {
	text := ""
	for _, repo := range createdRepos {
		text = text + "- " + repo.Repository.Name + "\\n"
	}
	return Notification{
		Title:   "New bitbucket repositories found",
		Text:    text,
		IconUrl: "https://raw.githubusercontent.com/sne11ius/furbnicator/main/bitbucket-logo.jpeg",
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

func (b *BitbucketModule) ReadExternalData(data []byte) error {
	err := json.Unmarshal(data, &b.repositoriesWithReadme)
	if err != nil {
		return err //log.Fatalf("Cannot read bitbucket project cache. Consider running with `-u` parameter.")
	}
	return nil
}

type BitbucketBrowseAction struct {
	repo BitbucketRepositoryWithReadme
}

func (b BitbucketBrowseAction) GetLabel() string {
	return "[bitbucket[] BROWSE " + b.repo.Repository.Name
}

func (b BitbucketBrowseAction) Run() string {
	url := b.repo.Repository.Links["html"].(map[string]interface{})["href"].(string)
	if err := launchUrl(url); err != nil {
		log.Fatalf("Could not browse %s: %v", url, err)
	}
	return "Opened " + url
}

type BitbucketCloneAction struct {
	repo BitbucketRepositoryWithReadme
}

func (b BitbucketCloneAction) GetLabel() string {
	return "[bitbucket[] CLONE " + b.repo.Repository.Name
}

func (b BitbucketCloneAction) Run() string {
	var cloneUrl string
	found := false
	cloneLinks := b.repo.Repository.Links["clone"].([]interface{})
	for _, link := range cloneLinks {
		linkMap := link.(map[string]interface{})
		name := linkMap["name"].(string)
		if name == "ssh" {
			cloneUrl = linkMap["href"].(string)
			found = true
			break
		}
	}
	if !found {
		log.Fatalf("No ssh clone link found for repo %s", b.repo.Repository.Name)
	}
	if err := runGitClone(cloneUrl); err != nil {
		log.Fatalf("Could not clone %s: %v", cloneUrl, err)
	}
	return "Cloned " + cloneUrl
}

func (b *BitbucketModule) CreateActions(tags []Tag) []action {
	var actions []action
	for _, repo := range b.repositoriesWithReadme {
		strs := []string{"bitbucket", "browse", repo.Repository.Name, repo.Readme, repo.Repository.Project.Name}
		if DoMatch(strs, tags) {
			actions = append(actions, BitbucketBrowseAction{repo: repo})
		}
		strs = []string{"bitbucket", "clone", repo.Repository.Name, repo.Readme, repo.Repository.Project.Name}
		if DoMatch(strs, tags) {
			actions = append(actions, BitbucketCloneAction{repo: repo})
		}
	}
	return actions
}
