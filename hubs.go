package main

import (
	"fmt"

	ws "github.com/gofiber/websocket/v2"
)

type UserInfo struct {
	Session SessionName
	UUID    UUIDType
	Client  Client
}
type MessageInfo struct {
	Session SessionName
	UUID    UUIDType
	Message MessageData
}
type MessageData struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	SrcUUID UUIDType    `json:"srcuuid"`
	DstUUID UUIDType    `json:"dstuuid"`
}
type MessageHub struct {
	register         chan UserInfo
	unregister       chan UserInfo
	sessionBroadcast chan MessageInfo
	newUserSignaling chan MessageInfo
	datastore        *ClientDataStore
}

func sendBroadcastMessage(cds *ClientDataStore, mi MessageInfo) error {
	err := cds.IterateSession(mi.Session, func(uuid UUIDType, client Client) error {
		if uuid != mi.UUID {
			err := client.conn.WriteJSON(mi.Message)
			if err != nil {
				fmt.Println("write error:", err)
				client.conn.WriteMessage(ws.CloseMessage, []byte{})
				client.conn.Close()
				return err
			}
		}

		return nil
	})

	return err
}

func sendSignalingMessage(cds *ClientDataStore, mi MessageInfo) error {
	client := cds.GetClientData(mi.Session, mi.Message.DstUUID)
	err := client.conn.WriteJSON(mi.Message)
	if err != nil {
		fmt.Println("write error:", err)
		client.conn.WriteMessage(ws.CloseMessage, []byte{})
		client.conn.Close()
		return err
	}

	return nil
}

func userHub(hub *MessageHub) {
	for {
		select {
		case registerUser := <-hub.register:
			hub.datastore.SetUserData(registerUser.Session, registerUser.UUID, registerUser.Client)

		case unregisterUser := <-hub.unregister:
			hub.datastore.DeleteUserData(unregisterUser.Session, unregisterUser.UUID)
			fmt.Printf("user %s leaved from %s\n", unregisterUser.UUID, unregisterUser.Session)
		}
	}
}

func messageHub(hub *MessageHub) {
	for {
		select {
		case messageData := <-hub.sessionBroadcast:
			err := sendBroadcastMessage(hub.datastore, messageData)
			if err != nil {
				fmt.Println("an error occured while sending broadcast message, but still process :", err)
			}
		case messageData := <-hub.newUserSignaling:
			err := sendSignalingMessage(hub.datastore, messageData)
			if err != nil {
				fmt.Println("an error occured while sending signaling message, but still process :", err)
			}
		}
	}
}

func CreateHub(clients *ClientDataStore) *MessageHub {
	hub := MessageHub{
		register:         make(chan UserInfo),
		unregister:       make(chan UserInfo),
		sessionBroadcast: make(chan MessageInfo),
		newUserSignaling: make(chan MessageInfo),
		datastore:        clients,
	}

	go userHub(&hub)
	go messageHub(&hub)

	return &hub
}

func (h *MessageHub) RegisterUser(session SessionName, uuid UUIDType, client Client) {
	h.register <- UserInfo{session, uuid, client}
}

func (h *MessageHub) UnregisterUser(session SessionName, uuid UUIDType, client Client) {
	h.unregister <- UserInfo{session, uuid, client}
}

func (h *MessageHub) SendBroadcastMessage(session SessionName, uuid UUIDType, message MessageData) {
	h.sessionBroadcast <- MessageInfo{session, uuid, message}
}

func (h *MessageHub) SendSignallingMessage(session SessionName, uuid UUIDType, message MessageData) {
	h.newUserSignaling <- MessageInfo{session, uuid, message}
}
