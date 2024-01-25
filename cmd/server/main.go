package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type gauge float64
type counter int

type MemStorage struct {
	metrics map[string]interface{} // словарь, где используются строки в качестве ключей и хранятся значения различных типов intreface{}
}

func main() {
	storage := NewMemStorage()

	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, func(w http.ResponseWriter, r *http.Request) {
		handleMetrics(w, r, storage)
	})

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]interface{}),
	}
}

func handleMetrics(w http.ResponseWriter, r *http.Request, s *MemStorage) {
	// println(r.URL.Path)
	if r.Method != http.MethodPost {
		http.Error(w, "only POST method allowed", http.StatusBadRequest)
		return
	}

	urlParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/update/"), "/")

	if len(urlParts) != 3 {
		http.Error(w, "incorrect request", http.StatusNotFound)
		return
	}

	// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	metricType := urlParts[0]
	metricName := urlParts[1]
	metricValueStr := urlParts[2]

	if metricName == "" {
		http.Error(w, "invalid metric name", http.StatusNotFound)
		return
	}

	// if metricValueStr == "" {
	// 	http.Error(w, "Invalid metric value", http.StatusBadRequest)
	// 	return
	// }

	var metricValue interface{}
	var err error

	switch metricType {
	case "gauge":

		metricValue, err = strconv.ParseFloat(metricValueStr, 64)

		if err != nil {
			err = fmt.Errorf("invalid gauge metric value: %s", metricValueStr)
		}

		s.metrics[metricName] = metricValue.(float64)
	case "counter":

		metricValue, err = strconv.ParseInt(metricValueStr, 10, 0)

		if err != nil {
			err = fmt.Errorf("invalid counter metric value: %s", metricValueStr)
		}

		prevValue, exists := s.metrics[metricName]
		if exists {
			currentValue, ok := prevValue.(int64)
			if ok {
				s.metrics[metricName] = prevValue.(int64) + currentValue
			} else {
				err = fmt.Errorf("stored value of counter metrics is not int64")
			}
		} else {
			s.metrics[metricName] = metricValue.(int64)
		}

	default:
		err = fmt.Errorf("invalid metric type: %s", metricType)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
