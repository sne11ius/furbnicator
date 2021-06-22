package main

import (
	"os"
)

type DelegatingNotificationsModule struct {
	notificationModules []NotificationModule
	notifications       []string
}

func NewDelegatingNotificationsModule(notificationModules []NotificationModule) *DelegatingNotificationsModule {
	module := new(DelegatingNotificationsModule)
	module.notificationModules = notificationModules
	return module
}

func (t *DelegatingNotificationsModule) Name() string {
	return "DelegatingNotificationsModule"
}

func (t *DelegatingNotificationsModule) Description() string {
	return "Sends notificationModule about changes in other modules to a ms teams incoming web hook"
}

func (t *DelegatingNotificationsModule) CanBeDisabled() bool {
	return false
}

func (t *DelegatingNotificationsModule) UpdateSettings() {
	// this intentionally empty
}

func (t *DelegatingNotificationsModule) NeedsExternalData() bool {
	return false
}

func (t *DelegatingNotificationsModule) UpdateExternalData() {
	// this intentionally empty
}

func (t *DelegatingNotificationsModule) WriteExternalData(_ *os.File) {
	// this intentionally empty
}

func (t *DelegatingNotificationsModule) CreateActions(_ []Tag) []action {
	return []action{}
}

func (t *DelegatingNotificationsModule) ReadExternalData(_ []byte) error {
	return nil
}

func (t *DelegatingNotificationsModule) AddNotification(text string) {
	t.notifications = append(t.notifications, text)
}

func (t *DelegatingNotificationsModule) notify() {
	for _, notificationModule := range t.notificationModules {
		notificationModule.notify(t.notifications)
	}
}
