package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/towerdefence-cc/grpc-api-specs/gen/go/service/player_tracker"
	"go.mongodb.org/mongo-driver/bson"
	mongoDb "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"player-tracker-go/mongo"
	"player-tracker-go/mongo/model"
)

var (
	playerCollection = mongo.Database.Collection("players")
)

func UpdatePlayerProxy(ctx context.Context, playerId uuid.UUID, username string, proxyId string) error {
	filter := bson.D{{"_id", playerId}}
	update := bson.D{{"$set", bson.D{{"_id", playerId}, {"username", username}, {"proxyId", proxyId}}}}
	opts := options.Update().SetUpsert(true)

	_, err := playerCollection.UpdateOne(ctx, filter, update, opts)

	return err
}

func UpdatePlayerServer(ctx context.Context, playerId uuid.UUID, username string, serverId string) error {
	filter := bson.D{{"_id", playerId}}
	update := bson.D{{"$set", bson.D{{"_id", playerId}, {"username", username}, {"serverId", serverId}}}}
	opts := options.Update().SetUpsert(true)

	_, err := playerCollection.UpdateOne(ctx, filter, update, opts)

	return err
}

func GetPlayerServer(ctx context.Context, playerId uuid.UUID) (*player_tracker.GetPlayerServerResponse, error) {
	filter := bson.D{{"_id", playerId}}

	var result model.Player
	err := playerCollection.FindOne(ctx, filter).Decode(&result)

	if err != nil {
		if err == mongoDb.ErrNoDocuments {
			return nil, status.New(codes.NotFound, fmt.Sprintf("Player with id %s not found", playerId)).Err()
		}
		return nil, err
	}

	return &player_tracker.GetPlayerServerResponse{
		Server: &player_tracker.OnlineServer{
			ServerId: result.ServerId,
			ProxyId:  result.ProxyId,
		}}, nil
}

func ProxyPlayerDisconnect(ctx context.Context, playerId uuid.UUID) error {
	filter := bson.D{{"_id", playerId}}
	result, err := playerCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New(fmt.Sprintf("Player with id %s is not connected", playerId))
	}
	return nil
}

func GetPlayerServers(ctx context.Context, playerIds []uuid.UUID) (*player_tracker.GetPlayerServersResponse, error) {
	filter := bson.D{{"_id", bson.D{{"$in", playerIds}}}}

	cursor, err := playerCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var mongoResult []model.Player
	err = cursor.All(ctx, &mongoResult)

	slice := make(map[string]*player_tracker.OnlineServer, len(mongoResult))
	for _, player := range mongoResult {
		slice[player.Id.String()] = &player_tracker.OnlineServer{
			ServerId: player.ServerId,
			ProxyId:  player.ProxyId,
		}
	}

	return &player_tracker.GetPlayerServersResponse{
		PlayerServers: slice,
	}, err
}

func GetServerPlayers(ctx context.Context, serverId string) (*player_tracker.GetServerPlayersResponse, error) {
	filter := bson.D{{"serverId", serverId}}

	cursor, err := playerCollection.Find(ctx, filter)
	if err != nil {
		if err == mongoDb.ErrNoDocuments {
			return nil, status.New(codes.NotFound, fmt.Sprintf("Server with id %s not found", serverId)).Err()
		}
		return nil, err
	}

	var mongoResult []model.Player
	err = cursor.All(ctx, &mongoResult)

	slice := make([]*player_tracker.OnlinePlayer, len(mongoResult))
	for i, player := range mongoResult {
		slice[i] = &player_tracker.OnlinePlayer{
			PlayerId: player.Id.String(),
			Username: player.Username,
		}
	}

	return &player_tracker.GetServerPlayersResponse{
		OnlinePlayers: slice,
	}, err
}

func GetServerPlayerCount(ctx context.Context, serverId string) (*player_tracker.GetServerPlayerCountResponse, error) {
	filter := bson.D{{"serverId", serverId}}

	count, err := playerCollection.CountDocuments(ctx, filter)
	if err != nil {
		if err == mongoDb.ErrNoDocuments {
			return nil, status.New(codes.NotFound, fmt.Sprintf("Server with id %s not found", serverId)).Err()
		}
		return nil, err
	}
	return &player_tracker.GetServerPlayerCountResponse{
		PlayerCount: uint32(count),
	}, nil
}
