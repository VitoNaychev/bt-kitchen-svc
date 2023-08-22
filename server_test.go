package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type StubKitchenStore struct {
	tickets []Ticket
}

func (s *StubKitchenStore) GetTicketByID(ticketID int) (Ticket, error) {
	for _, ticket := range s.tickets {
		if ticket.ID == ticketID {
			return ticket, nil
		}
	}

	return Ticket{}, fmt.Errorf("no ticket with ID = %d", ticketID)
}

func (s *StubKitchenStore) StoreTicket(ticket Ticket) (int, error) {
	ticket.ID = len(s.tickets)
	s.tickets = append(s.tickets, ticket)

	return ticket.ID, nil
}

func TestGETTicket(t *testing.T) {
	store := &StubKitchenStore{
		[]Ticket{
			{
				ID:     1,
				Status: STATUS_ACCEPTED,
				Items:  []string{"burger", "fries"},
			},
			{
				ID:     2,
				Status: STATUS_PENDING,
				Items:  []string{"pizza", "water"},
			},
		},
	}
	server := KitchenServer{store}
	t.Run("returns OK on valid ticket ID", func(t *testing.T) {
		request := newGetTicketRequest(1)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
	})

	t.Run("returns Bad Request on invalid ticket ID", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/ticket/asdff", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("returns ticket in JSON format when ID = 1", func(t *testing.T) {
		request := newGetTicketRequest(1)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)

		want := store.tickets[0]
		got := getTicketFromResponse(t, response.Body)

		assertTicket(t, got, want)
	})

	t.Run("returns ticket in JSON format when ID = 2", func(t *testing.T) {
		request := newGetTicketRequest(2)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)

		want := store.tickets[1]
		got := getTicketFromResponse(t, response.Body)

		assertTicket(t, got, want)
	})

	t.Run("returns Not Found on nonexistant ticket ID", func(t *testing.T) {
		request := newGetTicketRequest(3)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusNotFound)
	})
}

func TestCreateTicket(t *testing.T) {
	store := &StubKitchenStore{}
	server := KitchenServer{store}
	t.Run("returns Accepted on valid ticket JSON", func(t *testing.T) {
		ticket := Ticket{
			Items: []string{"burger", "fries"},
		}

		request := newCreateTicketRequest(ticket)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusAccepted)
	})

	t.Run("returns Bad Request on invalid ticket JSON", func(t *testing.T) {
		ticket := `{"text": "this is an invalid ticket JSON"}`
		buffer := bytes.NewBuffer([]byte(ticket))

		request, _ := http.NewRequest(http.MethodPost, "/ticket/", buffer)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("returns ticket ID on valid ticket JSON", func(t *testing.T) {
		ticket := Ticket{
			Items: []string{"burger", "fries"},
		}

		request := newCreateTicketRequest(ticket)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusAccepted)
		assertTicketResponse(t, response.Body, CreateTicketResponse{ID: 1})
	})

	t.Run("persists ticket and sets status to STATUS_PENDING", func(t *testing.T) {
		ticket := Ticket{
			Items: []string{"pizza", "water"},
		}

		request := newCreateTicketRequest(ticket)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusAccepted)
		assertTicketResponse(t, response.Body, CreateTicketResponse{ID: 2})

		want := Ticket{
			ID:     2,
			Status: STATUS_PENDING,
			Items:  ticket.Items,
		}
		assertTicketPersisted(t, store, want)
	})
}

func assertTicketPersisted(t testing.TB, store *StubKitchenStore, want Ticket) {
	t.Helper()

	got, err := store.GetTicketByID(want.ID)
	if err != nil {
		t.Errorf("server didn't persist order, %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("server didn't persist correct order, got %v want %v", got, want)
	}
}

func assertTicketResponse(t testing.TB, body io.Reader, want CreateTicketResponse) {
	ticketResponse := CreateTicketResponse{}
	err := json.NewDecoder(body).Decode(&ticketResponse)
	if err != nil {
		t.Errorf("unable to parse response from server %q into Ticket, %v", body, err)
	}

	if ticketResponse.ID != want.ID {
		t.Errorf("didn't receive correct Ticket ID, got %v want %v", ticketResponse.ID, want.ID)
	}
}

func newCreateTicketRequest(ticket Ticket) *http.Request {
	buffer := &bytes.Buffer{}
	json.NewEncoder(buffer).Encode(ticket)

	req, _ := http.NewRequest(http.MethodPost, "/ticket/", buffer)
	return req
}

func newGetTicketRequest(ticketID int) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/ticket/%d", ticketID), nil)
	return req
}

func assertStatus(t testing.TB, got, want int) {
	if got != want {
		t.Errorf("got status %v, want %v", got, want)
	}
}

func getTicketFromResponse(t testing.TB, body io.Reader) Ticket {
	t.Helper()

	ticket := Ticket{}
	err := json.NewDecoder(body).Decode(&ticket)
	if err != nil {
		t.Fatalf("Unable to parse response from server %q into Order, %v", body, err)
	}

	return ticket
}

func assertTicket(t testing.TB, got, want Ticket) {
	t.Helper()

	if !reflect.DeepEqual(want, got) {
		t.Errorf("got %v, want %v", got, want)
	}
}
