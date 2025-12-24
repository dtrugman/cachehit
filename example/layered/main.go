package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dtrugman/cachehit"
	redis_adapter "github.com/dtrugman/cachehit/adapter/redis/go-redis/v9"
	"github.com/dtrugman/cachehit/example/resource"
)

func showWelcome(timeToStale, timeToDead, redisExpiration time.Duration, cacheSize int) {
	fmt.Println("\n=== Three-Layer Cache :: Interactive Demo ===")
	fmt.Println()
	fmt.Println("┌─────────────────────┐  ┌─────────────────────┐  ┌─────────────────────┐")
	fmt.Println("│   Layer 1: SWR      │  │ Layer 2: LookThrough│  │   Layer 3: HTTP     │")
	fmt.Println("│    (In-Memory)      │  │      (Redis)        │  │    (GitHub API)     │")
	fmt.Println("├─────────────────────┤  ├─────────────────────┤  ├─────────────────────┤")
	fmt.Printf("│ Size: %-13d │  │ Expiration:         │  │ Source of Truth     │\n", cacheSize)
	fmt.Printf("│ Stale: %-12v │  │ %-19v │  │                     │\n", timeToStale, redisExpiration)
	fmt.Printf("│ Dead:  %-12v │  │                     │  │                     │\n", timeToDead)
	fmt.Println("└─────────────────────┘  └─────────────────────┘  └─────────────────────┘")
	fmt.Println()
	fmt.Println("Flow: SWR → LookThrough → HTTP")
	fmt.Println()
	fmt.Println("Data States:")
	fmt.Printf("  • Fresh (< %v): Instant from memory\n", timeToStale)
	fmt.Printf("  • Stale (%v-%v): From memory + background refresh\n", timeToStale, timeToDead)
	fmt.Printf("  • Dead (> %v): Fetch from Redis or GitHub API\n", timeToDead)
	fmt.Println()
	fmt.Println("Lets see it in action")
}

func showMenu() {
	fmt.Println("")
	fmt.Println("1. Get user")
	fmt.Println("2. List all cached users in Redis")
	fmt.Println("3. Remove user from Redis")
	fmt.Println("4. Exit")
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

	redisExpiration := 1 * time.Minute
	redisCache := redis_adapter.From[string, resource.GithubUser](
		redisDB, redis_adapter.WithExpiration(redisExpiration))

	httpRepo := resource.NewGithubUserRepository()

	lookthrough := cachehit.NewLookThrough(redisCache, httpRepo)

	cacheSize := 128
	timeToStale := 10 * time.Second
	timeToDead := 30 * time.Second
	swr, err := cachehit.NewSWR(cacheSize, lookthrough, timeToStale, timeToDead)
	if err != nil {
		return fmt.Errorf("new swr: %w", err)
	}

	scanner := bufio.NewScanner(os.Stdin)

	showWelcome(timeToStale, timeToDead, redisExpiration, cacheSize)

	for {
		showMenu()
		scanner.Scan()
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			fmt.Print("Enter GitHub username: ")
			scanner.Scan()
			username := strings.TrimSpace(scanner.Text())
			if len(username) == 0 {
				fmt.Println("Error: Username cannot be empty")
				continue
			}

			fmt.Println("Fetching user...")
			start := time.Now()
			user, found := swr.Get(ctx, username)
			elapsed := time.Since(start)

			if found {
				fmt.Printf("\nUser: %s\n", user.Login)
				fmt.Printf("Name: %s\n", user.Name)
				fmt.Printf("Company: %s\n", user.Company)
				fmt.Printf("Location: %s\n", user.Location)
				fmt.Printf("Followers: %d\n", user.Followers)
				fmt.Printf("(fetched in %v)\n", elapsed)
			} else {
				fmt.Printf("User not found!\n")
			}

		case "2":
			keys := redisDB.Keys(ctx, "*")
			if err = keys.Err(); err != nil {
				fmt.Printf("Error: Failed to list keys from Redis: %v\n", err)
				continue
			}

			keyList := keys.Val()
			fmt.Printf("Found %d users in Redis:\n", len(keyList))
			for _, key := range keyList {
				fmt.Printf("- %s\n", key)
			}

		case "3":
			fmt.Print("Enter username to remove from Redis: ")
			scanner.Scan()
			username := strings.TrimSpace(scanner.Text())
			if len(username) == 0 {
				fmt.Println("Error: Username cannot be empty")
				continue
			}

			del := redisDB.Del(ctx, username)
			if err = del.Err(); err != nil {
				fmt.Printf("Error: Failed to remove user from Redis: %v\n", err)
			} else {
				fmt.Println("Done")
			}

		case "4":
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
