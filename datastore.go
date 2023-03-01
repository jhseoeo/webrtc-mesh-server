package main

import (
	ws "github.com/gofiber/websocket/v2"
)

type SessionName string
type UUIDType string

// Client struct type. you can add any data here
type Client struct {
	Conn *ws.Conn
}

// Datastore of client
type SessionDataStore struct {
	dataStore map[UUIDType]Client
}

// Create new datastore
func MakeSessionDataStore() *SessionDataStore {
	return &SessionDataStore{
		dataStore: make(map[UUIDType]Client),
	}
}

// Get data of users in session from datastore
func (ds *SessionDataStore) GetSessionData(session SessionName) map[UUIDType]Client {
	res := map[UUIDType]Client{}
	for k, v := range ds.dataStore {
		res[k] = v
	}

	return res
}

// Get data of the user from datastore
func (ds *SessionDataStore) GetClientData(session SessionName, uuid UUIDType) Client {
	return ds.dataStore[uuid]
}

func (ds *SessionDataStore) IsEmpty(session SessionName) bool {
	return len(ds.dataStore) == 0
}

// Set user data in datastore
func (ds *SessionDataStore) SetUserData(session SessionName, uuid UUIDType, client Client) {
	ds.dataStore[uuid] = client
}

// Delete data of the user from datastore
func (ds *SessionDataStore) DeleteUserData(session SessionName, uuid UUIDType) {
	delete(ds.dataStore, uuid)
}
