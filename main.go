package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)
var (
	WSPort = ":3223"
)

type MsgType string

const (
	MsgType_Broadcast MsgType = "broadcast"
)

type ReqMsg struct {
	MsgType MsgType
	Client  *Client
	Data string
}

type RespMsg struct {
	MsgType MsgType
	Data string
	SenderID string
} 

type Client struct {
	ID 		string
	mu 		*sync.RWMutex
	conn 	*websocket.Conn
}

type Server struct {
	client 					map[string]*Client
	mu 							*sync.RWMutex
	joinServerCH 		chan *Client
	leaveServerCH 	chan *Client
	broadcastCH 	chan *ReqMsg
}

func NewRespMsg(msq *ReqMsg) *RespMsg {
	return &RespMsg{
		MsgType: msq.MsgType,
		Data: msq.Data,
		SenderID: msq.Client.ID,
	}
}

func NewClient(conn *websocket.Conn) *Client {
	ID := rand.Text()[:9]
	return &Client{
		ID: ID,
		mu: new(sync.RWMutex),
		conn: conn,
	}
}

func NewServer() *Server {
	return &Server{
		client:  map[string]*Client{},
		mu: new(sync.RWMutex),
		joinServerCH: make(chan *Client, 64),
		leaveServerCH: make(chan *Client, 64),
		broadcastCH: make(chan *ReqMsg, 64),
	}
}

func (c *Client) readMsgLoop(srv *Server) {
	defer func() {
		c.conn.Close()
		srv.leaveServerCH <- c
	}()

	for {
		_, b, err :=c.conn.ReadMessage()
		if err != nil {
			return 
		}

		msg :=new(ReqMsg)
		err =json.Unmarshal(b, msg)
		if err != nil {
			fmt.Printf("error unmarshalling message: %v\n", err)
			continue
		}
		msg.Client = c

		srv.broadcastCH <- msg

	}
}

func (s *Server)handleWS(w http.ResponseWriter,r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize: 512,
		WriteBufferSize: 512,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error http connection:", err)
		return 
	}

	client := NewClient(conn)
	s.joinServerCH <- client

	go client.readMsgLoop(s)
}

func (s *Server) AcceptLoop() {
	for {
		select {
			case c :=  <-s.joinServerCH:
				s.joinServer(c)
			case c := <-s.leaveServerCH:
				s.leaveServer(c)
			case msg := <-s.broadcastCH:
				cls := map[string]*Client{}
				for id, c := range s.client {
					if id != msg.Client.ID {
						cls[id] = c
					}	
				}
				go s.broadCast(msg, cls)
		}
	}
}

func(s *Server) joinServer(c *Client) {
	s.client[c.ID] = c
	fmt.Println("client joined the server", c.ID)
}

func(s *Server) leaveServer(c *Client) {
	delete(s.client, c.ID)
	fmt.Println("client left the server", c.ID)
}

func(s *Server) broadCast(msq *ReqMsg, cls map[string]*Client) {
	resp := NewRespMsg(msq)
	for _, c := range cls {
		err := c.conn.WriteJSON(resp)
		if err != nil {
			fmt.Println("error writing message:", err)
			continue
		}
	}

	fmt.Println("broadcast was sent")
}

func createWSServer() {
	s := NewServer()
	go s.AcceptLoop()
	http.HandleFunc("/", s.handleWS)

	fmt.Println("starting server on", WSPort)
	log.Fatal(http.ListenAndServe(WSPort, nil))
}

func main() {
	createWSServer()
}