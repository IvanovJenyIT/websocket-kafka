package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)
var (
	WSPort = ":3223"
)

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
}

func (s *Server) AcceptLoop() {
	for {
		select {
			case c :=  <-s.joinServerCH:
				s.joinServer(c)
			case c := <-s.leaveServerCH:
				s.leaveServer(c)
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