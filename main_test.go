package main

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type TestConfig struct {
	clientCount int
	wg *sync.WaitGroup
}

func DialServer(wg *sync.WaitGroup) {
	dialer := websocket.DefaultDialer
	
	conn,_, err := dialer.Dial(fmt.Sprintf("ws://localhost%s", WSPort), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("connected to server", conn.LocalAddr().String())
	time.Sleep(2 * time.Second)

	defer func(){
		wg.Done()
		conn.Close()
	}()
	
}

func TestConnection(t *testing.T) {
	go createWSServer()
	time.Sleep(1* time.Second)
	
		tc := TestConfig{
			clientCount: 50,
			wg: new(sync.WaitGroup),
		}

		tc.wg.Add(tc.clientCount)

		for range tc.clientCount {
			go DialServer(tc.wg)
		}

		tc.wg.Wait()
		fmt.Println("exiting test")

}