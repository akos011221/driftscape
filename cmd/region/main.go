package main

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var (
	rdb    *redis.Client
	domain = "default.svc.cluster.local"
)

func main() {
	// Connect to Redis
	rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("redis.%s:6379", domain),
	})
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		fmt.Println("Redis connection failed:", err)
	}

	http.HandleFunc("/desc", descHandler)
	fmt.Println("Region running on :8081")
	http.ListenAndServe(":8081", nil)
}

// descHandler tells the Coordinator what this region looks like
func descHandler(w http.ResponseWriter, r *http.Request) {
	// Get coordinates from env vars
	xStr := os.Getenv("REGION_X")
	yStr := os.Getenv("REGION_Y")
	x, _ := strconv.Atoi(xStr)
	y, _ := strconv.Atoi(yStr)

	// Seed randomness with x,y for consistency
	h := fnv.New32a()
	h.Write([]byte(fmt.Sprintf("%d,%d", x, y)))
	seed := h.Sum32()
	// Custom rand to avoid time-based seeding,
	// so restarts don't change region
	rand := newRand(int64(seed))

	// Pick a terrain type
	places := []string{"forest", "plains", "hill", "swamp"}
	place := places[rand.Intn(len(places))]

	// Save the type to Redis
	key := fmt.Sprintf("region:%d,%d", x, y)
	rdb.Set(context.Background(), key, place, 0)

	fmt.Fprintf(w, "You're in a %s", place)
}

// newRand creates a seeded random generator
func newRand(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}
