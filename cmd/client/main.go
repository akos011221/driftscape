package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func main() {
	fmt.Println("Welcome to DriftScape!")
	fmt.Println("Commands: move north/south/east/west, look, quit")

	// Your position on the grid starts at the center (0,0)
	x, y := 0, 0

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
			look(x, y) // Shows where you are
		case "move":
			if len(words) < 2 { // Direction is not provided
				fmt.Println("Where? Use: move north/south/east/west")
				continue
			}
			direction := words[1]
			move(&x, &y, direction) // Updates your position and tells the Coordinator
		default:
			fmt.Println("Huh? Try: move north, look, or quit")
		}
	}
}

// look asks the Coordinator what's at your current spot (x,y)
func look(x, y int) {
	// Builds a web address like "http://coordinator:8080/look?x=0y=0"
	url := fmt.Sprintf("http://coordinator.default.svc.cluster.local:8080/look?x=%d&y=%d", x, y)
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
func move(x, y *int, direction string) {
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
	url := fmt.Sprintf("http://coordinator.default.svc.cluster.local:8080/move?x=%d&y=%d", newX, newY)
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
