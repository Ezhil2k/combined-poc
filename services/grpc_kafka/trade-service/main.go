package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"proto/trade_engine_walletpb" // Import generated proto

	"github.com/IBM/sarama"
	"google.golang.org/grpc"
)

// Kafka configuration
var (
	brokers = []string{"localhost:9092"} // Kafka broker(s)
	topic   = "order-events"             // Kafka topic
)

func main() {
	fmt.Println("🚀 TradeService starting...")

	// Connect to WalletService gRPC server
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("❌ Could not connect: %v", err)
	}
	defer conn.Close()

	// Create a client for WalletService
	client := trade_engine_walletpb.NewWalletServiceClient(conn)

	// Call GetBalance for user123
	userID := "user456"
	fmt.Printf("📞 Requesting balances for user: %s\n", userID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	req := &trade_engine_walletpb.BalanceRequest{UserId: userID}
	res, err := client.GetBalance(ctx, req)
	if err != nil {
		log.Fatalf("❌ Error calling GetBalance: %v", err)
	}

	fmt.Println("✅ Received balances:")
	for currency, amount := range res.GetBalances() {
		fmt.Printf("   %s: %.4f\n", currency, amount)
	}

	// --- Kafka Producer Setup ---
	fmt.Println("🛠️ Initializing Kafka producer...")
	producer, err := sarama.NewSyncProducer(brokers, nil)
	if err != nil {
		log.Fatalf("❌ Failed to start Kafka producer: %v", err)
	}
	defer producer.Close()
	fmt.Println("📦 Kafka producer initialized.")

	// --- Send an OrderPlaced message ---
	for i := 1; i <= 10; i++ {
		msg := fmt.Sprintf(`{"order_id":"ORD%05d", "user_id":"user%d", "asset":"BTC", "amount":%.2f}`,
			i, i%100, float32(i)*0.001)
		fmt.Printf("🚚 Sending message %d: %s\n", i, msg)
		kafkaMsg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(msg),
		}
		partition, offset, err := producer.SendMessage(kafkaMsg)
		if err != nil {
			fmt.Printf("❌ Failed to send message %d: %v\n", i, err)
		} else {
			fmt.Printf("✅ Sent message %d to partition=%d offset=%d\n", i, partition, offset)
		}
	}
}
