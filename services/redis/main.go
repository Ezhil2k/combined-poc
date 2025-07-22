package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	// 1. String
	err := rdb.Set(ctx, "user:1000", "Alice", 10*time.Second).Err()
	if err != nil {
		panic(err)
	}
	val, err := rdb.Get(ctx, "user:1000").Result()
	fmt.Println("String example:", val)

	// 2. List
	rdb.Del(ctx, "tasks")
	rdb.RPush(ctx, "tasks", "task1", "task2", "task3")
	tasks, _ := rdb.LRange(ctx, "tasks", 0, -1).Result()
	fmt.Println("List example:", tasks)

	// 3. Set
	rdb.Del(ctx, "online_users")
	rdb.SAdd(ctx, "online_users", "alice", "bob", "carol")
	users, _ := rdb.SMembers(ctx, "online_users").Result()
	fmt.Println("Set example:", users)

	// 4. Sorted Set
	rdb.Del(ctx, "highscores")
	rdb.ZAdd(ctx, "highscores",
		redis.Z{Score: 100, Member: "alice"},
		redis.Z{Score: 200, Member: "bob"},
		redis.Z{Score: 150, Member: "carol"},
	)
	highscores, _ := rdb.ZRangeWithScores(ctx, "highscores", 0, -1).Result()
	fmt.Print("Sorted Set example:")
	for _, z := range highscores {
		fmt.Printf(" %s:%.0f", z.Member, z.Score)
	}
	fmt.Println()

	// 5. Hash
	rdb.Del(ctx, "user:1001")
	rdb.HSet(ctx, "user:1001", "name", "Bob", "email", "bob@example.com")
	userHash, _ := rdb.HGetAll(ctx, "user:1001").Result()
	fmt.Println("Hash example:", userHash)

	// 6. Bitmap
	rdb.Del(ctx, "user:activity")
	rdb.SetBit(ctx, "user:activity", 1, 1) // set bit at offset 1
	rdb.SetBit(ctx, "user:activity", 3, 1) // set bit at offset 3
	bit1, _ := rdb.GetBit(ctx, "user:activity", 1).Result()
	bit3, _ := rdb.GetBit(ctx, "user:activity", 3).Result()
	fmt.Println("Bitmap example: bit 1:", bit1, ", bit 3:", bit3)

	// 7. HyperLogLog
	rdb.PFAdd(ctx, "visitors", "alice", "bob", "carol", "alice")
	count, _ := rdb.PFCount(ctx, "visitors").Result()
	fmt.Println("HyperLogLog example: unique visitors:", count)

	// 8. Stream
	rdb.Del(ctx, "mystream")
	rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "mystream",
		Values: map[string]interface{}{"user": "alice", "action": "login"},
	})
	rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "mystream",
		Values: map[string]interface{}{"user": "bob", "action": "logout"},
	})
	stream, _ := rdb.XRange(ctx, "mystream", "-", "+").Result()
	fmt.Println("Stream example:")
	for _, msg := range stream {
		fmt.Println("  ID:", msg.ID, "Values:", msg.Values)
	}
}
