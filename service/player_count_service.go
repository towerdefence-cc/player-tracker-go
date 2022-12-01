package service

import (
	"context"
	"github.com/towerdefence-cc/grpc-api-specs/gen/go/service/player_tracker"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"player-tracker-go/mongo"
	"time"
)

var (
	playerRepo       = mongo.Database.Collection("players")
	playerCountCache = make(map[string]uint32)
	lastUpdated      = time.Now()
)

func InitPlayerCounter() {
	updatePlayerCounts()

	ticker := time.NewTicker(2 * time.Second)
	go func() {
		for range ticker.C {
			updatePlayerCounts()
		}
	}()
}

func GetServerTypePlayerCount(serverType string) *player_tracker.GetServerTypePlayerCountResponse {
	return &player_tracker.GetServerTypePlayerCountResponse{PlayerCount: playerCountCache[serverType], LastUpdated: timestamppb.New(lastUpdated)}
}

func updatePlayerCounts() {
	playerCountCache["lobby"] = getUnCachedPlayerCount("lobby")
	playerCountCache["velocity"] = getUnCachedPlayerCount("velocity")
	playerCountCache["tower-defence-game"] = getUnCachedPlayerCount("tower-defence-game")

	lastUpdated = time.Now()
}

func getUnCachedPlayerCount(serverType string) uint32 {
	var query bson.M
	if serverType == "velocity" {
		query = bson.M{"proxyId": bson.M{"$exists": true}}
	} else {
		query = bson.M{"serverId": bson.M{"$regex": "^" + serverType + "-"}}
	}
	value, err := playerRepo.CountDocuments(context.Background(), query)
	if err != nil {
		log.Fatalf("Error getting player count: %v", err)
	}

	return uint32(value)
}
