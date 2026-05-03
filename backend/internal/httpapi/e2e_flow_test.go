package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"surplusslot/backend/internal/domain"
	"surplusslot/backend/internal/store"
)

type memoryStore struct {
	mu     sync.Mutex
	offers map[string]domain.Offer
	orders map[string]domain.Order
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		offers: map[string]domain.Offer{},
		orders: map[string]domain.Order{},
	}
}

func (m *memoryStore) ListActiveOffers(ctx context.Context) ([]domain.Offer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.Offer, 0, len(m.offers))
	for _, o := range m.offers {
		if o.Status == "active" && o.Stock > 0 {
			out = append(out, o)
		}
	}
	return out, nil
}

func (m *memoryStore) CreateOffer(ctx context.Context, in domain.Offer) (domain.Offer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if in.Merchant == "" || in.Title == "" || in.Stock <= 0 || in.PriceCents <= 0 {
		return domain.Offer{}, store.ErrInvalidInput
	}
	in.ID = "offer-1"
	in.Status = "active"
	m.offers[in.ID] = in
	return in, nil
}

func (m *memoryStore) CreateOrder(ctx context.Context, offerID, customerName, customerEmail string, qty int) (domain.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.offers[offerID]
	if !ok {
		return domain.Order{}, store.ErrNotFound
	}
	if o.Stock < qty {
		return domain.Order{}, store.ErrOutOfStock
	}
	o.Stock -= qty
	m.offers[offerID] = o

	order := domain.Order{
		ID:            "order-1",
		OfferID:       offerID,
		CustomerName:  customerName,
		CustomerEmail: customerEmail,
		Quantity:      qty,
		TotalCents:    qty * o.PriceCents,
		Status:        "placed",
		PickupCode:    "PICK01",
	}
	m.orders[order.ID] = order
	return order, nil
}

func (m *memoryStore) ConfirmPickup(ctx context.Context, orderID string) (domain.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	order, ok := m.orders[orderID]
	if !ok {
		return domain.Order{}, store.ErrNotFound
	}
	if order.Status == "picked_up" {
		return domain.Order{}, store.ErrAlreadyPicked
	}
	order.Status = "picked_up"
	m.orders[orderID] = order
	return order, nil
}

func (m *memoryStore) RecordEvent(ctx context.Context, name string, properties map[string]any) error {
	return nil
}

func TestE2EFlow_CreateOfferReserveConfirmPickup(t *testing.T) {
	server := NewServer(newMemoryStore())
	routes := server.Routes()

	createOfferBody := map[string]any{
		"merchant":    "Cafe A",
		"title":       "Pastry Box",
		"description": "Fresh leftovers",
		"priceCents":  750,
		"stock":       3,
		"pickupStart": time.Now().Add(1 * time.Hour).Format(time.RFC3339),
		"pickupEnd":   time.Now().Add(3 * time.Hour).Format(time.RFC3339),
	}
	createOfferJSON, _ := json.Marshal(createOfferBody)
	createOfferReq := httptest.NewRequest(http.MethodPost, "/api/offers", bytes.NewReader(createOfferJSON))
	createOfferRec := httptest.NewRecorder()
	routes.ServeHTTP(createOfferRec, createOfferReq)
	if createOfferRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create offer, got %d", createOfferRec.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/offers", nil)
	listRec := httptest.NewRecorder()
	routes.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on list offers, got %d", listRec.Code)
	}

	createOrderBody := map[string]any{
		"offerId":       "offer-1",
		"customerName":  "Ada",
		"customerEmail": "ada@example.com",
		"quantity":      1,
	}
	createOrderJSON, _ := json.Marshal(createOrderBody)
	createOrderReq := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(createOrderJSON))
	createOrderRec := httptest.NewRecorder()
	routes.ServeHTTP(createOrderRec, createOrderReq)
	if createOrderRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create order, got %d", createOrderRec.Code)
	}

	pickupReq := httptest.NewRequest(http.MethodPost, "/api/orders/order-1/pickup", nil)
	pickupRec := httptest.NewRecorder()
	routes.ServeHTTP(pickupRec, pickupReq)
	if pickupRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on confirm pickup, got %d", pickupRec.Code)
	}
}
