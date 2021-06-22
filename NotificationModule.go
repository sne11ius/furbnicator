package main

type NotificationModule interface {
	notify(notifications []string)
}
