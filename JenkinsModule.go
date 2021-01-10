package main

import (
	"encoding/json"
	"fmt"
	. "github.com/Medisafe/jenkins-api/jenkins"
	"github.com/spf13/viper"
	"log"
	"os"
	"sync"
)

type JenkinsModule struct {
	httpUrl  string
	username string
	token    string
	jobs     []Job
}

func NewJenkinsModule() *JenkinsModule {
	return new(JenkinsModule)
}

func (j *JenkinsModule) Name() string {
	return "Jenkins"
}

func (j *JenkinsModule) Description() string {
	return "Provides access to jenkins jobs"
}

func (j *JenkinsModule) CanBeDisabled() bool {
	return true
}

func (j *JenkinsModule) UpdateSettings() {
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

func (j *JenkinsModule) NeedsExternalData() bool {
	return true
}

func (j *JenkinsModule) UpdateExternalData() {
	jenkinsApi := Init(&Connection{
		Username:    j.username,
		AccessToken: j.token,
		BaseUrl:     j.httpUrl,
	})
	jobs, err := jenkinsApi.GetJobs()
	if err != nil {
		log.Fatalf("Cannot list Jenkins Jobs: %s", err)
	}
	numJobs := len(jobs)

	// From https://stackoverflow.com/a/39065381
	var wg sync.WaitGroup
	wg.Add(numJobs)
	queue := make(chan Job, 1)
	for i := 0; i < numJobs; i++ {
		go func(i int) {
			job := jobs[i]
			details, err := jenkinsApi.GetJob(job.Name)
			if err != nil {
				log.Fatalf("Cannot get job detail for %s", job.Name)
			}
			queue <- *details
		}(i)
	}
	var allJobDetails []Job
	go func() {
		for t := range queue {
			allJobDetails = append(allJobDetails, t)
			wg.Done()
		}
	}()
	wg.Wait()
	j.jobs = allJobDetails
	fmt.Printf("  - Updated %d jobs\n", len(j.jobs))
}

func (j *JenkinsModule) WriteExternalData(file *os.File) {
	bytes, err := json.Marshal(j.jobs)
	if err != nil {
		log.Fatalf("Cannot serialize job data: %s", err)
	}
	if _, err = file.Write(bytes); err != nil {
		log.Fatalf("Cannot write job data to %v: %s", file, err)
	}
}
