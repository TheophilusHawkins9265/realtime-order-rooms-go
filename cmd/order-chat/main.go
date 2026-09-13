package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"example.com/checkout-room/orderchat"
)

type server struct{ rooms *orderchat.OrderRoom }

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}

	s := server{rooms: orderchat.NewOrderRoom(orderchat.NewRealtimeClient(key))}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /rooms/open", s.openRoom)
	mux.HandleFunc("POST /orders/update", s.updateOrder)
	mux.HandleFunc("GET /orders/{orderID}/presence", s.presence)
	log.Println("order chat listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func (s server) openRoom(w http.ResponseWriter, r *http.Request) {
	var input struct {
		OrderID   string `json:"order_id"`
		ClientID  string `json:"client_id"`
		RequestID string `json:"request_id"`
	}
	if err := decode(r, &input); err != nil || input.OrderID == "" || input.ClientID == "" || input.RequestID == "" {
		http.Error(w, "order_id, client_id, and request_id are required", http.StatusBadRequest)
		return
	}
	token, err := s.rooms.Open(r.Context(), input.OrderID, input.ClientID, input.RequestID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusCreated, json.RawMessage(token))
}

func (s server) updateOrder(w http.ResponseWriter, r *http.Request) {
	var update orderchat.OrderUpdate
	if err := decode(r, &update); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := s.rooms.Apply(r.Context(), update); err != nil {
		if strings.HasPrefix(err.Error(), "invalid order transition") || strings.Contains(err.Error(), "required") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"order_id": update.OrderID, "stage": string(update.To)})
}

func (s server) presence(w http.ResponseWriter, r *http.Request) {
	data, err := s.rooms.Viewers(r.Context(), r.PathValue("orderID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, json.RawMessage(data))
}

func decode(r *http.Request, dst any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode response: %v", err)
	}
}
