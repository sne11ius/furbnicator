package main

type NotificationModule interface {
	notify(notifications []Notification)
	Name() string
}
