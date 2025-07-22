# go-gorilla

A simple Go project demonstrating WebSocket communication using the [Gorilla WebSocket](https://github.com/gorilla/websocket) library.

## Structure

```
go-gorilla/
├── go.mod
├── server/
│   └── main.go        # WebSocket server
└── client/
    └── main.go        # CLI-based WebSocket client
```

## Features

- WebSocket server that accepts multiple concurrent clients
- Periodically pushes messages (e.g., current time) to all connected clients
- Receives and logs messages from clients
- Handles client disconnects cleanly
- CLI client that connects, prints server messages, and allows user input

## Usage

### 1. Start the Server

```
cd server
# Run the server
go run main.go
```

The server listens on `ws://localhost:8080/ws`.

### 2. Start a Client (in a new terminal)

```
cd client
# Run the client
go run main.go
```

- The client will print messages from the server.
- You can type messages and press Enter to send them to the server.
- Use Ctrl+C to exit gracefully.

You can run multiple clients in separate terminals to see real-time broadcast.
