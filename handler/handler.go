// Package handler exposes a web API to submit new events or get current usage
// statistics.
package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/orlangure/testing-kafka-in-go/reporter"
)

func Mux(prod Producer, rep Reporter) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		var e reporter.Event

		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if e.Type == "" {
			http.Error(w, "missing event type", http.StatusBadRequest)
			return
		}

		key := []byte(e.Type)
		e.Type = ""

		value, err := json.Marshal(e)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := prod.Produce(r.Context(), key, value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	})

	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		rep, err := rep.Report()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(rep); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	return mux
}

type Producer interface {
	Produce(context.Context, []byte, []byte) error
}

type Reporter interface {
	Report() (interface{}, error)
}
