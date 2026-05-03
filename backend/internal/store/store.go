package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"surplusslot/backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrOutOfStock     = errors.New("out of stock")
	ErrAlreadyPicked  = errors.New("already picked up")
)

type PostgresStore struct {
	db *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 8
	return pgxpool.NewWithConfig(ctx, cfg)
}

func NewPostgresStore(db *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) CreateOffer(ctx context.Context, in domain.Offer) (domain.Offer, error) {
	if strings.TrimSpace(in.Merchant) == "" || strings.TrimSpace(in.Title) == "" || in.PriceCents <= 0 || in.Stock <= 0 {
		return domain.Offer{}, ErrInvalidInput
	}
	if !in.PickupEnd.After(in.PickupStart) {
		return domain.Offer{}, ErrInvalidInput
	}

	id := uuid.NewString()
	var out domain.Offer
	err := s.db.QueryRow(ctx, `
		INSERT INTO offers (id, merchant, title, description, price_cents, stock, pickup_start, pickup_end, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'active')
		RETURNING id, merchant, title, description, price_cents, stock, pickup_start, pickup_end, status, created_at, updated_at
	`, id, strings.TrimSpace(in.Merchant), strings.TrimSpace(in.Title), strings.TrimSpace(in.Description), in.PriceCents, in.Stock, in.PickupStart, in.PickupEnd).
		Scan(&out.ID, &out.Merchant, &out.Title, &out.Description, &out.PriceCents, &out.Stock, &out.PickupStart, &out.PickupEnd, &out.Status, &out.CreatedAt, &out.UpdatedAt)
	return out, err
}

func (s *PostgresStore) ListActiveOffers(ctx context.Context) ([]domain.Offer, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, merchant, title, description, price_cents, stock, pickup_start, pickup_end, status, created_at, updated_at
		FROM offers
		WHERE status='active' AND pickup_end > NOW() AND stock > 0
		ORDER BY pickup_start ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	offers := make([]domain.Offer, 0)
	for rows.Next() {
		var o domain.Offer
		if err := rows.Scan(&o.ID, &o.Merchant, &o.Title, &o.Description, &o.PriceCents, &o.Stock, &o.PickupStart, &o.PickupEnd, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		offers = append(offers, o)
	}
	return offers, rows.Err()
}

func (s *PostgresStore) CreateOrder(ctx context.Context, offerID, customerName, customerEmail string, qty int) (domain.Order, error) {
	if strings.TrimSpace(offerID) == "" || strings.TrimSpace(customerName) == "" || strings.TrimSpace(customerEmail) == "" || qty <= 0 {
		return domain.Order{}, ErrInvalidInput
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return domain.Order{}, err
	}
	defer tx.Rollback(ctx)

	var stock, priceCents int
	err = tx.QueryRow(ctx, `SELECT stock, price_cents FROM offers WHERE id=$1 AND status='active' FOR UPDATE`, offerID).Scan(&stock, &priceCents)
	if err != nil {
		return domain.Order{}, ErrNotFound
	}
	if stock < qty {
		return domain.Order{}, ErrOutOfStock
	}

	_, err = tx.Exec(ctx, `UPDATE offers SET stock = stock - $1, updated_at = NOW() WHERE id=$2`, qty, offerID)
	if err != nil {
		return domain.Order{}, err
	}

	orderID := uuid.NewString()
	pickupCode := strings.ToUpper(uuid.NewString()[:6])
	total := priceCents * qty

	var out domain.Order
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (id, offer_id, customer_name, customer_email, quantity, total_cents, status, pickup_code)
		VALUES ($1,$2,$3,$4,$5,$6,'placed',$7)
		RETURNING id, offer_id, customer_name, customer_email, quantity, total_cents, status, pickup_code, created_at, updated_at
	`, orderID, offerID, strings.TrimSpace(customerName), strings.TrimSpace(customerEmail), qty, total, pickupCode).
		Scan(&out.ID, &out.OfferID, &out.CustomerName, &out.CustomerEmail, &out.Quantity, &out.TotalCents, &out.Status, &out.PickupCode, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return domain.Order{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Order{}, err
	}
	return out, nil
}

func (s *PostgresStore) ConfirmPickup(ctx context.Context, orderID string) (domain.Order, error) {
	if strings.TrimSpace(orderID) == "" {
		return domain.Order{}, ErrInvalidInput
	}

	var current string
	err := s.db.QueryRow(ctx, `SELECT status FROM orders WHERE id=$1`, orderID).Scan(&current)
	if err != nil {
		return domain.Order{}, ErrNotFound
	}
	if current == "picked_up" {
		return domain.Order{}, ErrAlreadyPicked
	}

	var out domain.Order
	err = s.db.QueryRow(ctx, `
		UPDATE orders
		SET status='picked_up', updated_at=NOW()
		WHERE id=$1
		RETURNING id, offer_id, customer_name, customer_email, quantity, total_cents, status, pickup_code, created_at, updated_at
	`, orderID).Scan(&out.ID, &out.OfferID, &out.CustomerName, &out.CustomerEmail, &out.Quantity, &out.TotalCents, &out.Status, &out.PickupCode, &out.CreatedAt, &out.UpdatedAt)
	return out, err
}

func (s *PostgresStore) RecordEvent(ctx context.Context, name string, properties map[string]any) error {
	_, err := s.db.Exec(ctx, `INSERT INTO events (id, name, properties) VALUES ($1,$2,$3)`, uuid.NewString(), name, fmt.Sprintf("%v", properties))
	return err
}

func ParseTime(v string) (time.Time, error) {
	return time.Parse(time.RFC3339, v)
}
