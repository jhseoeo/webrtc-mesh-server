package main

import (
	"sync"
)

// Set of user data for user channel
type UserInfo struct {
	Session SessionName
	UUID    UUIDType
	Client  Client
}

// Set of message data for message channel
type MessageInfo struct {
	Session SessionName
	Message MessageData
}

// Message data protocol
type MessageData struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	SrcUUID UUIDType    `json:"srcuuid"`
	DstUUID UUIDType    `json:"dstuuid"`
}

type ChannelSet struct {
	register      chan UserInfo
	unregister    chan UserInfo
	deleteSession chan bool
	broadcast     chan MessageInfo
	signaling     chan MessageInfo
}

// Set of channels
type Hub struct {
	mutex    sync.RWMutex
	channels map[SessionName]ChannelSet
}

func CreateHub() *Hub {
	hub := Hub{
		channels: make(map[SessionName]ChannelSet),
	}

	return &hub
}

func (h *Hub) RegisterUser(session SessionName, uuid UUIDType, client Client) error {
	if _, ok := h.channels[session]; !ok {
		channelSet, err := RunSessionLoop()
		if err != nil {
			return err
		}

		h.mutex.Lock()
		h.channels[session] = *channelSet
		h.mutex.Unlock()
	}

	h.mutex.RLock()
	h.channels[session].register <- UserInfo{session, uuid, client}
	h.mutex.RUnlock()

	return nil
}

func (h *Hub) UnregisterUser(session SessionName, uuid UUIDType, client Client) {
	h.mutex.RLock()
	h.channels[session].unregister <- UserInfo{session, uuid, client}
	toDetete := <-h.channels[session].deleteSession
	h.mutex.RUnlock()

	if toDetete {
		h.mutex.Lock()
		delete(h.channels, session)
		h.mutex.Unlock()
	}
}

func (h *Hub) SendBroadcastMessage(session SessionName, uuid UUIDType, message MessageData) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	h.channels[session].broadcast <- MessageInfo{session, message}
}

func (h *Hub) SendSignallingMessage(session SessionName, uuid UUIDType, message MessageData) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	h.channels[session].signaling <- MessageInfo{session, message}
}
