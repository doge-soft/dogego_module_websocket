package impls

import (
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type Connection struct {
	websocketConnection *websocket.Conn
	inChan              chan []byte
	outChan             chan []byte
	closeChan           chan byte
	closeMutex          sync.Mutex
	isClosed            bool
}

func NewConnection(wsConnection *websocket.Conn) (*Connection, error) {
	connection := &Connection{
		websocketConnection: wsConnection,
		inChan:              make(chan []byte, 5000),
		outChan:             make(chan []byte, 5000),
		closeChan:           make(chan byte, 1),
		closeMutex:          sync.Mutex{},
	}

	go connection.readLoop()
	go connection.writeLoop()

	return connection, nil
}

func (conn *Connection) ReadMessage() ([]byte, error) {
	var data []byte

	select {
	case data = <-conn.inChan:
	case <-conn.closeChan:
		return nil, errors.New("connection is closed.")
	}

	return data, nil
}

func (conn *Connection) Read() (chan []byte, error) {
	channel := make(chan []byte, 5000)

	go func() {
		for {
			data, err := conn.ReadMessage()

			if err != nil {
				log.Println(err)
			}

			channel <- data
		}
	}()

	return channel, nil
}

func (conn *Connection) Write() (chan []byte, error) {
	channel := make(chan []byte, 5000)

	go func() {
		select {
		case data := <-channel:
			err := conn.WriteMessage(data)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	return channel, nil
}

func (conn *Connection) WriteMessage(message []byte) error {
	select {
	case conn.outChan <- message:
	case <-conn.closeChan:
		return errors.New("connection is closed.")
	}

	return nil
}

func (conn *Connection) Close() {
	conn.websocketConnection.Close()

	conn.closeMutex.Lock()
	if !conn.isClosed {
		close(conn.closeChan)
		conn.isClosed = true
	}
	conn.closeMutex.Unlock()
}

func (conn *Connection) readLoop() {
	for {
		if _, data, err := conn.websocketConnection.ReadMessage(); err == nil {
			select {
			case conn.inChan <- data:
			case <-conn.closeChan:
				goto ERR
			}
		} else {
			goto ERR
		}
	}

ERR:
	conn.Close()
}

func (conn *Connection) writeLoop() {
	for {
		select {
		case data := <-conn.outChan:
			if err := conn.websocketConnection.WriteMessage(websocket.TextMessage, data); err != nil {
				goto ERR
			}
		case <-conn.closeChan:
			goto ERR
		}
	}

ERR:
	conn.Close()
}
