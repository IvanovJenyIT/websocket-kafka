package main

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type TestConfig struct {
	clientCount 		int
	wg 							*sync.WaitGroup
	brMsgCount 			*atomic.Int64
	targetMsgCount 	int
}

func DialServer(tc *TestConfig) *websocket.Conn  {
	exit := make(chan struct{})
	dialer := websocket.DefaultDialer
	
	conn,_, err := dialer.Dial(fmt.Sprintf("ws://localhost%s", WSPort), nil)
	if err != nil {
		log.Fatal(err)
	}

	go func(){
		for {
			time.Sleep(2*time.Second)
			if tc.targetMsgCount == int(tc.brMsgCount.Load()) {
				close(exit)
				return
			}
		}
	}()

	go func(){
		<-exit
		conn.Close()
		tc.wg.Done()
	}()
	
	go func(){
		for {
			_, b, err := conn.ReadMessage()
			if err != nil {
				return 
			}
			if len(b) > 0 {
				tc.brMsgCount.Add(1)
			}

			
		}
	}()
	 return conn
}

func TestConnection(t *testing.T) {
	go createWSServer()
	time.Sleep(1* time.Second)
		clientCount := 50
		brCount := 10
	
		tc := TestConfig{
			clientCount: clientCount,
			wg: new(sync.WaitGroup),
			brMsgCount: new(atomic.Int64),
			targetMsgCount: clientCount * brCount,
		}

		tc.wg.Add(tc.clientCount + 1)

		brClient := DialServer(&tc)

		for range tc.clientCount {
			go DialServer(&tc)
		}

		for range brCount{
			msg := ReqMsg{
				MsgType: MsgType_Broadcast,
				Data: "hello from test",
			}
			time.Sleep(100 * time.Millisecond)
			err := brClient.WriteJSON(msg)

			if(err != nil) {
				fmt.Println("error writing message:", err)
				return
			}
		}

		tc.wg.Wait()

		time.Sleep(1* time.Second)
		fmt.Println("exiting test")

}