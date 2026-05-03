INSERT INTO offers (id, merchant, title, description, price_cents, stock, pickup_start, pickup_end, status)
VALUES
  ('offer-seed-1', 'Sunrise Bakery', 'Mixed Pastry Rescue Box', 'Assorted pastries packed at closing.', 650, 10, NOW() + INTERVAL '1 hour', NOW() + INTERVAL '4 hour', 'active'),
  ('offer-seed-2', 'Bean & Leaf Cafe', 'Sandwich + Drink Surprise Bag', 'End-of-day sandwiches and bottled drinks.', 850, 6, NOW() + INTERVAL '2 hour', NOW() + INTERVAL '5 hour', 'active')
ON CONFLICT (id) DO NOTHING;
