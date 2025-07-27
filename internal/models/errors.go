package models

import "errors"

var (
	// ErrRoomFull is returned when trying to join a full room
	ErrRoomFull = errors.New("room is full")
	// ErrBroadcasterExists is returned when trying to set a broadcaster in a room that already has one
	ErrBroadcasterExists = errors.New("broadcaster already exists in this room")
	// ErrInvalidConnectionType is returned when an invalid connection type is provided
	ErrInvalidConnectionType = errors.New("invalid connection type")
)
