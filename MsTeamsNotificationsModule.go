package main

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"text/template"
)

type MsTeamsNotificationsModule struct {
	webhookUrl string
}

func NewMsTeamsNotificationsModule() *MsTeamsNotificationsModule {
	return new(MsTeamsNotificationsModule)
}

func (t *MsTeamsNotificationsModule) notify(notifications []Notification) {
	fmt.Printf("Sending MS Teams notifications\n")
	const jsonBodyTemplate = `
		{
			"@type": "MessageCard",
			"@context": "http://schema.org/extensions",
			"summary": "{{.Title}}",
			"sections": [
				{
					"activityTitle": "{{.Title}}",
					"activityImage": "{{.IconUrl}}",
					"Text": "{{.Text}}",
					"markdown": true
				}
			]
		}
	`
	bodyTemplate := template.Must(template.New("postBody").Parse(jsonBodyTemplate))
	for _, notification := range notifications {
		client := &http.Client{}
		var tpl bytes.Buffer
		if err := bodyTemplate.Execute(&tpl, notification); err != nil {
			log.Fatalf("Could not evaluate json template for teams webhook @ %s: %v", t.webhookUrl, err)
		}
		req, err := http.NewRequest("POST", t.webhookUrl, &tpl)
		if err != nil {
			log.Fatalf("Could not create POST request for teams webhook @ %s: %v", t.webhookUrl, err)
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Could not POST notification to teams webhook @ %s: %v", t.webhookUrl, err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Fatalf("Could not close response body: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("Could not POST notification to teams webhook @ %s (HTTP %v)", t.webhookUrl, resp.StatusCode)
		}
	}
}

func (t *MsTeamsNotificationsModule) Name() string {
	return "MsTeamsNotifications"
}

func (t *MsTeamsNotificationsModule) Description() string {
	return "Sends notifications about changes in other modules to via MS Teams"
}

func (t *MsTeamsNotificationsModule) CanBeDisabled() bool {
	return true
}

func (t *MsTeamsNotificationsModule) UpdateSettings() {
	configKey := t.Name() + ".webhook-url"
	if !viper.IsSet(configKey) {
		log.Fatalf("Missing configuration key `%s` (eg. 'https://my-company.webhook.office.com/webhookb2/webhook-id')", configKey)
	}
	t.webhookUrl = viper.GetString(configKey)
}

func (t *MsTeamsNotificationsModule) NeedsExternalData() bool {
	return false
}

func (t *MsTeamsNotificationsModule) UpdateExternalData() {
	// this intentionally empty
}

func (t *MsTeamsNotificationsModule) WriteExternalData(_ *os.File) {
	// this intentionally empty
}

func (t *MsTeamsNotificationsModule) CreateActions(_ []Tag) []action {
	return []action{}
}

func (t *MsTeamsNotificationsModule) ReadExternalData(_ []byte) error {
	return nil
}
