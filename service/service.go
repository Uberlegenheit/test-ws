package service

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"
	"websocket-test/ws"
)

type Service interface {
	AddConnectionToMarketDataStream(ctx context.Context, userID uint64, wsConn *websocket.Conn)
}

type ServiceFacade struct {
	wsPool   ws.Pool
	wsMPPool ws.Readers
}

func NewService() (*ServiceFacade, error) {
	p := ws.NewPool()
	mpp := ws.NewMPPool()

	return &ServiceFacade{
		wsPool:   p,
		wsMPPool: mpp,
	}, nil
}

func ConnectWebSocketStreamMarketPrice() (*websocket.Conn, error) {
	url := fmt.Sprintf("%s/ws/!markPrice@arr", "wss://fstream.binance.com")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (s *ServiceFacade) ConnectMarketDataStream() {
	if exists := s.wsPool.ClientAlreadyInPool(ws.MPWebsocketName); exists {
		return
	}

	conn, err := ConnectWebSocketStreamMarketPrice()
	if err != nil {
		log.Printf("binance.ConnectWebSocketStreamMarketPrice: %v", err)
		return
	}

	client := ws.NewWebSocketClient(conn, ws.MPWebsocketName)

	err = s.wsPool.AddClient(client)
	if err != nil {
		log.Printf("wsMPPool.AddConnection: [%v] ", err)
		return
	}

	defer func() {
		s.wsPool.DeleteClient(client)
	}()

	client.Launch(context.Background())

	refreshTicker := time.NewTicker(time.Minute * 10)
	defer refreshTicker.Stop()

	for {
		select {
		case message, ok := <-client.Listen():
			if !ok {
				client.Close()
				return
			} else {
				s.wsMPPool.SendMessages(message)
			}
		case err := <-client.Error():
			log.Printf("websocket error: %v", err)
		case <-client.Ping():
			err = conn.WriteMessage(websocket.PongMessage, []byte{})
			if err != nil {
				log.Printf("conn.WriteMessage: %v", err)
				return
			}
		case <-refreshTicker.C:
			err = conn.WriteMessage(websocket.PongMessage, []byte{})
			if err != nil {
				log.Printf("conn.WriteMessage: %v", err)
				return
			}
		case <-client.Done():
			return
		}
	}
}

func (s *ServiceFacade) AddConnectionToMarketDataStream(ctx context.Context, userID uint64, wsConn *websocket.Conn) {
	if s.wsMPPool.ClientAlreadyInPool(userID) {
		log.Printf("сlientAlreadyInPool: [%d]", userID)
		return
	}

	err := s.wsMPPool.AddConnection(userID, wsConn)
	if err != nil {
		log.Printf("wsMPPool.AddConnection: [%v] ", err)
		return
	}

	defer func() {
		s.wsMPPool.DeleteConnection(userID)
		wsConn.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			log.Printf("websocket closed by request")
			return
		default:

		}
	}
}
