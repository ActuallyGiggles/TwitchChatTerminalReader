package main

import (
	"github.com/gempir/go-twitch-irc/v3"
)

var clientLoggedIn *twitch.Client
var clientAnonymous *twitch.Client

// twitchClientLoggedIn creates a logged in twitch client and connects it.
func twitchClientLoggedIn() {
	if config.Username == "" {
		return
	}

	clientLoggedIn = &twitch.Client{}
	clientLoggedIn = twitch.NewClient(config.Username, config.OAuth)

	err := clientLoggedIn.Connect()
	if err != nil {
		panic(err)
	}
}

// Say sends a message to a specific twitch chatroom.
func Say(channel string, message string) {
	clientLoggedIn.Say(channel, message)
}

// twitchAnonymousClient creates an anonymous twitch client and connects it.
func twitchAnonymousClient() {
	clientAnonymous = &twitch.Client{}
	clientAnonymous = twitch.NewAnonymousClient()

	clientAnonymous.OnPrivateMessage(func(message twitch.PrivateMessage) {
		m := twitchMsg{User: message.User.DisplayName, UserColor: message.User.Color, Message: message.Message}

		if _, ok := message.User.Badges["subscriber"]; ok {
			m.Badge = "subscriber"
		}

		if _, ok := message.User.Badges["broadcaster"]; ok {
			m.Badge = "broadcaster"
		}

		if _, ok := message.User.Badges["moderator"]; ok {
			m.Badge = "moderator"
		}

		if _, ok := message.User.Badges["vip"]; ok {
			m.Badge = "vip"
		}

		incomingMessages <- m
	})

	clientAnonymous.OnWhisperMessage(func(message twitch.WhisperMessage) {
		incomingMessages <- twitchMsg{User: message.User.DisplayName, UserColor: message.User.Color, Message: message.Message, IsWhisper: true}
	})
	clientAnonymous.OnNoticeMessage(func(message twitch.NoticeMessage) {
		incomingMessages <- twitchMsg{IsNotice: true, Message: message.Message}
	})
	clientAnonymous.OnUserNoticeMessage(func(message twitch.UserNoticeMessage) {
		incomingMessages <- twitchMsg{IsNotice: true, Message: message.Message}
	})

	if config.LastMonitoredRoom != "" {
		Join(config.LastMonitoredRoom)
	}

	err := clientAnonymous.Connect()
	if err != nil {
		panic(err)
	}
}

// Join joins a twitch chatroom.
func Join(channel string) {
	clientAnonymous.Join(channel)
}

// Depart departs a twitch chatroom.
func Depart(channel string) {
	clientAnonymous.Depart(channel)
}
