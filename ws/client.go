package ws

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	maxMessageSize = 1024
	writeWait      = 5 * time.Second
	pingPeriod     = 5 * time.Minute
	pongWait       = 60 * time.Second
)

// Client responsible for simplifying work with WebSocket
type Client interface {
	ID() string
	Launch(ctx context.Context)
	Close() error
	Listen() <-chan []byte
	Done() <-chan interface{}
	Ping() <-chan interface{}
	Error() <-chan error
}

type client struct {
	id           string
	ws           *websocket.Conn
	messagesChan chan []byte
	errorsChan   chan error
	doneChan     chan interface{}
	pingChan     chan interface{}
	sync.Mutex
	sync.Once
}

func NewWebSocketClient(ws *websocket.Conn, apiKey string) Client {
	return &client{
		id:           apiKey,
		ws:           ws,
		messagesChan: make(chan []byte),
		errorsChan:   make(chan error),
		doneChan:     make(chan interface{}),
		pingChan:     make(chan interface{}),
	}
}

// ID returns a unique identifier of the WebSocket connection
func (c *client) ID() string {
	return c.id
}

// Launch launches the client, so it starts listening to new messages
func (c *client) Launch(ctx context.Context) {
	c.ws.SetReadLimit(maxMessageSize * 100)

	c.Do(func() { go c.launch(ctx) })
}

func (c *client) launch(ctx context.Context) {
	var wg sync.WaitGroup

	cancellationCtx, cancel := context.WithCancel(ctx)

	defer func() {
		cancel()
		c.write(websocket.CloseMessage)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		c.read(cancellationCtx)

		cancel()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		c.ping(cancellationCtx)

		cancel()
	}()

	wg.Wait()

	c.doneChan <- struct{}{}
}

// ping is responsible for sending periodical ping. The goroutine is finished when the context is done
func (c *client) ping(ctx context.Context) {
	pingTicker := time.NewTicker(pingPeriod)

	for {
		select {
		case <-ctx.Done():
			return

		case <-pingTicker.C:
			// c.write(websocket.PingMessage)
			c.pingChan <- struct{}{}
		}
	}
}

// Close closes WebSocket connection
func (c *client) Close() error {
	close(c.messagesChan)

	return c.ws.Close()
}

// Listen returns a channel with incoming messages
func (c *client) Listen() <-chan []byte {
	return c.messagesChan
}

// Done returns a channel that closes when work is done (WebSocket connection closed or should be closed)
func (c *client) Done() <-chan interface{} {
	return c.doneChan
}

func (c *client) Ping() <-chan interface{} {
	return c.pingChan
}

// Error returns a channel with errors that happened during WebSocket listening
func (c *client) Error() <-chan error {
	return c.errorsChan
}

// write private method sends service messages (ping, close connection)
func (c *client) write(messageType int) {
	c.Lock()
	defer c.Unlock()

	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	if err := c.ws.WriteMessage(messageType, nil); err != nil {
		c.handleError(err)
	}
}

func (c *client) handleError(err error) {
	var closeError *websocket.CloseError
	if errors.As(err, &closeError) {
		return
	}

	if errors.Is(err, websocket.ErrCloseSent) {
		return
	}

	c.errorsChan <- err
}

// read is responsible for listening to incoming messages. It publishes them to the channel (the channel is returned by Listen method). The goroutine is finished when the context is done or when the read operation returns an error
func (c *client) read(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := c.ws.ReadMessage()
			if err != nil {
				c.handleError(err)
				return
			}

			c.messagesChan <- message
		}
	}
}
