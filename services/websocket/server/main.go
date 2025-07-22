package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type client struct {
	conn *websocket.Conn
	send chan []byte
}

type hub struct {
	clients    map[*client]bool
	register   chan *client
	unregister chan *client
	broadcast  chan []byte
	mu         sync.Mutex
}

func newHub() *hub {
	h := &hub{
		clients:    make(map[*client]bool),
		register:   make(chan *client),
		unregister: make(chan *client),
		broadcast:  make(chan []byte),
	}
	return h
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c] = true
			h.mu.Unlock()
		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
			}
			h.mu.Unlock()
		case msg := <-h.broadcast:
			h.mu.Lock()
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
					close(c.send)
					delete(h.clients, c)
				}
			}
			h.mu.Unlock()
		}
	}
}

// Add purchase struct

type Purchase struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

// Add global DB and Kafka producer
var db *sql.DB
var producer sarama.SyncProducer

func (c *client) readPump(h *hub) {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		log.Printf("Received from client: %s", msg)

		// Parse purchase JSON
		var p Purchase
		if err := json.Unmarshal(msg, &p); err != nil {
			log.Printf("Invalid purchase JSON: %v", err)
			continue
		}

		// Insert into Postgres
		if db != nil {
			_, err := db.Exec("INSERT INTO purchases (user_id, amount, ts) VALUES ($1, $2, NOW())", p.UserID, p.Amount)
			if err != nil {
				log.Printf("Postgres insert error: %v", err)
			} else {
				log.Printf("Inserted purchase into Postgres: %+v", p)
			}
		}

		// Publish to Kafka
		if producer != nil {
			kmsg, _ := json.Marshal(p)
			msg := &sarama.ProducerMessage{
				Topic: "user_purchase",
				Value: sarama.ByteEncoder(kmsg),
			}
			partition, offset, err := producer.SendMessage(msg)
			if err != nil {
				log.Printf("Kafka publish error: %v", err)
			} else {
				log.Printf("Published to Kafka (partition %d, offset %d): %+v", partition, offset, p)
			}
		}

		h.broadcast <- msg // Broadcast received message to all clients
	}
}

func (c *client) writePump() {
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
	c.conn.Close()
}

func serveWs(h *hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	c := &client{conn: conn, send: make(chan []byte, 256)}
	h.register <- c
	go c.writePump()
	go c.readPump(h)
}

func main() {
	// Connect to Postgres
	var err error
	pgConn := "host=localhost port=5432 user=postgres password=postgres dbname=pocdb sslmode=disable"
	db, err = sql.Open("postgres", pgConn)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("Postgres ping error: %v", err)
	}
	log.Println("Connected to Postgres")

	// Ensure purchases table exists
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS purchases (
		id SERIAL PRIMARY KEY,
		user_id TEXT,
		amount DOUBLE PRECISION,
		ts TIMESTAMP
	)`)
	if err != nil {
		log.Fatalf("Failed to ensure purchases table: %v", err)
	}

	// Connect to Kafka
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	producer, err = sarama.NewSyncProducer([]string{"localhost:9092"}, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to Kafka: %v", err)
	}
	log.Println("Connected to Kafka")

	h := newHub()
	go h.run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(h, w, r)
	})

	// Periodically broadcast time to all clients
	go func() {
		for {
			time.Sleep(7 * time.Second)
			msg := []byte(fmt.Sprintf("Server time: %s", time.Now().Format(time.RFC3339)))
			h.broadcast <- msg
		}
	}()

	log.Println("WebSocket server started on :8080/ws")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
