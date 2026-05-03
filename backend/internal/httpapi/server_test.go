package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"surplusslot/backend/internal/domain"
	"surplusslot/backend/internal/store"
)

type mockStore struct{}

func (m *mockStore) ListActiveOffers(ctx context.Context) ([]domain.Offer, error) {
	return []domain.Offer{{ID: "o1", Merchant: "Cafe", Title: "Box", PriceCents: 500, Stock: 2, PickupStart: time.Now(), PickupEnd: time.Now().Add(time.Hour), Status: "active"}}, nil
}
func (m *mockStore) CreateOffer(ctx context.Context, in domain.Offer) (domain.Offer, error) {
	if in.Title == "" {
		return domain.Offer{}, store.ErrInvalidInput
	}
	in.ID = "o1"
	in.Status = "active"
	return in, nil
}
func (m *mockStore) CreateOrder(ctx context.Context, offerID, customerName, customerEmail string, qty int) (domain.Order, error) {
	if qty > 2 {
		return domain.Order{}, store.ErrOutOfStock
	}
	return domain.Order{ID: "ord1", OfferID: offerID, CustomerName: customerName, CustomerEmail: customerEmail, Quantity: qty, TotalCents: 1000, Status: "placed", PickupCode: "ABC123"}, nil
}
func (m *mockStore) ConfirmPickup(ctx context.Context, orderID string) (domain.Order, error) {
	return domain.Order{ID: orderID, Status: "picked_up"}, nil
}
func (m *mockStore) RecordEvent(ctx context.Context, name string, properties map[string]any) error {
	return nil
}

func TestCreateOrderOutOfStock(t *testing.T) {
	s := NewServer(&mockStore{})
	body := map[string]any{
		"offerId":       "o1",
		"customerName":  "Ada",
		"customerEmail": "ada@example.com",
		"quantity":      5,
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(b))
	rec := httptest.NewRecorder()

	s.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestListOffers(t *testing.T) {
	s := NewServer(&mockStore{})
	req := httptest.NewRequest(http.MethodGet, "/api/offers", nil)
	rec := httptest.NewRecorder()

	s.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
