package main

import (
	"fmt"
	"math/rand"
	"net/http"
)

func main() {
	http.HandleFunc("/desc", descHandler)
	fmt.Println("Region running on :8081")
	http.ListenAndServe(":8081", nil)
}

// descHandler tells the Coordinator what this region looks like
func descHandler(w http.ResponseWriter, r *http.Request) {
	// Pick a random place
	places := []string{"forest", "plains", "hill", "swamp"}
	place := places[rand.Intn(len(places))]

	fmt.Fprintf(w, "You're in a %s", place)
}
