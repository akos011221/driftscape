package main

import (
	"fmt"
	"net/http"
	"strconv"
)

// A map to track, where the key is the region (x,y), value is the region's pod name
var regions = make(map[string]string)

func main() {
	http.HandleFunc("/look", lookHandler)
	http.HandleFunc("/move", moveHandler)

	fmt.Println("Coordinator running on :8080")
	http.ListenAndServe(":8080", nil)
}

// lookHandler responds when the Client asks what's at (x,y)
func lookHandler(w http.ResponseWriter, r *http.Request) {
	// Gets x and y from the request (e.g., "?x=0&y=1")
	xStr := r.URL.Query().Get("x")
	yStr := r.URL.Query().Get("y")

	x, err := strconv.Atoi(xStr)
	if err != nil {
		http.Error(w, "Bad x!", 400)
		return
	}
	y, err := strconv.Atoi(yStr)
	if err != nil {
		http.Error(w, "Bad y!", 400)
		return
	}

	key := fmt.Sprintf("%d,%d", x, y)

	if _, exists := regions[key]; !exists {
		// No real pods yet.
		regions[key] = "region-" + key
	}

	// Ask the region what it looks like (fake URL for now)
	url := fmt.Sprintf("http://%s:8081/desc", regions[key])
	resp, err := http.Get(url)
	if err != nil {
		// Fake response for now.
		fmt.Fprintf(w, "You're near the ocean at (%d,%d)", x, y)
		return
	}
	defer resp.Body.Close()

	// Pass the region's description back to the Client
	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	w.Write(buf[:n])
}

// moveHandler handles when you move to a new spot
func moveHandler(w http.ResponseWriter, r *http.Request) {
	xStr := r.URL.Query().Get("x")
	yStr := r.URL.Query().Get("y")

	x, err := strconv.Atoi(xStr)
	if err != nil {
		http.Error(w, "Bad x!", 400)
		return
	}
	y, err := strconv.Atoi(yStr)
	if err != nil {
		http.Error(w, "Bad y!", 400)
		return
	}

	key := fmt.Sprintf("%d,%d", x, y)
	if _, exists := regions[key]; !exists {
		regions[key] = "region-" + key // Fake pod spawn
	}

	// Tell the Client what's there
	url := fmt.Sprintf("http://%s:8080/desc", regions[key])
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(w, "You moved to a quiet place at (%d,%d)", x, y)
		return
	}
	defer resp.Body.Close()

	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	w.Write(buf[:n])
}
