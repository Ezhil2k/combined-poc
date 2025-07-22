package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"proto/trade_engine_walletpb" // Import generated proto

	"github.com/IBM/sarama"
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
)

// Kafka config
var (
	brokers       = []string{"localhost:9092"} // Kafka brokers
	topic         = "user_purchase"            // Topic to subscribe
	consumerGroup = "wallet-group"             // Unique consumer group for wallet service
	redisClient   *redis.Client
)

// walletServer implements WalletServiceServer interface
// Now reads from Redis
type walletServer struct {
	trade_engine_walletpb.UnimplementedWalletServiceServer
}

// GetTotalSpent returns the user's total spend from Redis
type GetTotalSpentRequest struct {
	UserId string `json:"user_id"`
}

type GetTotalSpentResponse struct {
	Total float64 `json:"total"`
}

func (s *walletServer) GetTotalSpent(ctx context.Context, req *trade_engine_walletpb.BalanceRequest) (*trade_engine_walletpb.BalanceResponse, error) {
	userID := req.GetUserId()
	val, err := redisClient.Get(ctx, userID).Result()
	if err == redis.Nil {
		return &trade_engine_walletpb.BalanceResponse{Balances: map[string]float64{"total": 0}}, nil
	} else if err != nil {
		return nil, err
	}
	total, _ := strconv.ParseFloat(val, 64)
	return &trade_engine_walletpb.BalanceResponse{Balances: map[string]float64{"total": total}}, nil
}

func (s *walletServer) GetBalance(ctx context.Context, req *trade_engine_walletpb.BalanceRequest) (*trade_engine_walletpb.BalanceResponse, error) {
	userID := req.GetUserId()
	log.Printf("Received GetBalance gRPC request for user_id: %s", userID)
	val, err := redisClient.Get(ctx, userID).Result()
	if err == redis.Nil {
		return &trade_engine_walletpb.BalanceResponse{Balances: map[string]float64{"total": 0}}, nil
	} else if err != nil {
		return nil, err
	}
	total, _ := strconv.ParseFloat(val, 64)
	return &trade_engine_walletpb.BalanceResponse{Balances: map[string]float64{"total": total}}, nil
}

func main() {
	fmt.Println("🚀 WalletService starting...")

	// Connect to Redis
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	// Start gRPC server
	go startGRPCServer()

	// Start Kafka consumer group
	startKafkaConsumerGroup()
}

func startGRPCServer() {
	server := grpc.NewServer()
	trade_engine_walletpb.RegisterWalletServiceServer(server, &walletServer{})

	port := os.Getenv("WALLET_GRPC_PORT")
	if port == "" {
		port = "50051"
	}
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("gRPC server listening on :%s", port)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// --- New Consumer Group Handler for Wallet Service ---
type walletConsumerGroupHandler struct{}

func (walletConsumerGroupHandler) Setup(sess sarama.ConsumerGroupSession) error {
	fmt.Println("🔑 WalletService joined consumer group. Assigned partitions:")
	for topic, partitions := range sess.Claims() {
		fmt.Printf("   Topic: %s, Partitions: %v\n", topic, partitions)
	}
	return nil
}

func (walletConsumerGroupHandler) Cleanup(sess sarama.ConsumerGroupSession) error {
	fmt.Println("🔑 WalletService leaving consumer group.")
	return nil
}

func (walletConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	fmt.Printf("👀 WalletService consuming partition %d\n", claim.Partition())
	ctx := context.Background()
	for msg := range claim.Messages() {
		fmt.Printf("✅ [Kafka] WalletService received message: %s (partition=%d offset=%d)\n", string(msg.Value), msg.Partition, msg.Offset)
		// Parse purchase JSON
		var p struct {
			UserID string  `json:"user_id"`
			Amount float64 `json:"amount"`
		}
		if err := json.Unmarshal(msg.Value, &p); err != nil {
			log.Printf("Invalid purchase JSON: %v", err)
			sess.MarkMessage(msg, "")
			continue
		}
		// Increment user's total in Redis
		newTotal, err := redisClient.IncrByFloat(ctx, p.UserID, p.Amount).Result()
		if err != nil {
			log.Printf("Redis increment error for user %s: %v", p.UserID, err)
		} else {
			log.Printf("Updated total for user %s: %.2f", p.UserID, newTotal)
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}

func startKafkaConsumerGroup() {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Consumer.Return.Errors = true

	client, err := sarama.NewConsumerGroup(brokers, consumerGroup, config)
	if err != nil {
		log.Fatalf("❌ Error creating consumer group client: %v", err)
	}
	defer client.Close()

	handler := walletConsumerGroupHandler{}
	ctx := context.Background()

	for {
		fmt.Println("🔄 WalletService waiting for messages...")
		err := client.Consume(ctx, []string{topic}, handler)
		if err != nil {
			log.Fatalf("❌ Error consuming from topic: %v", err)
		}
	}
}
