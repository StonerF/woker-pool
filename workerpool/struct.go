package workerpool

import "time"

// Request- представляет собой запрос, который должен быть обработан worker-ом.
type Request struct {
	Handler    RequestHandler
	Type       int
	Data       interface{}
	Timeout    time.Duration // максимальное время для обработки
	Retries    int           // количетсво ретраев
	MaxRetries int           // максимальное количетво ретраев
}

// RequestHandler определяет тип функции для обработки запросов.
type RequestHandler func(interface{}) error
