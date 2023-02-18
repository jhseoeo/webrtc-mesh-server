package main

import (
	"fmt"

	ws "github.com/gofiber/websocket/v2"
)

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

func WebsocketConnectionLoop(cds *ClientDataStore, hub *MessageHub, conn *ws.Conn) {
	session := SessionName(conn.Params("session"))
	joinMessage, uuid, err := getUUID(conn)
	if err != nil {
		fmt.Println("an error occured getting uuid:", err)
		conn.Close()
	}

	fmt.Printf("user %s joined on %s\n", uuid, session)
	hub.SendBroadcastMessage(session, uuid, joinMessage)

	client := Client{conn: conn}
	hub.RegisterUser(session, uuid, client)

	err = sendUserList(cds, conn, session)
	if err != nil {
		fmt.Println("an error occured sending user list:", err)
		conn.Close()
	}

	defer func() {
		hub.SendBroadcastMessage(session, uuid, createLeaveMessage(uuid))
		hub.UnregisterUser(session, uuid, client)
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
