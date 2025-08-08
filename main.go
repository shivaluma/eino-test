// Root main.go - Simple wrapper for backwards compatibility
// Actual server implementation is in cmd/server/main.go

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	// Check if this is being run from the project root
	if _, err := os.Stat("cmd/server/main.go"); os.IsNotExist(err) {
		log.Fatal("Error: This command must be run from the project root directory")
	}

	// Run the actual server
	fmt.Println("Starting Eino Agent server...")
	cmd := exec.Command("go", "run", "cmd/server/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}