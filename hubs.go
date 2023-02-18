package main

import (
	"fmt"

	ws "github.com/gofiber/websocket/v2"
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
	UUID    UUIDType
	Message MessageData
}

// Message data protocol
type MessageData struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	SrcUUID UUIDType    `json:"srcuuid"`
	DstUUID UUIDType    `json:"dstuuid"`
}

// Set of channels
type Hub struct {
	register   chan UserInfo
	unregister chan UserInfo
	broadcast  chan MessageInfo
	signaling  chan MessageInfo
	datastore  *ClientDataStore
}

// Send broadcast message to every user in session
func sendBroadcastMessage(cds *ClientDataStore, mi MessageInfo) error {
	err := cds.ForEachUser(mi.Session, func(uuid UUIDType, client Client) error {
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

// Send signaling message to specific user in session
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

// Run goroutine that handling users
func runUserHub(clients *ClientDataStore) (chan UserInfo, chan UserInfo) {
	registerChannel := make(chan UserInfo)
	unregisterChannel := make(chan UserInfo)

	go func() {
		for {
			select {
			case registerUser := <-registerChannel:
				clients.SetUserData(registerUser.Session, registerUser.UUID, registerUser.Client)

			case unregisterUser := <-unregisterChannel:
				clients.DeleteUserData(unregisterUser.Session, unregisterUser.UUID)
				fmt.Printf("user %s leaved from %s\n", unregisterUser.UUID, unregisterUser.Session)
			}
		}
	}()

	return registerChannel, unregisterChannel
}

// Run goroutine that handling messages
func runMessageHub(clients *ClientDataStore) (chan MessageInfo, chan MessageInfo) {
	broadcastChannel := make(chan MessageInfo)
	signalingChannel := make(chan MessageInfo)

	go func() {
		for {
			select {
			case messageData := <-broadcastChannel:
				err := sendBroadcastMessage(clients, messageData)
				if err != nil {
					fmt.Println("an error occured while sending broadcast message, but still process :", err)
				}
			case messageData := <-signalingChannel:
				err := sendSignalingMessage(clients, messageData)
				if err != nil {
					fmt.Println("an error occured while sending signaling message, but still process :", err)
				}
			}
		}
	}()

	return broadcastChannel, signalingChannel
}

// Create new hub
func CreateHub(clients *ClientDataStore) *Hub {
	registerChannel, unregisterChannel := runUserHub(clients)
	broadcastChannel, signalingChannel := runMessageHub(clients)

	hub := Hub{
		register:   registerChannel,
		unregister: unregisterChannel,
		broadcast:  broadcastChannel,
		signaling:  signalingChannel,
		datastore:  clients,
	}

	return &hub
}

func (h *Hub) RegisterUser(session SessionName, uuid UUIDType, client Client) {
	h.register <- UserInfo{session, uuid, client}
}

func (h *Hub) UnregisterUser(session SessionName, uuid UUIDType, client Client) {
	h.unregister <- UserInfo{session, uuid, client}
}

func (h *Hub) SendBroadcastMessage(session SessionName, uuid UUIDType, message MessageData) {
	h.broadcast <- MessageInfo{session, uuid, message}
}

func (h *Hub) SendSignallingMessage(session SessionName, uuid UUIDType, message MessageData) {
	h.signaling <- MessageInfo{session, uuid, message}
}
