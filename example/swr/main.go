package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dtrugman/cachehit"
	redis_adapter "github.com/dtrugman/cachehit/adapter/redis/go-redis/v9"
	"github.com/dtrugman/cachehit/example/resource"
	"github.com/dtrugman/cachehit/internal"
)

type Item struct {
	Value string
}

func showWelcome(timeToStale, timeToDead time.Duration, cacheSize int) {
	fmt.Println("\n=== SWR Cache :: Interactive Demo ===")
	fmt.Println()
	fmt.Println("┌─────────────────────┐  ┌─────────────────────┐")
	fmt.Println("│   Layer 1: SWR      │  │   Layer 2: Redis    │")
	fmt.Println("│    (In-Memory)      │  │   (Remote Cache)    │")
	fmt.Println("├─────────────────────┤  ├─────────────────────┤")
	fmt.Printf("│ Size: %-13d │  │ Direct Storage      │\n", cacheSize)
	fmt.Printf("│ Stale: %-12v │  │                     │\n", timeToStale)
	fmt.Printf("│ Dead:  %-12v │  │                     │\n", timeToDead)
	fmt.Println("└─────────────────────┘  └─────────────────────┘")
	fmt.Println()
	fmt.Println("Flow: SWR → Redis")
	fmt.Println()
	fmt.Println("Data States:")
	fmt.Printf("  • Fresh (< %v): Instant from memory\n", timeToStale)
	fmt.Printf("  • Stale (%v-%v): From memory + background refresh\n", timeToStale, timeToDead)
	fmt.Printf("  • Dead (> %v): Fetch from Redis\n", timeToDead)
	fmt.Println()
	fmt.Println("Let's see it in action")
}

func showMenu() {
	fmt.Println("")
	fmt.Println("1. Add item to Redis")
	fmt.Println("2. Remove item from Redis")
	fmt.Println("3. List all items from Redis")
	fmt.Println("4. Get from Cache")
	fmt.Println("5. Exit")
	fmt.Print("Choose an option: ")
}

func run() error {
	ctx := context.Background()

	redisInstance, err := resource.RedisRun(ctx)
	if err != nil {
		return fmt.Errorf("init redis: %w", err)
	}
	defer func() {
		if err := redisInstance.Cleanup(); err != nil {
			fmt.Println("Cleanup redis failed: ", err)
		}
	}()

	redisDB, err := resource.RedisConn(ctx, redisInstance.DSN)
	if err != nil {
		return fmt.Errorf("connect to redis: %w", err)
	}

	redisAdapter := redis_adapter.From[string, string](redisDB)

	cacheSize := 128
	timeToStale := 10 * time.Second
	timeToDead := 30 * time.Second

	errorCallback := func(err error) {
		fmt.Println("Error:", err)
	}

	swr, err := cachehit.NewSWR(
		cacheSize, redisAdapter, timeToStale, timeToDead,
		cachehit.SWRWithErrorCallback(errorCallback),
	)
	if err != nil {
		return fmt.Errorf("new swr: %w", err)
	}

	scanner := bufio.NewScanner(os.Stdin)

	showWelcome(timeToStale, timeToDead, cacheSize)

	for {
		showMenu()
		scanner.Scan()
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			fmt.Print("Enter item key: ")
			scanner.Scan()
			key := strings.TrimSpace(scanner.Text())
			if len(key) == 0 {
				fmt.Println("Error: Key cannot be empty")
				continue
			}

			fmt.Print("Enter item value: ")
			scanner.Scan()
			value := strings.TrimSpace(scanner.Text())

			fmt.Print("Enter expiry duration (0 = never expire): ")
			scanner.Scan()
			expirationStr := strings.TrimSpace(scanner.Text())
			expiration, err := time.ParseDuration(expirationStr)
			if err != nil {
				fmt.Printf("Error: Invalid duration: %v\n", err)
				continue
			}

			set := redisDB.Set(ctx, key, value, expiration)
			if err = set.Err(); err != nil {
				fmt.Printf("Error: Failed to add item to Redis: %v\n", err)
				continue
			}

			fmt.Println("Done")

		case "2":
			fmt.Print("Enter item key to remove: ")
			scanner.Scan()
			key := strings.TrimSpace(scanner.Text())
			if len(key) == 0 {
				fmt.Println("Error: Key cannot be empty")
				continue
			}

			del := redisDB.Del(ctx, key)
			if err = del.Err(); err != nil {
				fmt.Printf("Error: Failed to remove item from Redis: %v\n", err)
			}

			fmt.Println("Done")

		case "3":
			keys := redisDB.Keys(ctx, "*")
			if err = keys.Err(); err != nil {
				fmt.Printf("Error: Failed to list keys from Redis: %v\n", err)
				continue
			}

			keyList := keys.Val()
			fmt.Printf("Found %d items in Redis:\n", len(keyList))
			for _, key := range keyList {
				val := redisDB.Get(ctx, key)
				if val.Err() == nil {
					fmt.Printf("- %s = %s\n", key, val.Val())
				}
			}

		case "4":
			fmt.Print("Enter key to get from cache: ")
			scanner.Scan()
			key := strings.TrimSpace(scanner.Text())
			if len(key) == 0 {
				fmt.Println("Error: Key cannot be empty")
				continue
			}

			fmt.Println("Fetching item...")
			start := time.Now()
			value, err := swr.Get(ctx, key)
			elapsed := time.Since(start)

			if errors.Is(err, internal.ErrNotFound) {
				fmt.Printf("Not found!\n")
			} else if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("\nFound: %s = %s\n", key, value)
				fmt.Printf("(fetched in %v)\n", elapsed)
			}

		case "5":
			fmt.Println("Goodbye!")
			return nil

		default:
			fmt.Println("Invalid choice")
		}
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
	}
}
