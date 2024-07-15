package redis

import (
	"fmt"
	"strings"
	"testing"

	"github.com/samdevo/solana-pools-redis/config"
)

func getMint(id int) Mint {
	return Mint{
		Address:   fmt.Sprintf("address%d", id),
		ProgramID: fmt.Sprintf("program%d", id),
		Symbol:    fmt.Sprintf("symbol%d", id),
		Name:      fmt.Sprintf("name%d", id),
		Decimals:  id,
	}
}

func getTestPool(id int) PoolInfo {
	return PoolInfo{
		MintA:       getMint(1),
		MintB:       getMint(2),
		PoolID:      fmt.Sprintf("pool%d", id),
		Price:       1.0,
		MintAmountA: 2.0,
		MintAmountB: 3.0,
		FeeRate:     0.01,
		Type:        "Standard",
		Day: TimeBlock{
			Volume: 10000.0,
		},
		Month: TimeBlock{},
		Week:  TimeBlock{},
	}
}

func getRedis(t *testing.T) *RedisClient {
	cfg, _ := config.LoadConfig()
	redis := SetupRedis(cfg.RedisAddress)
	if redis == nil {
		t.Errorf("Failed to setup Redis")
	}
	return redis
}

func TestSetupRedis(t *testing.T) {
	redis := getRedis(t)
	defer redis.Close()
}

func TestAddPool(t *testing.T) {
	redis := getRedis(t)
	defer redis.Close()
	pool := getTestPool(1)
	err := redis.AddPoolToRedis(pool)
	if err != nil {
		t.Errorf("Failed to add pool: %v", err)
	}
}

func TestSetPool(t *testing.T) {
	redis := getRedis(t)
	defer redis.Close()
	pool := getTestPool(2)
	err := redis.AddPoolToRedis(pool)
	if err != nil {
		t.Errorf("Failed to add pool: %v", err)
	}
	poolFromRedis, err := redis.GetPool(pool.PoolKey(), pool.PoolID)
	if err != nil {
		t.Errorf("Failed to get pool: %v", err)
	}
	if poolFromRedis.PoolID != pool.PoolID {
		t.Errorf("PoolID mismatch: %v != %v", poolFromRedis.PoolID, pool.PoolID)
	}
}

func TestSetMint(t *testing.T) {
	redis := getRedis(t)
	defer redis.Close()
	mint := getMint(3)
	err := redis.SetMint(mint)
	if err != nil {
		t.Errorf("Failed to add mint: %v", err)
	}
	mintFromRedis, err := redis.GetMint(mint.Address)
	if err != nil {
		t.Errorf("Failed to get mint: %v", err)
	}
	if mintFromRedis.Address != mint.Address {
		t.Errorf("Mint address mismatch: %v != %v", mintFromRedis.Address, mint.Address)
	}
}

func TestSetSwappable(t *testing.T) {
	redis := getRedis(t)
	defer redis.Close()
	mintA := getMint(4)
	mintB := getMint(5)
	err := redis.SetSwappable(mintA.Address, mintB.Address)
	if err != nil {
		t.Errorf("failed to add swappable mints: %v", err)
	}
	swappable, err := redis.GetSwappable(mintA.Address)
	if err != nil {
		t.Errorf("failed to get swappable mints: %v", err)
	}
	if swappable == nil || len(swappable) != 1 {
		t.Errorf("Swappable mismatch: %v != %v", swappable, []string{mintB.Address})
	}
	if mintB.Address != swappable[0] {
		t.Errorf("Swappable mismatch: %v != %v", swappable, true)
	}
}

func TestLoadDB(t *testing.T) {
	redis := getRedis(t)
	defer redis.Close()
	err := redis.LoadRedisDB(100000)
	if err != nil {
		t.Errorf("Failed to load DB: %v", err)
	}
	keys := redis.rdb.Keys(ctx, "*").Val()
	if len(keys) < 100 {
		t.Errorf("DB not loaded: found %v keys", len(keys))
	}
	mint1 := strings.Split(keys[0], ":")[1]
	mintFromDB, err := redis.GetMint(mint1)
	if err != nil {
		t.Errorf("Failed to get mint from DB: %v", err)
	}
	if mintFromDB.Address == "" {
		t.Errorf("Failed to get mint from DB")
	}
}
