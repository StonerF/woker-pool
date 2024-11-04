package workerpool

import "context"

// WorkerLauncher интерфейс для запуска worker-ов
type WorkerLauncher interface {
	LaunchWorker(in chan Request, stopCh chan struct{})
}

// Dispatcher интерфейс для управления пулом worker-ов.
type Dispatcher interface {
	AddWorker(w WorkerLauncher)
	RemoveWorker(minWorkers int)
	LaunchWorker(id int, w WorkerLauncher)
	ScaleWorkers(minWorkers, maxWorkers, loadThreshold int)
	MakeRequest(Request)
	Stop(ctx context.Context)
}
