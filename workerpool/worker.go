package workerpool

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Worker struct {
	Id         int
	Wg         *sync.WaitGroup
	ReqHandler map[int]RequestHandler
}

// Launch Worker запускает worker для обработки входящих запросов.
// Он выполняется в отдельной программе, непрерывно отслеживающей входящие запросы по входному каналу.
// Worker автоматически останавливается, когда либо входной канал закрыт, либо он получает сигнал остановки.
func (w *Worker) LaunchWorker(in chan Request, stopCh chan struct{}) {
	go func() {
		defer w.Wg.Done()
		for {
			select {
			case msg, open := <-in:
				if !open {
					// Если канал закрыт, остановите обработку и вернитесь обратно
					// если мы пропустим проверку закрытия канала, то после закрытия канала
					// worker продолжит считывать пустые значения из закрытого канала.
					fmt.Println("Stopping worker:", w.Id)
					return
				}
				w.processRequest(msg)
				time.Sleep(1 * time.Microsecond)
			case <-stopCh:
				fmt.Println("Stopping worker:", w.Id)
				return
			}
		}
	}()
}

// processRequest обрабатывает отдельный запрос
func (w *Worker) processRequest(msg Request) {
	fmt.Printf("Worker %d processing request: %v\n", w.Id, msg)
	var handler RequestHandler
	var ok bool
	if handler, ok = w.ReqHandler[msg.Type]; !ok {
		fmt.Println("Handler not implemented: workerID:", w.Id)
	} else {
		if msg.Timeout == 0 {
			msg.Timeout = time.Duration(10 * time.Millisecond)
		}
		for attempt := 0; attempt <= msg.MaxRetries; attempt++ {
			var err error
			done := make(chan struct{})
			ctx, cancel := context.WithTimeout(context.Background(), msg.Timeout)
			defer cancel()

			go func() {
				err = handler(msg.Data)
				close(done)
			}()

			select {
			case <-done:
				if err == nil {
					return
				}
				fmt.Printf("Worker %d: Error processing request: %v\n", w.Id, err)
			case <-ctx.Done():
				fmt.Printf("Worker %d: Timeout processing request: %v\n", w.Id, msg.Data)
			}
			fmt.Printf("Worker %d: Retry %d for request %v\n", w.Id, attempt, msg.Data)
		}
		fmt.Printf("Worker %d: Failed to process request %v after %d retries\n", w.Id, msg.Data, msg.MaxRetries)
	}
}
