package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"sync"
	"time"
)

// randomUsername selects a random username from a list
func randomUsername() string {
	usernames := []string{"user1", "user2", "user3", "user4", "user5"}
	rand.Seed(time.Now().UnixNano())
	return usernames[rand.Intn(len(usernames))]
}

// runSSHCommand executes an SSH command using the system's SSH CLI
func runSSHCommand(username, host string, port int) error {
	// The SSH command to execute (you might need to configure the SSH binary to avoid interactive prompts)
	cmd := exec.Command("ssh", fmt.Sprintf("%s@%s", username, host), "-p", fmt.Sprintf("%d", port))

	// Get the output from the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run SSH command: %w | output: %s", err, output)
	}

	log.Printf("host: %s | port: %d", host, port)
	return nil
}

func main() {
	// Define command-line flags for externalizing the number of connections and concurrency level
	numConnections := flag.Int("connections", 10, "Total number of SSH connections to establish")
	maxConcurrent := flag.Int("concurrent", 3, "Maximum number of concurrent SSH connections")
	host := flag.String("host", "example.com", "ssh host")
	port := flag.Int("port", 80, "ssh port")

	// Parse the flags
	flag.Parse()

	// Print the configuration for confirmation
	log.Printf("Starting with %d total connections and %d concurrent connections", *numConnections, *maxConcurrent)

	var wg sync.WaitGroup
	// Limit the number of concurrent connections
	semaphore := make(chan struct{}, *maxConcurrent)

	// Counter for successful connections
	var successfulConnections int
	var mu sync.Mutex

	for i := 0; i < *numConnections; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			// Limit concurrency
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Random username for each connection
			username := randomUsername()

			// Run the SSH command using the system's SSH client
			if err := runSSHCommand(username, *host, *port); err != nil {
				log.Printf("Failed to connect as %s: %v", username, err)
			} else {
				mu.Lock()
				successfulConnections++
				mu.Unlock()
			}
		}()
	}

	// Wait for all connections to finish
	wg.Wait()

	// Output total number of successful connections
	fmt.Printf("Total successful connections: %d/%d\n", successfulConnections, *numConnections)
}
