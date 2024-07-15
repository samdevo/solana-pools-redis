# Raydium Pool Tracker

This keeps track of live updates to Raydium pools using the Raydium API for initial load and using a RPC connection to Solana for live updates.

## Usage

1. Create a config/config.json with the following structure:
```json
{
    "redis_address": "YOUR_HOST:YOUR_PORT",
    "rpc_address": "YOUR_SOLANA_RPC_ADDRESS",
}
```
2. Start a local redis server with RedisJSON enabled.
```bash
 docker run -p 6379:6379 --name redis-redisjson redislabs/rejson:latest
```
3. Go to the root directory and run:
```bash
go run cmd/main.go
```


