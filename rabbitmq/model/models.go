package model

import "github.com/google/uuid"

type ConnectEventDataPackage struct {
	PlayerId uuid.UUID `json:"playerId"`
	Username string    `json:"username"`
}

type DisconnectEventDataPackage struct {
	PlayerId uuid.UUID `json:"playerId"`
}
