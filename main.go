package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	wp "github.com/StonerF/workerpool"
)

func main() {

	// Configuration
	bufferSize := 50000
	maxWorkers := 20
	minWorkers := 3
	loadThreshold := 40000
	requests := 5

	var wg sync.WaitGroup
	dispatcher := wp.NewDispatcher(bufferSize, &wg, maxWorkers)

	// инициализируем воркеры
	for i := 0; i < minWorkers; i++ {
		fmt.Printf("Starting worker with id %d\n", i)
		w := &wp.Worker{
			Wg:         &wg,
			Id:         i,
			ReqHandler: wp.ReqHandler,
		}
		dispatcher.AddWorker(w)
	}

	//запускаем scale
	go dispatcher.ScaleWorkers(minWorkers, maxWorkers, loadThreshold)

	// отправляем запросы
	for i := 0; i < requests; i++ {
		req := wp.Request{
			Data:    fmt.Sprintf("(Msg_id: %d) -> Hello", i),
			Handler: func(result interface{}) error { return nil },
			Type:    1,
			Timeout: 5 * time.Second,
		}
		dispatcher.MakeRequest(req)
	}

	// Gracefule shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	dispatcher.Stop(ctx)
	fmt.Println("Exiting main!")
}
