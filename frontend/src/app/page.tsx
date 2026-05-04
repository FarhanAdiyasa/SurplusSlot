"use client";

import { FormEvent, useEffect, useState, useCallback } from "react";
import { api, Offer, Order } from "@/lib/api";
import styles from "./page.module.css";

type Notification = {
  text: string;
  type: "success" | "error" | "info";
};

type OfferForm = {
  merchant: string;
  title: string;
  description: string;
  priceCents: number;
  stock: number;
  pickupStart: string;
  pickupEnd: string;
};

const initialForm: OfferForm = {
  merchant: "",
  title: "",
  description: "",
  priceCents: 500,
  stock: 10,
  pickupStart: "",
  pickupEnd: "",
};

export default function Home() {
  const [offers, setOffers] = useState<Offer[]>([]);
  const [offerForm, setOfferForm] = useState<OfferForm>(initialForm);
  const [loading, setLoading] = useState(false);
  const [notification, setNotification] = useState<Notification | null>(null);
  const [orderResult, setOrderResult] = useState<Order | null>(null);
  const [mounted, setMounted] = useState(false);
  
  // Modal State
  const [confirmingOffer, setConfirmingOffer] = useState<Offer | null>(null);
  const [customerName, setCustomerName] = useState("");

  const notify = useCallback((text: string, type: Notification["type"] = "info") => {
    setNotification({ text, type });
  }, []);

  const loadOffers = async () => {
    setLoading(true);
    try {
      const result = await api.listOffers();
      setOffers(result.offers);
    } catch (err: any) {
      notify(`Failed to load: ${err.message}`, "error");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    setMounted(true);
    let active = true;
    const run = async () => {
      setLoading(true);
      try {
        const result = await api.listOffers();
        if (active) {
          setOffers(result.offers);
        }
      } catch (err: any) {
        if (active) {
          notify(`System offline: ${err.message || "Unknown error"}`, "error");
        }
      } finally {
        if (active) {
          setLoading(false);
        }
      }
    };
    void run();
    return () => {
      active = false;
    };
  }, [notify]);

  // Auto-dismiss notification
  useEffect(() => {
    if (notification) {
      const timer = setTimeout(() => setNotification(null), 5000);
      return () => clearTimeout(timer);
    }
  }, [notification]);

  const submitOffer = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setNotification(null);
    try {
      await api.createOffer({
        ...offerForm,
        pickupStart: new Date(offerForm.pickupStart).toISOString(),
        pickupEnd: new Date(offerForm.pickupEnd).toISOString(),
      });
      notify("Offer published successfully", "success");
      setOfferForm(initialForm);
      await loadOffers();
    } catch {
      notify("Failed to publish offer", "error");
    }
  };

  const startReservation = (offer: Offer) => {
    setConfirmingOffer(offer);
    setCustomerName("");
  };

  const confirmOrder = async () => {
    if (!confirmingOffer) return;
    
    setNotification(null);
    setOrderResult(null);
    const offerToBuy = confirmingOffer;
    setConfirmingOffer(null);

    try {
      const result = await api.createOrder({
        offerId: offerToBuy.id,
        customerName: customerName || "Guest User",
        customerEmail: "customer@example.com",
        quantity: 1,
      });
      setOrderResult(result.order);
      notify("Box reserved! Check your pickup code", "success");
      await loadOffers();
    } catch {
      notify("Reservation failed. Try again", "error");
    }
  };

  const confirmPickup = async () => {
    if (!orderResult) return;
    try {
      const result = await api.confirmPickup(orderResult.id);
      setOrderResult(result.order);
      notify("Pickup confirmed. Enjoy!", "success");
    } catch {
      notify("Pickup confirmation failed", "error");
    }
  };

  const getToastClass = () => {
    if (!notification) return "";
    switch (notification.type) {
      case "success": return styles.toastSuccess;
      case "error": return styles.toastError;
      default: return styles.toastInfo;
    }
  };

  return (
    <main className={styles.page}>
      <header className={styles.header}>
        <div>
          <h1 className={styles.heroTitle}>Surplus<br />Slot</h1>
          <p className={styles.heroSubtitle}>Rescue Boxes • Timed Pickup • Zero Waste</p>
        </div>
        <div className={styles.statsRow}>
          <div className={styles.statCard}>
            <span className={styles.statLabel}>Active Boxes</span>
            <span className={styles.statValue}>{offers?.length || 0}</span>
          </div>
          <div className={styles.statCard}>
            <span className={styles.statLabel}>Status</span>
            <span className={styles.statValue}>{loading ? "..." : "Online"}</span>
          </div>
        </div>
      </header>

      <div className={styles.grid}>
        <aside className={styles.panel}>
          <h2 className={styles.panelTitle}>Publish</h2>
          <form onSubmit={submitOffer} className={styles.form}>
            <div className={`${styles.fieldGroup} ${styles.fullWidth}`}>
              <label className={styles.label}>Merchant</label>
              <input className={styles.input} required placeholder="Cafe Name" value={offerForm.merchant} onChange={(e) => setOfferForm({ ...offerForm, merchant: e.target.value })} />
            </div>
            <div className={`${styles.fieldGroup} ${styles.fullWidth}`}>
              <label className={styles.label}>Title</label>
              <input className={styles.input} required placeholder="e.g. Pastry Box" value={offerForm.title} onChange={(e) => setOfferForm({ ...offerForm, title: e.target.value })} />
            </div>
            <div className={`${styles.fieldGroup} ${styles.fullWidth}`}>
              <label className={styles.label}>Description</label>
              <textarea className={styles.textarea} placeholder="What's inside?" value={offerForm.description} onChange={(e) => setOfferForm({ ...offerForm, description: e.target.value })} />
            </div>
            <div className={styles.fieldGroup}>
              <label className={styles.label}>Price ($)</label>
              <input className={styles.input} type="number" min={0} step={0.01} required value={offerForm.priceCents / 100} onChange={(e) => setOfferForm({ ...offerForm, priceCents: Math.round(Number(e.target.value) * 100) })} />
            </div>
            <div className={styles.fieldGroup}>
              <label className={styles.label}>Stock</label>
              <input className={styles.input} type="number" min={1} required value={offerForm.stock} onChange={(e) => setOfferForm({ ...offerForm, stock: Number(e.target.value) })} />
            </div>
            <div className={styles.fieldGroup}>
              <label className={styles.label}>Pickup Start</label>
              <input className={styles.input} type="datetime-local" required value={offerForm.pickupStart} onChange={(e) => setOfferForm({ ...offerForm, pickupStart: e.target.value })} />
            </div>
            <div className={styles.fieldGroup}>
              <label className={styles.label}>Pickup End</label>
              <input className={styles.input} type="datetime-local" required value={offerForm.pickupEnd} onChange={(e) => setOfferForm({ ...offerForm, pickupEnd: e.target.value })} />
            </div>
            <button className={`${styles.button} ${styles.fullWidth}`} type="submit">Create Offer</button>
          </form>
        </aside>

        <section className={styles.panel}>
          <h2 className={styles.panelTitle}>Available</h2>
          <div className={styles.offerList}>
            {offers?.map((offer) => (
              <article key={offer.id} className={styles.offerCard}>
                <div className={styles.offerHeader}>
                  <div>
                    <h3 className={styles.offerTitle}>{offer.title}</h3>
                    <p className={styles.offerMerchant}>{offer.merchant}</p>
                  </div>
                  <div className={styles.offerPrice}>${(offer.priceCents / 100).toFixed(2)}</div>
                </div>
                <p className={styles.offerDescription}>{offer.description}</p>
                <div className={styles.offerMeta}>
                  <span>Stock: {offer.stock}</span>
                  <span>Until: {mounted ? new Date(offer.pickupEnd).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : "--:--"}</span>
                </div>
                <button className={styles.button} onClick={() => startReservation(offer)} disabled={offer.stock < 1}>
                  Reserve Now
                </button>
              </article>
            ))}
            {!loading && (offers?.length === 0 || !offers) && (
              <p style={{ gridColumn: '1/-1', textAlign: 'center', opacity: 0.5, padding: '40px' }}>No active boxes available right now.</p>
            )}
          </div>
        </section>
      </div>

      {orderResult && (
        <section className={`${styles.panel} ${styles.reservation}`}>
          <div className={styles.reservationGroup}>
            <label className={styles.label}>Order ID</label>
            <p className={styles.orderId}>{mounted ? orderResult.id : "..."}</p>
          </div>
          <div className={styles.reservationGroup}>
            <label className={styles.label}>Pickup Code</label>
            <p className={styles.pickupCode}>{mounted ? orderResult.pickupCode : "..."}</p>
          </div>
          <div className={styles.reservationGroup}>
            <label className={styles.label}>Status</label>
            <p className={styles.orderStatus}>{mounted ? orderResult.status : "..."}</p>
          </div>
          <div style={{ display: 'flex', alignItems: 'flex-end' }}>
            {orderResult.status !== "picked_up" && (
              <button className={`${styles.button} ${styles.buttonSecondary}`} onClick={confirmPickup} style={{ width: '100%' }}>
                Confirm Pickup
              </button>
            )}
          </div>
        </section>
      )}

      {/* Confirmation Modal */}
      {confirmingOffer && (
        <div className={styles.modalOverlay}>
          <div className={styles.modal}>
            <h2 className={styles.panelTitle}>Confirm Reservation</h2>
            <p>You are about to reserve <strong>{confirmingOffer.title}</strong> from <strong>{confirmingOffer.merchant}</strong> for <strong>${(confirmingOffer.priceCents/100).toFixed(2)}</strong>.</p>
            
            <div className={styles.fieldGroup}>
              <label className={styles.label}>Your Name</label>
              <input 
                className={styles.input} 
                placeholder="Enter your name" 
                value={customerName} 
                onChange={(e) => setCustomerName(e.target.value)}
                autoFocus
              />
            </div>

            <div className={styles.modalActions}>
              <button className={`${styles.button} ${styles.buttonSecondary}`} onClick={() => setConfirmingOffer(null)}>
                Cancel
              </button>
              <button className={styles.button} onClick={confirmOrder}>
                Confirm & Pay
              </button>
            </div>
          </div>
        </div>
      )}

      {notification && (
        <div className={`${styles.toast} ${getToastClass()}`}>
          <span>
            {notification.type === "success" && "✓ "}
            {notification.type === "error" && "⚠ "}
            {notification.type === "info" && "ℹ "}
          </span>
          {notification.text}
        </div>
      )}
    </main>
  );
}
