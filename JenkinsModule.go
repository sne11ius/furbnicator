package main

import (
	"github.com/spf13/viper"
	"log"
)

type JenkinsModule struct {
	httpUrl  string
	username string
	token    string
}

func NewJenkinsModule() *JenkinsModule {
	return new(JenkinsModule)
}

func (j JenkinsModule) Name() string {
	return "Jenkins"
}

func (j JenkinsModule) Description() string {
	return "Provides access to jenkins jobs"
}

func (j JenkinsModule) CanBeDisabled() bool {
	return true
}

func (j JenkinsModule) NeedsExternalData() bool {
	return true
}

func (j JenkinsModule) UpdateSettings() {
	configKey := j.Name() + ".http-url"
	if !viper.IsSet(configKey) {
		log.Fatalf("Missing configuration key `%s` (eg. 'https://jenkins.example.com')", configKey)
	}
	j.httpUrl = viper.GetString(configKey)

	configKey = j.Name() + ".username"
	if !viper.IsSet(configKey) {
		log.Fatalf("Missing configuration key `%s` (eg. 'myusername')", configKey)
	}
	j.username = viper.GetString(configKey)

	configKey = j.Name() + ".token"
	if !viper.IsSet(configKey) {
		log.Fatalf("Missing configuration key `%s` (eg. 'mytoken')", configKey)
	}
	j.token = viper.GetString(configKey)
}
