package workerpool

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ReqHandler это мапа, в которой находятся типы запросов
var ReqHandler = map[int]RequestHandler{
	1: func(data interface{}) error {
		return nil
	},
}

// dispatcher управляет worker pool и распределяет входящие запросы между ними.
type dispatcher struct {
	inCh        chan Request
	wg          *sync.WaitGroup
	mu          sync.Mutex
	workerCount int
	stopCh      chan struct{} // канал, который сигнализирует для остановки воркера
}

// AddWorker добавляет нового воркера и увеличивает их клоичетво на один
func (d *dispatcher) AddWorker(w WorkerLauncher) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.workerCount++
	d.wg.Add(1)
	w.LaunchWorker(d.inCh, d.stopCh)
}

// RemoveWorker удаляет один воркер если количество воркеров больше, чем минимальное количетсво
func (d *dispatcher) RemoveWorker(minWorkers int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.workerCount > minWorkers {
		d.workerCount--
		d.stopCh <- struct{}{} // Signal a worker to stop
	}
}

// ScaleWorkers динамически регулирует количество работников в зависимости от нагрузки.
func (d *dispatcher) ScaleWorkers(minWorkers, maxWorkers, loadThreshold int) {
	ticker := time.NewTicker(time.Microsecond)
	defer ticker.Stop()

	for range ticker.C {
		load := len(d.inCh) //  load количество ожидающих запросов в канале
		if load > loadThreshold && d.workerCount < maxWorkers {
			fmt.Println("Scaling Triggered")
			newWorker := &Worker{
				Wg:         d.wg,
				Id:         d.workerCount,
				ReqHandler: ReqHandler,
			}
			d.AddWorker(newWorker)
		} else if float64(load) < 0.75*float64(loadThreshold) && d.workerCount > minWorkers {
			fmt.Println("Reducing Triggered")
			d.RemoveWorker(minWorkers)
		}
	}
}

// LaunchWorker запускает воркера и увеличивает счетчик на один
func (d *dispatcher) LaunchWorker(id int, w WorkerLauncher) {
	w.LaunchWorker(d.inCh, d.stopCh)
	d.mu.Lock()
	d.workerCount++
	d.mu.Unlock()
}

// MakeRequest добвляет запрос в канал, если канал полон то мы дропаемся
func (d *dispatcher) MakeRequest(r Request) {
	select {
	case d.inCh <- r:
	default:
		fmt.Println("Request channel is full. Dropping request.")
	}
}

// Graceful shutdown воркеров
func (d *dispatcher) Stop(ctx context.Context) {
	fmt.Println("\nstop called")
	close(d.inCh)
	done := make(chan struct{})

	go func() {
		d.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("All workers stopped gracefully")
	case <-ctx.Done():
		fmt.Println("Timeout reached, forcing shutdown")
		for i := 0; i < d.workerCount; i++ {
			d.stopCh <- struct{}{}
		}
	}

	d.wg.Wait()
}

// NewDispatcher - констуктор dispatcher
func NewDispatcher(b int, wg *sync.WaitGroup, maxWorkers int) Dispatcher {
	return &dispatcher{
		inCh:   make(chan Request, b),
		wg:     wg,
		stopCh: make(chan struct{}, maxWorkers),
	}
}
