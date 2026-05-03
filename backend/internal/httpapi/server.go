package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"surplusslot/backend/internal/domain"
	"surplusslot/backend/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	store Store
}

type Store interface {
	ListActiveOffers(ctx context.Context) ([]domain.Offer, error)
	CreateOffer(ctx context.Context, in domain.Offer) (domain.Offer, error)
	CreateOrder(ctx context.Context, offerID, customerName, customerEmail string, qty int) (domain.Order, error)
	ConfirmPickup(ctx context.Context, orderID string) (domain.Order, error)
	RecordEvent(ctx context.Context, name string, properties map[string]any) error
}

func NewServer(st Store) *Server {
	return &Server{store: st}
}

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(cors)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Get("/api/offers", s.listOffers)
	r.Post("/api/offers", s.createOffer)
	r.Post("/api/orders", s.createOrder)
	r.Post("/api/orders/{id}/pickup", s.confirmPickup)
	r.Post("/api/events", s.trackEvent)

	return r
}

func (s *Server) listOffers(w http.ResponseWriter, r *http.Request) {
	offers, err := s.store.ListActiveOffers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list offers")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"offers": offers})
}

type createOfferReq struct {
	Merchant    string `json:"merchant"`
	Title       string `json:"title"`
	Description string `json:"description"`
	PriceCents  int    `json:"priceCents"`
	Stock       int    `json:"stock"`
	PickupStart string `json:"pickupStart"`
	PickupEnd   string `json:"pickupEnd"`
}

func (s *Server) createOffer(w http.ResponseWriter, r *http.Request) {
	var req createOfferReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request payload")
		return
	}
	start, err := store.ParseTime(req.PickupStart)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_pickup_start", "pickupStart must be RFC3339")
		return
	}
	end, err := store.ParseTime(req.PickupEnd)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_pickup_end", "pickupEnd must be RFC3339")
		return
	}

	offer, err := s.store.CreateOffer(r.Context(), domain.Offer{
		Merchant:    req.Merchant,
		Title:       req.Title,
		Description: req.Description,
		PriceCents:  req.PriceCents,
		Stock:       req.Stock,
		PickupStart: start,
		PickupEnd:   end,
	})
	if err != nil {
		mapStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"offer": offer})
}

type createOrderReq struct {
	OfferID       string `json:"offerId"`
	CustomerName  string `json:"customerName"`
	CustomerEmail string `json:"customerEmail"`
	Quantity      int    `json:"quantity"`
}

func (s *Server) createOrder(w http.ResponseWriter, r *http.Request) {
	var req createOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request payload")
		return
	}
	order, err := s.store.CreateOrder(r.Context(), req.OfferID, req.CustomerName, req.CustomerEmail, req.Quantity)
	if err != nil {
		mapStoreErr(w, err)
		return
	}
	_ = s.store.RecordEvent(r.Context(), "order_created", map[string]any{"orderId": order.ID, "offerId": order.OfferID})
	writeJSON(w, http.StatusCreated, map[string]any{"order": order})
}

func (s *Server) confirmPickup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	order, err := s.store.ConfirmPickup(r.Context(), id)
	if err != nil {
		mapStoreErr(w, err)
		return
	}
	_ = s.store.RecordEvent(r.Context(), "pickup_confirmed", map[string]any{"orderId": order.ID})
	writeJSON(w, http.StatusOK, map[string]any{"order": order})
}

func (s *Server) trackEvent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name       string         `json:"name"`
		Properties map[string]any `json:"properties"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeError(w, http.StatusBadRequest, "invalid_event", "name is required")
		return
	}
	if err := s.store.RecordEvent(r.Context(), req.Name, req.Properties); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to record event")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true})
}

func mapStoreErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid_input", "request is invalid")
	case errors.Is(err, store.ErrOutOfStock):
		writeError(w, http.StatusConflict, "out_of_stock", "requested quantity exceeds stock")
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "resource not found")
	case errors.Is(err, store.ErrAlreadyPicked):
		writeError(w, http.StatusConflict, "already_picked", "order already picked up")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": message,
			"details": map[string]any{},
		},
	})
}
