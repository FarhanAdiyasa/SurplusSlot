package store

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"surplusslot/backend/internal/domain"

	"github.com/google/uuid"
)

type MemoryStore struct {
	mu     sync.RWMutex
	offers map[string]domain.Offer
	orders map[string]domain.Order
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		offers: make(map[string]domain.Offer),
		orders: make(map[string]domain.Order),
	}
}

func (s *MemoryStore) CreateOffer(ctx context.Context, in domain.Offer) (domain.Offer, error) {
	if strings.TrimSpace(in.Merchant) == "" || strings.TrimSpace(in.Title) == "" || in.PriceCents <= 0 || in.Stock <= 0 {
		return domain.Offer{}, ErrInvalidInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	in.ID = uuid.NewString()
	in.Status = "active"
	in.CreatedAt = time.Now()
	in.UpdatedAt = time.Now()
	s.offers[in.ID] = in
	return in, nil
}

func (s *MemoryStore) ListActiveOffers(ctx context.Context) ([]domain.Offer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]domain.Offer, 0)
	for _, o := range s.offers {
		if o.Status == "active" && o.Stock > 0 {
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].PickupStart.Before(out[j].PickupStart)
	})
	return out, nil
}

func (s *MemoryStore) CreateOrder(ctx context.Context, offerID, customerName, customerEmail string, qty int) (domain.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	o, ok := s.offers[offerID]
	if !ok {
		return domain.Order{}, ErrNotFound
	}
	if o.Stock < qty {
		return domain.Order{}, ErrOutOfStock
	}

	o.Stock -= qty
	o.UpdatedAt = time.Now()
	s.offers[offerID] = o

	order := domain.Order{
		ID:            uuid.NewString(),
		OfferID:       offerID,
		CustomerName:  customerName,
		CustomerEmail: customerEmail,
		Quantity:      qty,
		TotalCents:    qty * o.PriceCents,
		Status:        "placed",
		PickupCode:    strings.ToUpper(uuid.NewString()[:6]),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	s.orders[order.ID] = order
	return order, nil
}

func (s *MemoryStore) ConfirmPickup(ctx context.Context, orderID string) (domain.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[orderID]
	if !ok {
		return domain.Order{}, ErrNotFound
	}
	if order.Status == "picked_up" {
		return domain.Order{}, ErrAlreadyPicked
	}

	order.Status = "picked_up"
	order.UpdatedAt = time.Now()
	s.orders[orderID] = order
	return order, nil
}

func (s *MemoryStore) RecordEvent(ctx context.Context, name string, properties map[string]any) error {
	return nil
}
