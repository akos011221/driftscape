package main

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"net/http"
	"os"
	"strconv"
)

func main() {
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
	fmt.Fprintf(w, "You're in a %s", place)
}

// newRand creates a seeded random generator
func newRand(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}
