package main

import (
	"fmt"
	"github.com/sne11ius/bitclient"
	"github.com/spf13/viper"
	"log"
	"os"
)

type BitbucketModule struct {
	gitUrl   string
	httpUrl  string
	username string
	password string
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
	// Der Kollege schneitet den Pfad von unserer URL ab. Er ist wohl kein
	// /bitbucket-Suffix gewohnt. Werden wir vermutlich in unserem eigenen Fork
	// korrigieren müssen.
	client := bitclient.NewBitClient(b.httpUrl, b.username, b.password)

	requestParams := bitclient.PagedRequest{
		Limit: 10000,
		Start: 0,
	}
	projectsResponse, err := client.GetProjects(requestParams)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, project := range projectsResponse.Values {
		fmt.Printf("Project : %d - %s\n", project.Id, project.Key)
	}
}

func (b *BitbucketModule) WriteExternalData(file *os.File) {
	panic("implement me")
}
