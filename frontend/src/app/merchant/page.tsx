"use client";

import { FormEvent, useState, useEffect, useCallback } from "react";
import { api, Offer } from "@/lib/api";
import styles from "../page.module.css";
import Link from "next/link";

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

export default function MerchantPage() {
  const [offerForm, setOfferForm] = useState<OfferForm>(initialForm);
  const [notification, setNotification] = useState<Notification | null>(null);
  const [myOffers, setMyOffers] = useState<Offer[]>([]);
  const [loading, setLoading] = useState(false);

  const notify = useCallback((text: string, type: Notification["type"] = "info") => {
    setNotification({ text, type });
  }, []);

  const loadMyOffers = async () => {
    try {
      const result = await api.listOffers();
      // In a real app, we would filter by the logged-in merchant's ID
      setMyOffers(result.offers);
    } catch (err: any) {
      notify(`Failed to load: ${err.message}`, "error");
    }
  };

  useEffect(() => {
    loadMyOffers();
  }, []);

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
    setLoading(true);
    try {
      await api.createOffer({
        ...offerForm,
        pickupStart: new Date(offerForm.pickupStart).toISOString(),
        pickupEnd: new Date(offerForm.pickupEnd).toISOString(),
      });
      notify("Offer published successfully", "success");
      setOfferForm(initialForm);
      await loadMyOffers();
    } catch {
      notify("Failed to publish offer", "error");
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className={styles.page}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.heroTitle}>Merchant<br />Dashboard</h1>
          <p className={styles.heroSubtitle}>Manage your surplus boxes and reduce waste.</p>
        </div>
        <Link href="/" className={styles.button} style={{ width: 'auto', padding: '10px 20px' }}>
          ← Back to Marketplace
        </Link>
      </div>

      <div className={styles.grid}>
        <aside className={styles.panel}>
          <h2 className={styles.panelTitle}>Publish New Box</h2>
          <form onSubmit={submitOffer} className={styles.form}>
            <div className={`${styles.fieldGroup} ${styles.fullWidth}`}>
              <label className={styles.label}>Merchant Name</label>
              <input className={styles.input} required placeholder="e.g. Blue Bottle Coffee" value={offerForm.merchant} onChange={(e) => setOfferForm({ ...offerForm, merchant: e.target.value })} />
            </div>
            <div className={`${styles.fieldGroup} ${styles.fullWidth}`}>
              <label className={styles.label}>Box Title</label>
              <input className={styles.input} required placeholder="e.g. Afternoon Pastry Bag" value={offerForm.title} onChange={(e) => setOfferForm({ ...offerForm, title: e.target.value })} />
            </div>
            <div className={`${styles.fieldGroup} ${styles.fullWidth}`}>
              <label className={styles.label}>Description</label>
              <textarea className={styles.textarea} placeholder="What might be inside? (e.g. 3-4 random pastries)" value={offerForm.description} onChange={(e) => setOfferForm({ ...offerForm, description: e.target.value })} />
            </div>
            <div className={styles.fieldGroup}>
              <label className={styles.label}>Price ($)</label>
              <input className={styles.input} type="number" min={0} step={0.01} required value={offerForm.priceCents / 100} onChange={(e) => setOfferForm({ ...offerForm, priceCents: Math.round(Number(e.target.value) * 100) })} />
            </div>
            <div className={styles.fieldGroup}>
              <label className={styles.label}>Stock Available</label>
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
            <button className={`${styles.button} ${styles.fullWidth}`} type="submit" disabled={loading}>
              {loading ? "Publishing..." : "Create Offer"}
            </button>
          </form>
        </aside>

        <section className={styles.panel}>
          <h2 className={styles.panelTitle}>Your Active Offers</h2>
          <div className={styles.offerList}>
            {myOffers.map((offer) => (
              <div key={offer.id} className={styles.offerCard} style={{ opacity: offer.stock === 0 ? 0.5 : 1 }}>
                <div className={styles.offerHeader}>
                  <h3 className={styles.offerTitle}>{offer.title}</h3>
                  <div className={styles.offerPrice}>${(offer.priceCents / 100).toFixed(2)}</div>
                </div>
                <p className={styles.offerDescription}>{offer.description}</p>
                <div className={styles.offerMeta}>
                  <span>Stock: <strong>{offer.stock}</strong></span>
                  <span>Ends: {new Date(offer.pickupEnd).toLocaleTimeString()}</span>
                </div>
              </div>
            ))}
            {myOffers.length === 0 && (
              <p style={{ textAlign: 'center', opacity: 0.5, padding: '40px' }}>You haven't published any boxes yet.</p>
            )}
          </div>
        </section>
      </div>

      {notification && (
        <div className={`${styles.toast} ${notification.type === "success" ? styles.toastSuccess : styles.toastError}`}>
          {notification.text}
        </div>
      )}
    </main>
  );
}
