package domain

import "time"

type Offer struct {
	ID          string    `json:"id"`
	Merchant    string    `json:"merchant"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	PriceCents  int       `json:"priceCents"`
	Stock       int       `json:"stock"`
	PickupStart time.Time `json:"pickupStart"`
	PickupEnd   time.Time `json:"pickupEnd"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Order struct {
	ID           string    `json:"id"`
	OfferID      string    `json:"offerId"`
	CustomerName string    `json:"customerName"`
	CustomerEmail string   `json:"customerEmail"`
	Quantity     int       `json:"quantity"`
	TotalCents   int       `json:"totalCents"`
	Status       string    `json:"status"`
	PickupCode   string    `json:"pickupCode"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
