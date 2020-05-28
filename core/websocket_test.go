package core

import (
	"fmt"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// go StartWsServer(":12345")
	time.Sleep(time.Second)
	c, _ := NewWsClient("ws://127.0.0.1:12345/echo", "http://127.0.0.1:12345/", "test")
	// defer c.Close()
	go c.ReceiveLoop()
	go func() {
		for {
			time.Sleep(10 * time.Second)
			if c.Closed {
				break
			}
			c.SendMessage("2")
		}
	}()
	for {
		message := <-c.Message
		fmt.Println(message)
		if c.Closed {
			break
		}
	}
	fmt.Println("test main end")
}
