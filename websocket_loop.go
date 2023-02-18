package main

import (
	"fmt"

	ws "github.com/gofiber/websocket/v2"
)

// When a client has connected, client send its uuid.
// Get uuid from this message.
func getUUID(conn *ws.Conn) (MessageData, UUIDType, error) {
	var messageData MessageData
	err := conn.ReadJSON(&messageData)

	if err != nil {
		if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
			fmt.Println("read error:", err)
		}
		return MessageData{}, "", err
	} else if messageData.Type != "join" {
		return MessageData{}, "", fmt.Errorf("unexpected message type. expected 'join' but got %s", messageData.Type)
	}

	return messageData, UUIDType(messageData.SrcUUID), nil

}

// Send users list in the session to the client
func sendUserList(cds *ClientDataStore, conn *ws.Conn, session SessionName) error {
	users := cds.GetSessionData(session)

	uuidList := make([]string, 0, len(users))
	for uuid := range users {
		uuidList = append(uuidList, string(uuid))
	}

	data := map[string]interface{}{
		"type": "users",
		"data": map[string]interface{}{
			"users": uuidList,
		},
		"srcuuid": "",
		"dstuuid": "",
	}

	err := conn.WriteJSON(data)
	if err != nil {
		return err
	} else {
		return nil
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

// Websocket Session Loop for each client
func WebsocketConnectionLoop(cds *ClientDataStore, hub *Hub, conn *ws.Conn) {
	session := SessionName(conn.Params("session"))
	joinMessage, uuid, err := getUUID(conn)
	if err != nil {
		fmt.Println("an error occured getting uuid:", err)
		conn.Close()
	}

	fmt.Printf("user %s joined on %s\n", uuid, session)
	hub.SendBroadcastMessage(session, uuid, joinMessage) // broadcast current users uuid

	client := Client{conn: conn}
	hub.RegisterUser(session, uuid, client) // add current user's information to users list

	err = sendUserList(cds, conn, session) // send users list to user
	if err != nil {
		fmt.Println("an error occured sending user list:", err)
		conn.Close()
	}

	defer func() { // when user leaves
		hub.SendBroadcastMessage(session, uuid, createLeaveMessage(uuid)) // broadcast current users uuid
		hub.UnregisterUser(session, uuid, client)                         // delete current user's information from users list
		conn.Close()
	}()

	for {
		var messageData MessageData
		err := conn.ReadJSON(&messageData)

		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				fmt.Println("read error:", err)
			}
			return
		}

		hub.SendSignallingMessage(session, uuid, messageData)
	}
}
