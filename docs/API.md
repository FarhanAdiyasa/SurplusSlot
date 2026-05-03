# API Documentation

Base URL: `http://localhost:8080`

## Error Schema
```json
{
  "error": {
    "code": "string",
    "message": "human readable",
    "details": {}
  }
}
```

## `GET /health`
- 200: `{ "status": "ok" }`

## `GET /api/offers`
- 200:
```json
{
  "offers": [
    {
      "id": "offer-seed-1",
      "merchant": "Sunrise Bakery",
      "title": "Mixed Pastry Rescue Box",
      "description": "Assorted pastries packed at closing.",
      "priceCents": 650,
      "stock": 10,
      "pickupStart": "2026-04-23T16:00:00Z",
      "pickupEnd": "2026-04-23T19:00:00Z",
      "status": "active"
    }
  ]
}
```

## `POST /api/offers`
Create merchant offer.

Request:
```json
{
  "merchant": "Bean & Leaf Cafe",
  "title": "Sandwich Rescue",
  "description": "Unsold sandwiches",
  "priceCents": 850,
  "stock": 6,
  "pickupStart": "2026-04-23T15:00:00Z",
  "pickupEnd": "2026-04-23T18:00:00Z"
}
```

- 201: `{ "offer": { ... } }`
- 400: invalid input

## `POST /api/orders`
Reserve one or more offer units.

Request:
```json
{
  "offerId": "offer-seed-1",
  "customerName": "Ada",
  "customerEmail": "ada@example.com",
  "quantity": 1
}
```

- 201: `{ "order": { ... } }`
- 404: offer not found
- 409: out of stock

## `POST /api/orders/:id/pickup`
Confirm order picked up.

- 200: `{ "order": { ... "status": "picked_up" } }`
- 404 / 409 for invalid states.

## `POST /api/events`
Manual event ingestion.

Request:
```json
{
  "name": "page_viewed",
  "properties": { "page": "home" }
}
```
- 201: `{ "ok": true }`
