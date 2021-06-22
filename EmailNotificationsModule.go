package main

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"net/smtp"
	"os"
	"strings"
)

type EmailNotificationsModule struct {
	smtpServer       string
	smptPort         int
	username         string
	password         string
	recipientAddress string
}

func NewEmailNotificationsModule() *EmailNotificationsModule {
	return new(EmailNotificationsModule)
}

func (e *EmailNotificationsModule) notify(notifications []string) {
	for _, notification := range notifications {
		auth := smtp.PlainAuth("", e.username, e.password, e.smtpServer)
		contentType := "Content-Type: text/html; charset=UTF-8"

		s := fmt.Sprintf("To:%s\r\nFrom:%s\r\nSubject:%s\r\n%s\r\n\r\n%s",
			e.recipientAddress, e.username, "[ﴪ] Notification", contentType, strings.Replace(notification, "\n", "<br>", -1))
		msg := []byte(s)
		addr := fmt.Sprintf("%s:%d", e.smtpServer, e.smptPort)
		err := smtp.SendMail(addr, auth, e.username, []string{e.recipientAddress}, msg)
		if err != nil {
			log.Fatalf("Could not send notification mail: %v", err)
		}
	}
}

func (e *EmailNotificationsModule) Name() string {
	return "EmailNotifications"
}

func (e *EmailNotificationsModule) Description() string {
	return "Sends notificationModule about changes in other modules to via email"
}

func (e *EmailNotificationsModule) CanBeDisabled() bool {
	return true
}

func (e *EmailNotificationsModule) UpdateSettings() {
	configKey := e.Name() + ".smtp-server"
	if !viper.IsSet(configKey) {
		log.Fatalf("Missing configuration key `%s` (eg. 'smtp.example.com')", configKey)
	}
	e.smtpServer = viper.GetString(configKey)
	configKey = e.Name() + ".smtp-port"
	if !viper.IsSet(configKey) {
		log.Fatalf("Missing configuration key `%s` (eg. '587')", configKey)
	}
	e.smptPort = viper.GetInt(configKey)
	configKey = e.Name() + ".username"
	if !viper.IsSet(configKey) {
		log.Fatalf("Missing configuration key `%s` (eg. 'my.user@example.com')", configKey)
	}
	e.username = viper.GetString(configKey)
	configKey = e.Name() + ".password"
	if !viper.IsSet(configKey) {
		log.Fatalf("Missing configuration key `%s` (eg. 'my_password')", configKey)
	}
	e.password = viper.GetString(configKey)
	configKey = e.Name() + ".recipient-address"
	if !viper.IsSet(configKey) {
		log.Fatalf("Missing configuration key `%s` (eg. 'recipient@example.com')", configKey)
	}
	e.recipientAddress = viper.GetString(configKey)
}

func (e *EmailNotificationsModule) NeedsExternalData() bool {
	return false
}

func (e *EmailNotificationsModule) UpdateExternalData() {
	// this intentionally empty
}

func (e *EmailNotificationsModule) WriteExternalData(_ *os.File) {
	// this intentionally empty
}

func (e *EmailNotificationsModule) CreateActions(_ []Tag) []action {
	return []action{}
}

func (e *EmailNotificationsModule) ReadExternalData(_ []byte) error {
	return nil
}
