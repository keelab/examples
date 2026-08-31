// Command client calls the WebSocket echo route from the parent example.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/coder/websocket"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	connection, _, err := websocket.Dial(ctx, "ws://127.0.0.1:8093/ws", nil)
	if err != nil {
		panic(fmt.Errorf("dial WebSocket server: %w", err))
	}
	defer func() {
		if err := connection.Close(websocket.StatusNormalClosure, ""); err != nil {
			log.Printf("close WebSocket connection: %v", err)
		}
	}()

	payload := []byte("hello-keelith")
	if err := connection.Write(ctx, websocket.MessageText, payload); err != nil {
		panic(fmt.Errorf("write WebSocket message: %w", err))
	}
	messageType, response, err := connection.Read(ctx)
	if err != nil {
		panic(fmt.Errorf("read WebSocket message: %w", err))
	}
	if messageType != websocket.MessageText {
		panic(fmt.Errorf("unexpected WebSocket message type %v", messageType))
	}
	fmt.Printf("websocket response: %s\n", response)
}
