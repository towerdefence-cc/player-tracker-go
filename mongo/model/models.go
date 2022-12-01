package model

import "github.com/google/uuid"

type Player struct {
	Id       uuid.UUID `bson:"_id"`
	Username string    `bson:"username"`
	ProxyId  string    `bson:"proxyId"`
	ServerId string    `bson:"serverId"`
}
