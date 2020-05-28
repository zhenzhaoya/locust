package core

import (
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/websocket"
)

type WsServer struct {
	conn *websocket.Conn
}

func (c *WsServer) Close() {
	c.conn.Close()
}
func StartWsServer(addr string) {
	fmt.Println("start server ", addr)
	http.Handle("/echo", websocket.Handler(EchoServer))
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
func EchoServer(conn *websocket.Conn) {
	io.Copy(conn, conn)
	// fmt.Printf("a new ws conn: %s-&gt;%s\n", conn.RemoteAddr().String(), conn.LocalAddr().String())
	// var err error
	// // for {
	// var reply string
	// err = websocket.Message.Receive(conn, &reply)
	// if err != nil {
	// 	fmt.Println("receive err:", err.Error())
	// 	// break
	// }
	// fmt.Println("Received from client: " + reply)
	// if err = websocket.Message.Send(conn, reply); err != nil {
	// 	fmt.Println("send err:", err.Error())
	// 	// break
	// }
	// }
}
