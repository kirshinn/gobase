package main

import (
	"fmt"
	"time"
)

func main() {
	getPage()
}

func getPage() {
	resultCh := getComments()

	time.Sleep(1 * time.Second)
	fmt.Println("get related articles")

	commentsData := <-resultCh
	fmt.Println("main goroutine", commentsData)
}

func getComments() chan string {
	// используем буферизированный канал с 1 значением
	result := make(chan string, 1)
	go func(out chan<- string) {
		time.Sleep(2 * time.Second)
		fmt.Println("async operation ready, get comments")
		out <- "32 комментария"
	}(result)
	return result
}
