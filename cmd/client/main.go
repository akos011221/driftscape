package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"strconv"
)

func main() {
	coordAddr := os.Getenv("COORDINATOR_ADDR")
	if coordAddr == "" {
		coordAddr = "http://localhost:8080" // Default for local testing
		fmt.Println("No COORDINATOR_ADDR set, using default:", coordAddr)
	}

	// Fetch starting position from Coordinator
	x, y, err := getStartingPosition(coordAddr)
	if err != nil {
		fmt.Println("Failed to get starting position, defaulting to (0,0):",err)
		x, y = 0, 0
	}

	fmt.Println("Welcome to DriftScape!")
	fmt.Println("Commands: move north/south/east/west, look, quit")

	// A loop to keep asking for commands
	scanner := bufio.NewScanner(os.Stdin) // Reads the keyboard input
	for {
		fmt.Print("> ")                // Shows a prompt to the user
		scanner.Scan()                 // Waits for Enter hit
		input := scanner.Text()        // Grabs what you typed as a string
		words := strings.Fields(input) // Splits the input into words

		// If you didn't type anything, skip and ask again
		if len(words) == 0 {
			continue
		}

		// The first word is the command (e.g., "move")
		command := words[0]

		// Decide what to do based on the command
		switch command {
		case "quit":
			fmt.Println("See you next time!")
			return
		case "look":
			look(coordAddr, x, y) // Shows where you are
		case "move":
			if len(words) < 2 { // Direction is not provided
				fmt.Println("Where? Use: move north/south/east/west")
				continue
			}
			direction := words[1]
			move(coordAddr, &x, &y, direction) // Updates your position and tells the Coordinator
		default:
			fmt.Println("Huh? Try: move north, look, or quit")
		}
	}
}

// getStartingPosition asks the Coordinator the starting spot
func getStartingPosition(coordAddr string) (int, int, error) {
	url := fmt.Sprintf("%s/position", coordAddr)
	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	posStr := string(buf[:n])
	x, y := parsePosition(posStr)
	return x, y, nil
}

// look asks the Coordinator what's at your current spot (x,y)
func look(coordAddr string, x, y int) {
	// Builds a web address like "http://coordinator:8080/look?x=0y=0"
	url := fmt.Sprintf("%s/look?x=%d&y=%d", coordAddr, x, y)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Can't see anything-world's not responding!")
		return
	}
	defer resp.Body.Close() // Cleans up after we're done

	// Reads the Coordinator's answer (e.g., "You are in a forest")
	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	fmt.Println(string(buf[:n])) // Cuts off empty space
}

// move updates your position and tells the Coordinator you moved
func move(coordAddr string, x, y *int, direction string) {
	newX, newY := *x, *y // Copies your current spot

	// Adjust position based on direction
	switch direction {
	case "north":
		newY++ // Up on the map
	case "south":
		newY-- // Down
	case "east":
		newX++ // Right
	case "west":
		newX-- // Left
	default:
		fmt.Println("Which way? Use: north, south, east, west")
		return
	}

	// Tell the Coordinator: "I'm moving to (newX, newY)"
	url := fmt.Sprintf("%s/move?x=%d&y=%d", coordAddr, newX, newY)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Can't move-world's not responding!")
		return
	}
	defer resp.Body.Close()

	// Read the response (e.g., "You're in a plains now")
	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	fmt.Println(string(buf[:n]))

	// If it worked, update your position
	*x, *y = newX, newY
}

// parsePosition converts (x,y) string to x, y int
func parsePosition(pos string) (int, int) {
	parts := strings.Split(pos, ",")
	x, _ := strconv.Atoi(parts[0])
	y, _ := strconv.Atoi(parts[1])
	return x, y
}
