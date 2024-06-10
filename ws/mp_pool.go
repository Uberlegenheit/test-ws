package ws

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

const MPWebsocketName = "mp_ws_connection"

type Readers interface {
	AddConnection(id uint64, conn *websocket.Conn) error
	DeleteConnection(id uint64)
	ClientAlreadyInPool(id uint64) bool
	SendMessages(message []byte)
}

type mpPool struct {
	connections map[uint64]*websocket.Conn
	sync.Mutex
}

func NewMPPool() Readers {
	return &mpPool{
		connections: make(map[uint64]*websocket.Conn),
	}
}

func (p *mpPool) AddConnection(id uint64, conn *websocket.Conn) error {
	p.Lock()
	defer p.Unlock()

	_, alreadyExist := p.connections[id]
	if alreadyExist {
		return ErrorAlreadyPooled
	}

	p.connections[id] = conn

	return nil
}

func (p *mpPool) DeleteConnection(id uint64) {
	p.Lock()
	defer p.Unlock()

	delete(p.connections, id)

}

func (p *mpPool) ClientAlreadyInPool(id uint64) bool {
	p.Lock()
	defer p.Unlock()

	_, alreadyExist := p.connections[id]

	return alreadyExist
}

func (p *mpPool) SendMessages(message []byte) {
	p.Lock()
	defer p.Unlock()

	for _, conn := range p.connections {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("conn.WriteMessage: %v", err)
		}
	}
}
