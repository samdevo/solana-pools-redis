package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	rejson "github.com/nitishm/go-rejson/v4"
	"github.com/redis/go-redis/v9"
)

const (
	POOLDB_KEY      = "pools"
	MINTDB_KEY      = "mints"
	MIN_POOL_VOLUME = 1000
)

var ctx = context.Background()

type RedisClient struct {
	rdb *redis.Client
	rj  *rejson.Handler
}

func SetupRedis(address string) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr: address,
	})
	rj := rejson.NewReJSONHandler()
	rj.SetGoRedisClientWithContext(ctx, rdb)

	// Clear the current database
	err := rdb.FlushDB(ctx).Err()
	if err != nil {
		fmt.Printf("Failed to clear Redis DB: %v\n", err)
		return nil
	}
	fmt.Println("Redis DB cleared")

	// res, _ := rdb.Ping(ctx).Result()
	// fmt.Println("Redis Connection:", res)

	_, err = rj.JSONGet(POOLDB_KEY, ".")
	if err != nil && err.Error() == "redis: nil" {
		_, err = rj.JSONSet(POOLDB_KEY, ".", map[string]interface{}{})
		if err != nil {
			fmt.Printf("Failed to create pools key: %v\n", err)
			return nil
		}
	}

	return &RedisClient{rdb: rdb, rj: rj}
}

type Mint struct {
	Address   string `json:"address"`
	ProgramID string `json:"programId"`
	Symbol    string `json:"symbol"`
	Name      string `json:"name"`
	Decimals  int    `json:"decimals"`
}

type TimeBlock struct {
	Volume    float64 `json:"volume"`
	VolumeFee float64 `json:"volumeFee"`
}

type PoolInfo struct {
	PoolID      string    `json:"id"`
	MintA       Mint      `json:"mintA"`
	MintB       Mint      `json:"mintB"`
	Price       float64   `json:"price"`
	MintAmountA float64   `json:"mintAmountA"`
	MintAmountB float64   `json:"mintAmountB"`
	FeeRate     float64   `json:"feeRate"`
	Type        string    `json:"type"`
	Day         TimeBlock `json:"day"`
	Week        TimeBlock `json:"week"`
	Month       TimeBlock `json:"allTime"`
}

type ApiResponse struct {
	Success bool         `json:"success"`
	Data    ResponseData `json:"data"`
	ID      string       `json:"id"`
}

type ResponseData struct {
	Data        []PoolInfo `json:"data"`
	HasNextPage bool       `json:"hasNextPage"`
	Count       int        `json:"count"`
}

func (c *RedisClient) LoadRedisDB() error {
	// fetch all pages
	baseUrl := "https://api-v3.raydium.io/pools/info/list?poolType=standard&poolSortField=volume24h&sortType=desc&pageSize=1000&page="
	for page := 1; ; page++ {
		url := fmt.Sprintf("%s%d", baseUrl, page)
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to fetch pools: %v", err)
		}
		defer resp.Body.Close()

		var apiResp ApiResponse
		err = json.NewDecoder(resp.Body).Decode(&apiResp)
		if err != nil {
			return fmt.Errorf("failed to decode response: %v", err)
		}
		fmt.Printf("Fetched %d pools\n", apiResp.Data.Count)

		// add pools to redis
		for _, pool := range apiResp.Data.Data {
			if pool.Day.Volume < MIN_POOL_VOLUME {
				return nil
			}
			err = c.AddPoolToRedis(pool)
			if err != nil {
				return fmt.Errorf("failed to add pool to redis: %v", err)
			}
		}
		if !apiResp.Data.HasNextPage {
			break
		}
	}
	return nil
}

func (c *RedisClient) addMintInfo(mint Mint) error {
	exists, err := c.rj.JSONGet(MINTDB_KEY, "$."+mint.Address)
	if err != nil && err != redis.Nil {
		return err
	}
	if exists == redis.Nil {
		_, err = c.rj.JSONSet(MINTDB_KEY, "$."+mint.Address, mint)
		if err != nil && err != redis.Nil {
			return err
		}
	}
	return nil
}

func (c *RedisClient) addSwappableMints(mintA, mintB string) error {
	_, err := c.rdb.SAdd(ctx, "swappable:"+mintA, mintB).Result()
	if err != nil {
		return err
	}
	_, err = c.rdb.SAdd(ctx, "swappable:"+mintB, mintA).Result()
	if err != nil {
		return err
	}
	return nil
}

func (c *RedisClient) AddPoolToRedis(pool PoolInfo) error {
	err := c.addMintInfo(pool.MintA)
	if err != nil {
		return err
	}
	err = c.addMintInfo(pool.MintB)
	if err != nil {
		return err
	}

	err = c.addSwappableMints(pool.MintA.Address, pool.MintB.Address)
	if err != nil {
		return err
	}

	err = c.SetPool(pool)
	if err != nil {
		return err
	}
	return nil
}

func (p *PoolInfo) PoolKey() string {
	if p.MintA.Address < p.MintB.Address {
		return p.MintA.Address + ":" + p.MintB.Address
	}
	return p.MintB.Address + ":" + p.MintA.Address
}

func (c *RedisClient) SetPool(pool PoolInfo) error {

	// Ensure the parent key for the specific pool key exists
	poolKey := pool.PoolKey()
	_, err := c.rj.JSONGet(POOLDB_KEY, "."+poolKey)
	if err != nil && err == redis.Nil {
		_, err = c.rj.JSONSet(POOLDB_KEY, "."+poolKey, map[string]interface{}{})
		if err != nil {
			return err
		}
	}

	path := "." + pool.PoolKey() + "." + pool.PoolID
	_, err = c.rj.JSONSet(POOLDB_KEY, path, pool)
	if err != nil {
		return err
	}
	return nil
}

func (c *RedisClient) GetPool(poolKey, poolID string) (PoolInfo, error) {
	var pool PoolInfo
	_, err := c.rj.JSONGet("pools", "."+poolKey+"."+poolID)
	if err != nil {
		return pool, err
	}
	return pool, nil
}

func (c *RedisClient) Close() error {
	return c.rdb.Close()
}
