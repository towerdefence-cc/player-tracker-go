package servers

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/towerdefence-cc/grpc-api-specs/gen/go/service/player_tracker"
	"google.golang.org/grpc"
	"log"
	"net"
	"player-tracker-go/service"
)

const (
	port = 9090
)

func InitGrpc() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	// start the grpc server in the background
	go func() {
		server := grpc.NewServer()
		player_tracker.RegisterPlayerTrackerServer(server, &playerTrackerServer{})
		log.Printf("Starting server on port %d", port)
		if err := server.Serve(lis); err != nil {
			panic(err)
		}
	}()
}

type playerTrackerServer struct {
	player_tracker.UnimplementedPlayerTrackerServer
}

func (s *playerTrackerServer) GetPlayerServer(ctx context.Context, request *player_tracker.PlayerRequest) (*player_tracker.GetPlayerServerResponse, error) {
	playerId, err := uuid.Parse(request.PlayerId)
	if err != nil {
		log.Printf("Failed to parse player id (%s): %s", request.PlayerId, err)
		return nil, err
	}
	return service.GetPlayerServer(ctx, playerId)
}

func (s *playerTrackerServer) GetPlayerServers(ctx context.Context, request *player_tracker.PlayersRequest) (*player_tracker.GetPlayerServersResponse, error) {
	playerIds := make([]uuid.UUID, len(request.PlayerIds))
	for i, id := range request.PlayerIds {
		playerIds[i], _ = uuid.Parse(id)
	}
	servers, err := service.GetPlayerServers(ctx, playerIds)
	return servers, err
}

func (s *playerTrackerServer) GetServerPlayers(ctx context.Context, request *player_tracker.ServerIdRequest) (*player_tracker.GetServerPlayersResponse, error) {
	return service.GetServerPlayers(ctx, request.ServerId)
}

func (s *playerTrackerServer) GetServerPlayerCount(ctx context.Context, request *player_tracker.ServerIdRequest) (*player_tracker.GetServerPlayerCountResponse, error) {
	return service.GetServerPlayerCount(ctx, request.ServerId)
}

func (s *playerTrackerServer) GetServerTypePlayerCount(_ context.Context, request *player_tracker.ServerTypeRequest) (*player_tracker.GetServerTypePlayerCountResponse, error) {
	return service.GetServerTypePlayerCount(request.ServerType), nil
}
