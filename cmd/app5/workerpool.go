package main

import "time"

func main() {
	workerInput := make(chan string, 2)

	for i := 0; i < gourutinesNum; i++ {
		go startWorker(i, workerInput)
	}

	months := []string{
		"январь",
		"февраль",
		"март",
		"апрель",
		"май",
		"июнь",
		"июль",
		"август",
		"сентябрь",
		"октябрь",
		"ноябрь",
		"декабрь",
	}

	for _, monthName := range months {
		workerInput <- monthName
	}

	time.Sleep(time.Millisecond)
}

func startWorker(workerNum int, input <-chan string) {
	
}
