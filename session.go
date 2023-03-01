package main

import (
	"fmt"

	ws "github.com/gofiber/websocket/v2"
)

func handleUserRegister(sds *SessionDataStore, ui UserInfo) error {
	joinMessage := createJoinMessage(ui.UUID)
	err := sendBroadcastMessage(sds, MessageInfo{ui.Session, joinMessage}) // broadcast current users uuid
	if err != nil {
		return err
	}

	userListMessage := createUserListMessage(sds, ui)
	sds.SetUserData(ui.Session, ui.UUID, ui.Client)
	err = sendSignalingMessage(sds, MessageInfo{ui.Session, userListMessage})
	if err != nil {
		return err
	}

	return nil
}

func handleUserUnregister(sds *SessionDataStore, ui UserInfo) error {
	leaveMessage := createLeaveMessage(ui.UUID)
	err := sendBroadcastMessage(sds, MessageInfo{ui.Session, leaveMessage})
	if err != nil {
		return err
	}

	sds.DeleteUserData(ui.Session, ui.UUID)
	return nil
}

// Send broadcast message to every user in session
func sendBroadcastMessage(sds *SessionDataStore, mi MessageInfo) error {
	clients := sds.GetSessionData(mi.Session)

	for uuid, client := range clients {
		if uuid != mi.Message.SrcUUID {
			err := client.Conn.WriteJSON(mi.Message)
			if err != nil {
				fmt.Println("write error:", err)
				client.Conn.WriteMessage(ws.CloseMessage, []byte{})
				client.Conn.Close()
				return err
			}
		}
	}

	return nil
}

// Send signaling message to specific user in session
func sendSignalingMessage(sds *SessionDataStore, mi MessageInfo) error {
	client := sds.GetClientData(mi.Session, mi.Message.DstUUID)
	err := client.Conn.WriteJSON(mi.Message)
	if err != nil {
		fmt.Println("write error:", err)
		client.Conn.WriteMessage(ws.CloseMessage, []byte{})
		client.Conn.Close()
		return err
	}

	return nil
}

func createJoinMessage(uuid UUIDType) MessageData {
	return MessageData{
		Type:    "join",
		Data:    nil,
		SrcUUID: uuid,
		DstUUID: "",
	}
}

// Send users list in the session to the client
func createUserListMessage(sds *SessionDataStore, user UserInfo) MessageData {
	users := sds.GetSessionData(user.Session)

	uuidList := make([]string, 0, len(users))
	for uuid := range users {
		uuidList = append(uuidList, string(uuid))
	}

	return MessageData{
		Type: "users",
		Data: map[string]interface{}{
			"users": uuidList,
		},
		SrcUUID: "",
		DstUUID: user.UUID,
	}
}

func createLeaveMessage(uuid UUIDType) MessageData {
	return MessageData{
		Type:    "leave",
		Data:    nil,
		SrcUUID: uuid,
		DstUUID: "",
	}
}

func RunSessionLoop() (*ChannelSet, error) {
	channelSet := ChannelSet{
		register:      make(chan UserInfo),
		unregister:    make(chan UserInfo),
		deleteSession: make(chan bool),
		broadcast:     make(chan MessageInfo),
		signaling:     make(chan MessageInfo),
	}

	clients := MakeSessionDataStore()

	go func() {
	loop:
		for {
			select {
			case registerUser := <-channelSet.register:
				err := handleUserRegister(clients, registerUser)
				if err != nil {
					fmt.Println("an error occurred while handling user registration, but still process :", err)
				}

			case unregisterUser := <-channelSet.unregister:
				err := handleUserUnregister(clients, unregisterUser)
				if err != nil {
					fmt.Println("an error occurred while handling user unregistration, but still process :", err)
				}

				if clients.IsEmpty(unregisterUser.Session) {
					channelSet.deleteSession <- true
					break loop
				} else {
					channelSet.deleteSession <- false
				}

			case messageData := <-channelSet.broadcast:
				err := sendBroadcastMessage(clients, messageData)
				if err != nil {
					fmt.Println("an error occurred while sending broadcast message, but still process :", err)
				}

			case messageData := <-channelSet.signaling:
				err := sendSignalingMessage(clients, messageData)
				if err != nil {
					fmt.Println("an error occurred while sending signaling message, but still process :", err)
				}
			}
		}
	}()

	return &channelSet, nil
}
