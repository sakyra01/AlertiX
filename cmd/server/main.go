package main

import (
    // "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"
    // "runtime"
	"sync"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

// type RuntimeMetrics struct {
//     NumGoroutine   int
//     NumGC          uint32
//     Alloc          uint64
//     TotalAlloc     uint64
//     Sys            uint64
//     Lookups        uint64
//     Mallocs        uint64
//     Frees          uint64
//     HeapAlloc      uint64
//     HeapSys        uint64
//     HeapIdle       uint64
//     HeapInuse      uint64
//     HeapReleased   uint64
//     HeapObjects    uint64
//     StackInuse     uint64
//     StackSys       uint64
//     MSpanInuse     uint64
//     MSpanSys       uint64
//     MCacheInuse    uint64
//     MCacheSys      uint64
//     BuckHashSys    uint64
//     GCSys          uint64
//     OtherSys       uint64
// }

// Определяем тип метрик:
type MetricType string

const (
    Gauge   MetricType = "gauge"
    Counter MetricType = "counter"
)

// Структура MetricsStore
type MemStorage struct{
	mu       sync.RWMutex
    gauges   map[string]float64
    counters map[string]int64
}

func NewMetricStore() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
        counters: make(map[string]int64),
	}
}

//	Обновление метрик
func (metrics *MemStorage) UpdateMetric(mType MetricType, name string, value string) error {
    metrics.mu.Lock()
    defer metrics.mu.Unlock()

    switch mType {
    case Gauge:
        v, err := strconv.ParseFloat(value, 64)
        if err != nil {
            return fmt.Errorf("invalid gauge value")
        }
        metrics.gauges[name] = v
    case Counter:
        v, err := strconv.ParseInt(value, 10, 64)
        if err != nil {
            return fmt.Errorf("invalid counter value")
        }
        metrics.counters[name] += v
    default:
        return fmt.Errorf("invalid metric type")
    }
    return nil
}

// HTTP-обработчик для обновления метрик
func (metrics *MemStorage) updateMetricHandler(w http.ResponseWriter, r *http.Request) {
    mType := MetricType(chi.URLParam(r, "type"))
    name := chi.URLParam(r, "name")
    value := chi.URLParam(r, "value")

    if name == "" {
        http.Error(w, "metric name not provided", http.StatusNotFound)
        return
    }
    if err := metrics.UpdateMetric(mType, name, value); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    w.WriteHeader(http.StatusOK)
}


func main() {
	store := NewMetricStore()

    r := chi.NewRouter() // Создаем новый маршрутизатор
    r.Use(middleware.Logger) // middleware для логирования запрсов

	// Определяем маршруты для обновления метрик
	r.Post("/update/{type}/{name}/{value}", store.updateMetricHandler)

    log.Println("Server is running on port 8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}