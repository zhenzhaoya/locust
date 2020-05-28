package core

import (
	"log"

	"golang.org/x/net/websocket"
)

// var (
// 	ws  *websocket.Conn
// )
type WsClient struct {
	conn    *websocket.Conn
	Message chan string
	Closed  bool
	Data    map[string]interface{}
}

// origin := "http://localhost/"
// url := "ws://localhost:12345/ws"
func NewWsClient(url string, origin string, protocol string) (*WsClient, error) {
	c := &WsClient{}
	c.Message = make(chan string)
	ws, err := websocket.Dial(url, protocol, origin)
	if err != nil {
		log.Print(err)
	} else {
		c.conn = ws
	}
	return c, err
}
func (c *WsClient) Close() {
	if !c.Closed {
		c.Message <- "-1"
		c.Closed = true
		c.conn.Close()
	}
}
func (c *WsClient) close(msg string) {
	if !c.Closed {
		c.Message <- "-1 " + msg
		c.Closed = true
		c.conn.Close()
	}
}
func (c *WsClient) Send(data []byte) {
	if _, err := c.conn.Write(data); err != nil {
		log.Print(err)
	}
}
func (c *WsClient) Receive() []byte {
	var n int
	var err error
	var msg = make([]byte, 512)
	if n, err = c.conn.Read(msg); err != nil {
		log.Print(err)
	}
	return msg[:n]
}
func (c *WsClient) SendMessage(message string) {
	websocket.Message.Send(c.conn, message)
}
func (c *WsClient) ReceiveMessage() string {
	var message string
	websocket.Message.Receive(c.conn, &message)
	return message
}

func (c *WsClient) ReceiveLoop() {
	for {
		var reply string
		err := websocket.Message.Receive(c.conn, &reply)
		if c.Closed || err != nil {
			if !c.Closed {
				c.close(err.Error())
			}
			break
		} else {
			if reply != "" {
				c.Message <- reply
			}
		}
	}
}

func (c *WsClient) SendJson(message *interface{}) {
	websocket.JSON.Send(c.conn, message)
}
func (c *WsClient) ReceiveJson() *interface{} {
	var message *interface{}
	websocket.JSON.Receive(c.conn, &message)
	return message
}
