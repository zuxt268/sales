package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	fmt.Println("industry")

	resp, err := http.Get("https://taido-fc.com")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(resp.StatusCode)
	fmt.Println(string(body))
}
