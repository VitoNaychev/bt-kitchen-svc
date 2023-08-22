package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const (
	STATUS_PENDING int = iota
	STATUS_ACCEPTED
	STATUS_COMPLETED
)

type Ticket struct {
	ID     int
	Status int
	Items  []string
}

type CreateTicketResponse struct {
	ID int
}

type KitchenStore interface {
	GetTicketByID(int) (Ticket, error)
	StoreTicket(Ticket) (int, error)
}

type KitchenServer struct {
	store KitchenStore
}

func (k *KitchenServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		k.getTicket(w, r)
	case http.MethodPost:
		k.createTicket(w, r)
	}
}

func (k *KitchenServer) getTicket(w http.ResponseWriter, r *http.Request) {
	stringID := strings.TrimPrefix(r.URL.Path, "/ticket/")
	ticketID, err := strconv.Atoi(stringID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ticket, err := k.store.GetTicketByID(ticketID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ticket)
}

func (k *KitchenServer) createTicket(w http.ResponseWriter, r *http.Request) {
	ticket, err := getTicketFromRequestBody(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ticket.Status = STATUS_PENDING
	id, err := k.store.StoreTicket(*ticket)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(CreateTicketResponse{ID: id})
}

func getTicketFromRequestBody(body io.Reader) (*Ticket, error) {
	d := json.NewDecoder(body)
	d.DisallowUnknownFields()

	ticket := Ticket{}
	err := d.Decode(&ticket)

	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal ticket JSON, %v", err)
	}

	if !isTicketValid(ticket) {
		return nil, fmt.Errorf("some fields of ticket JSON are empty, cannot persist")
	}

	return &ticket, nil
}

func isTicketValid(ticket Ticket) bool {
	if ticket.Items == nil {
		return false
	}

	return true
}
