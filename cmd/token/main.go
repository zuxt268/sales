package main

import (
	"fmt"
	"os"

	"github.com/zuxt268/sales/internal/auth"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: token <password>")
		os.Exit(1)
	}

	password := os.Args[1]
	token, err := auth.GenerateToken(password)
	if err != nil {
		fmt.Printf("Error generating token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(token)
}
