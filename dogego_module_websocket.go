package dogego_module_websocket

import (
	"github.com/doge-soft/dogego_module_websocket/impls"
	"github.com/gorilla/websocket"
	"net/http"
)

func Upgrade(w http.ResponseWriter, r *http.Request) (*impls.Connection, error) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return nil, err
	}

	return impls.NewConnection(conn)
}
