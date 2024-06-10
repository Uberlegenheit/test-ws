package ws

import (
	"errors"
	"sync"
)

var (
	ErrorAlreadyPooled = errors.New("conflict - already in the pool")
)

type Pool interface {
	AddClient(client Client) error
	DeleteClient(client Client)
	DeleteClientByID(id string)
	ClientAlreadyInPool(id string) bool
}

type pool struct {
	clients map[string]Client
	sync.Mutex
}

func NewPool() Pool {
	return &pool{
		clients: make(map[string]Client),
	}
}

func (p *pool) AddClient(client Client) error {
	p.Lock()
	defer p.Unlock()

	_, alreadyExist := p.clients[client.ID()]
	if alreadyExist {
		return ErrorAlreadyPooled
	}

	p.clients[client.ID()] = client

	return nil
}

func (p *pool) DeleteClient(client Client) {
	p.Lock()

	client.Close()

	delete(p.clients, client.ID())

	p.Unlock()
}

func (p *pool) DeleteClientByID(id string) {
	p.Lock()

	delete(p.clients, id)

	p.Unlock()
}

func (p *pool) ClientAlreadyInPool(id string) bool {
	p.Lock()
	defer p.Unlock()

	_, alreadyExist := p.clients[id]

	return alreadyExist
}
