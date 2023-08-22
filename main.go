package main

import (
	"log"
	"net/http"
)

type InMemoryKitchenStore struct{}

func (i *InMemoryKitchenStore) GetTicketByID(ticketId int) (Ticket, error) {
	return Ticket{}, nil
}

func (i *InMemoryKitchenStore) StoreTicket(Ticket) (int, error) {
	return 123, nil
}

func main() {
	store := &InMemoryKitchenStore{}
	server := &KitchenServer{store}

	log.Fatal(http.ListenAndServe(":5000", server))
}
