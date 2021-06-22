package main

import (
	"encoding/json"
	"fmt"
	. "github.com/Medisafe/jenkins-api/jenkins"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"sync"
)

type JenkinsModule struct {
	httpUrl            string
	username           string
	token              string
	jobs               []Job
	notificationModule *DelegatingNotificationsModule
}

func NewJenkinsModule(notifications *DelegatingNotificationsModule) *JenkinsModule {
	j := new(JenkinsModule)
	j.notificationModule = notifications
	return j
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

	var newJobDetails []Job
	for _, job := range allJobDetails {
		found := false
		for _, existingJob := range j.jobs {
			if job.Name == existingJob.Name {
				found = true
				break
			}
		}
		if !found {
			newJobDetails = append(newJobDetails, job)
		}
	}

	if len(newJobDetails) > 0 {
		message := "Es gibt neue Jenkins-Jobs\n"
		for _, job := range newJobDetails {
			message = message + "- " + job.Name + " -> " + job.Url + "\n"
		}
		j.notificationModule.AddNotification(message)
	}

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

func (j *JenkinsModule) ReadExternalData(data []byte) error {
	err := json.Unmarshal(data, &j.jobs)
	if err != nil {
		log.Fatalf("Cannot read jenkins job cache. Consider running with `-u` parameter.")
	}
	return nil
}

type JenkinsBrowseAction struct {
	job Job
}

func (j JenkinsBrowseAction) GetLabel() string {
	return "[jenkins[] BROWSE " + j.job.Name
}

func (j JenkinsBrowseAction) Run() string {
	url := j.job.Url
	if err := launchUrl(url); err != nil {
		log.Fatalf("Could not browse %s: %v", url, err)
	}
	return "Opened " + url
}

type JenkinsRunJobAction struct {
	job      Job
	username string
	token    string
}

func (j JenkinsRunJobAction) GetLabel() string {
	return "[jenkins[] RUN " + j.job.Name
}

func (j JenkinsRunJobAction) Run() string {
	url := j.job.Url + "/build"
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Fatalf("Could not create request to start job %s: %v", j.job.Name, err)
	}
	req.SetBasicAuth(j.username, j.token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Could not start job %s: %v", j.job.Name, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatalf("Could not close response body: %v", err)
		}
	}()
	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Could not start job %s (HTTP %v): %v", j.job.Name, resp.StatusCode, err)
	}
	return "Started job " + j.job.Name
}

func (j *JenkinsModule) CreateActions(tags []Tag) []action {
	var actions []action
	for _, job := range j.jobs {
		strs := []string{"jenkins", "browse", job.Name, job.Description}
		if DoMatch(strs, tags) {
			actions = append(actions, JenkinsBrowseAction{job: job})
		}
		strs = []string{"jenkins", "run", job.Name, job.Description}
		if DoMatch(strs, tags) {
			actions = append(actions, JenkinsRunJobAction{
				job:      job,
				username: j.username,
				token:    j.token,
			})
		}
	}
	return actions
}
