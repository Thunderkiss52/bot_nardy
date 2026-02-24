//go:build !wails

package main

import "fmt"

func main() {
	fmt.Println("Desktop build requires Wails and tag 'wails'.")
	fmt.Println("Run: go run -tags wails .")
}
