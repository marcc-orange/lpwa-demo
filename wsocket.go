package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// NotificationMessage is sent to browser
type NotificationMessage struct {
	Type         string               `json:"type"`
	Device       *DeviceMessage       `json:"device,omitempty"`
	CommandState *CommandStateMessage `json:"command,omitempty"`
}

// DeviceMessage is sent to browser when a device has sent data
type DeviceMessage struct {
	LightOn    bool   `json:"lightOn"`
	Luminosity uint32 `json:"luminosity"`
}

// CommandMessage is sent from browser for each command
type CommandMessage struct {
	Light string `json:"light"`
}

// CommandStateMessage is sent to browser when a command state is updated
type CommandStateMessage struct {
	FCnt  uint16 `json:"fCnt"`
	State string `json:"state"`
}

// CommandHandler defines a func accepting a CommandMessage
type CommandHandler func(command CommandMessage)

// Shared channel of command for all websockets
var wsCommandChan chan CommandMessage

// Track the list of active websockets channels
var wsChannels map[chan NotificationMessage]struct{}
var wsChannelsLock sync.Mutex

func addNotificationChan(nChan chan NotificationMessage) {
	wsChannelsLock.Lock()
	wsChannels[nChan] = struct{}{}
	wsChannelsLock.Unlock()
}

func removeNotificationChan(nChan chan NotificationMessage) {
	wsChannelsLock.Lock()
	delete(wsChannels, nChan)
	wsChannelsLock.Unlock()
}

func notifyAllWebsockets(notification NotificationMessage) {
	wsChannelsLock.Lock()
	for notifChan := range wsChannels {
		notifChan <- notification
	}
	wsChannelsLock.Unlock()
}

func registerCommandHandler(handler CommandHandler) {
	wsCommandChan = make(chan CommandMessage, 100)
	wsChannels = make(map[chan NotificationMessage]struct{})

	// All received commands are sent to the handler
	go func() {
		for command := range wsCommandChan {
			handler(command)
		}
	}()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, //!\\ Cross-Origin check bypass
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer c.Close()

	// Register a notification channel for this websocket connection
	notificationChan := make(chan NotificationMessage, 100)
	addNotificationChan(notificationChan)
	defer removeNotificationChan(notificationChan)
	defer close(notificationChan)

	// Read Notifications from chan, send to browser
	go func() {
		for notification := range notificationChan {
			log.Printf("send: %+v\n", notification)
			if err := c.WriteJSON(notification); err != nil {
				log.Println(err)
			}
		}
	}()

	// Wait for browser's incomming commands, send to chan
	for {
		var command CommandMessage
		if err := c.ReadJSON(&command); err != nil {
			log.Println(err)
			return
		}
		log.Printf("recv: %+v\n", command)
		wsCommandChan <- command
	}
}
