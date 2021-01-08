package main

import (
	"github.com/spf13/viper"
	"log"
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

func (b BitbucketModule) Name() string {
	return "Bitbucket"
}

func (b BitbucketModule) Description() string {
	return "Provides access to bitbucket projects"
}

func (b BitbucketModule) CanBeDisabled() bool {
	return true
}

func (b BitbucketModule) UpdateSettings() {
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

func (b BitbucketModule) NeedsExternalData() bool {
	return true
}

func (b BitbucketModule) UpdateExternalData() {
	panic("implement me")
}
