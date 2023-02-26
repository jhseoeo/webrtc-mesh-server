package main

import (
	"fmt"

	ws "github.com/gofiber/websocket/v2"
)

// When a client has connected, client send its uuid.
// Get uuid from this message.
func getUUID(conn *ws.Conn) (UUIDType, error) {
	var messageData MessageData
	err := conn.ReadJSON(&messageData)

	if err != nil {
		if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
			fmt.Println("read error:", err)
		}

		return "", err
	} else if messageData.Type != "join" {
		return "", fmt.Errorf("unexpected message type. expected 'join' but got %s", messageData.Type)
	}

	return messageData.SrcUUID, nil
}

// Websocket Session Loop for each client
func WebsocketConnectionLoop(hub *Hub, conn *ws.Conn) {
	session := SessionName(conn.Params("session"))

	uuid, err := getUUID(conn)
	if err != nil {
		fmt.Println("an error occurred getting uuid:", err)
		conn.Close()
	}

	fmt.Printf("user %s joined on %s\n", uuid, session)

	client := Client{Conn: conn}
	hub.RegisterUser(session, uuid, client) // add current user's information to users list

	defer func() { // when user leaves
		fmt.Printf("user %s leaved from %s\n", uuid, session)
		hub.UnregisterUser(session, uuid, client) // delete current user's information from users list
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
