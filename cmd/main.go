package main

import (
	"fmt"
	"log"

	"github.com/samdevo/solana-pools-redis/config"
	"github.com/samdevo/solana-pools-redis/redis"
)

func main() {
	fmt.Println("Loading Config...")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("Setting up Redis...")
	redisClient := redis.SetupRedis(cfg.RedisAddress)
	defer redisClient.Close()
	fmt.Println("Loading Redis DB...")
	err = redisClient.LoadRedisDB()
	if err != nil {
		log.Fatalf("Failed to load Redis DB: %v", err)
	}
}
