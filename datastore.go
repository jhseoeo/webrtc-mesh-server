package main

import (
	"sync"

	ws "github.com/gofiber/websocket/v2"
)

type SessionName string
type UUIDType string

// Client struct type. you can add any data here
type Client struct {
	conn *ws.Conn
}

// Datastore of client
type ClientDataStore struct {
	mutex        sync.RWMutex
	sessionMutex map[SessionName]sync.RWMutex
	dataStore    map[SessionName]map[UUIDType]Client
}

// Create new datastore
func MakeClientDataStore() *ClientDataStore {
	return &ClientDataStore{
		sessionMutex: make(map[SessionName]sync.RWMutex),
		dataStore:    make(map[SessionName]map[UUIDType]Client),
	}
}

// Get data of users in session from datastore
func (ds *ClientDataStore) GetSessionData(session SessionName) map[UUIDType]Client {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	res := make(map[UUIDType]Client, len(ds.dataStore[session]))
	for key, val := range ds.dataStore[session] {
		res[key] = val
	}

	return res
}

// Get data of the user from datastore
func (ds *ClientDataStore) GetClientData(session SessionName, uuid UUIDType) Client {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	return ds.dataStore[session][uuid]
}

// Set user data in datastore
func (ds *ClientDataStore) SetUserData(session SessionName, uuid UUIDType, client Client) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	if _, ok := ds.dataStore[session]; !ok {
		ds.dataStore[session] = make(map[UUIDType]Client)
	}

	ds.dataStore[session][uuid] = client
}

// Delete data of the user from datastore
func (ds *ClientDataStore) DeleteUserData(session SessionName, uuid UUIDType) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	delete(ds.dataStore[session], uuid)
	if len(ds.dataStore[session]) == 0 {
		delete(ds.dataStore, session)
	}
}
