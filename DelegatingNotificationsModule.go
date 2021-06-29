package main

import (
	"os"
)

type DelegatingNotificationsModule struct {
	activationModule    *ActivationModule
	notificationModules []NotificationModule
	notifications       []Notification
}

func NewDelegatingNotificationsModule(activationModule *ActivationModule, notificationModules []NotificationModule) *DelegatingNotificationsModule {
	module := new(DelegatingNotificationsModule)
	module.activationModule = activationModule
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

func (t *DelegatingNotificationsModule) AddNotification(notification Notification) {
	t.notifications = append(t.notifications, notification)
}

func (t *DelegatingNotificationsModule) notify() {
	for _, notificationModule := range t.notificationModules {
		if t.activationModule.IsNotificationModuleActive(notificationModule) {
			notificationModule.notify(t.notifications)
		}
	}
}
