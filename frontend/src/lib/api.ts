const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://127.0.0.1:8080";

export type Offer = {
  id: string;
  merchant: string;
  title: string;
  description: string;
  priceCents: number;
  stock: number;
  pickupStart: string;
  pickupEnd: string;
  status: string;
};

export type Order = {
  id: string;
  offerId: string;
  customerName: string;
  customerEmail: string;
  quantity: number;
  totalCents: number;
  status: string;
  pickupCode: string;
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: { "Content-Type": "application/json", ...(init?.headers ?? {}) },
  });
  if (!res.ok) {
    throw new Error(`request failed: ${res.status}`);
  }
  return res.json() as Promise<T>;
}

export const api = {
  listOffers: () => request<{ offers: Offer[] }>("/api/offers"),
  createOffer: (payload: {
    merchant: string;
    title: string;
    description: string;
    priceCents: number;
    stock: number;
    pickupStart: string;
    pickupEnd: string;
  }) =>
    request<{ offer: Offer }>("/api/offers", { method: "POST", body: JSON.stringify(payload) }),
  createOrder: (payload: { offerId: string; customerName: string; customerEmail: string; quantity: number }) =>
    request<{ order: Order }>("/api/orders", { method: "POST", body: JSON.stringify(payload) }),
  confirmPickup: (orderId: string) => request<{ order: Order }>(`/api/orders/${orderId}/pickup`, { method: "POST" }),
};
