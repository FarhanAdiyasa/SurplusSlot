"use client";

import { useEffect, useState, useCallback } from "react";
import { api, Offer, Order } from "@/lib/api";
import styles from "./page.module.css";
import Link from "next/link";
import { motion, AnimatePresence } from "framer-motion";

type Notification = {
  text: string;
  type: "success" | "error" | "info";
};

function CountdownTimer({ targetDate }: { targetDate: string }) {
  const [timeLeft, setTimeLeft] = useState("");
  const [isUrgent, setIsUrgent] = useState(false);

  useEffect(() => {
    const calculate = () => {
      const difference = new Date(targetDate).getTime() - new Date().getTime();
      if (difference <= 0) {
        setTimeLeft("EXPIRED");
        setIsUrgent(true);
        return;
      }
      const hours = Math.floor((difference / (1000 * 60 * 60)) % 24);
      const minutes = Math.floor((difference / 1000 / 60) % 60);
      const seconds = Math.floor((difference / 1000) % 60);
      setIsUrgent(difference < 30 * 60 * 1000);
      setTimeLeft(`${hours > 0 ? hours + "h " : ""}${minutes}m ${seconds}s`);
    };
    calculate();
    const timer = setInterval(calculate, 1000);
    return () => clearInterval(timer);
  }, [targetDate]);

  return (
    <span style={{ color: isUrgent ? "#ef4444" : "inherit", fontWeight: 900, fontVariantNumeric: 'tabular-nums' }}>
      {timeLeft}
    </span>
  );
}

function OfferSkeleton() {
  return (
    <div className={`${styles.offerCard} ${styles.skeleton}`} style={{ height: '240px' }}>
      <div style={{ height: '30px', width: '60%', background: '#ccc', marginBottom: '10px' }} />
      <div style={{ height: '20px', width: '40%', background: '#ccc', marginBottom: '20px' }} />
      <div style={{ height: '60px', width: '100%', background: '#ccc', marginBottom: '20px' }} />
      <div style={{ height: '40px', width: '100%', background: '#ccc' }} />
    </div>
  );
}

export default function Home() {
  const [offers, setOffers] = useState<Offer[]>([]);
  const [loading, setLoading] = useState(true);
  const [notification, setNotification] = useState<Notification | null>(null);
  const [orderResult, setOrderResult] = useState<Order | null>(null);
  const [mounted, setMounted] = useState(false);
  
  const [confirmingOffer, setConfirmingOffer] = useState<Offer | null>(null);
  const [customerName, setCustomerName] = useState("");
  const [quantity, setQuantity] = useState(1);

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
      // Small delay for smooth transition from skeleton
      setTimeout(() => setLoading(false), 800);
    }
  };

  useEffect(() => {
    setMounted(true);
    loadOffers();
  }, [notify]);

  useEffect(() => {
    if (notification) {
      const timer = setTimeout(() => setNotification(null), 5000);
      return () => clearTimeout(timer);
    }
  }, [notification]);

  const startReservation = (offer: Offer) => {
    setConfirmingOffer(offer);
    setCustomerName("");
    setQuantity(1);
  };

  const confirmOrder = async () => {
    if (!confirmingOffer) return;
    setOrderResult(null);
    const offerToBuy = confirmingOffer;
    const buyQty = quantity;
    setConfirmingOffer(null);

    try {
      const result = await api.createOrder({
        offerId: offerToBuy.id,
        customerName: customerName || "Guest User",
        customerEmail: "customer@example.com",
        quantity: buyQty,
      });
      setOrderResult(result.order);
      notify(`Reserved ${buyQty} box(es)! Show your code at pickup.`, "success");
      await loadOffers();
    } catch {
      notify("Reservation failed. Out of stock?", "error");
      await loadOffers();
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

  return (
    <motion.main 
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      className={styles.page}
    >
      <header className={styles.header}>
        <motion.div initial={{ x: -20 }} animate={{ x: 0 }}>
          <h1 className={styles.heroTitle}>Rescue<br />Boxes</h1>
          <p className={styles.heroSubtitle}>Find end-of-day surplus near you.</p>
        </motion.div>
        <div className={styles.statsRow}>
          <div className={styles.statCard}>
            <span className={styles.statLabel}>Available</span>
            <span className={styles.statValue}>{loading ? "..." : offers?.length}</span>
          </div>
          <Link href="/merchant" className={styles.button} style={{ width: 'auto', background: '#fff', color: 'var(--accent)' }}>
            I'm a Merchant →
          </Link>
        </div>
      </header>

      <section className={styles.panel}>
        <div className={styles.offerList} style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))' }}>
          {loading ? (
            <>
              <OfferSkeleton />
              <OfferSkeleton />
              <OfferSkeleton />
            </>
          ) : (
            <AnimatePresence mode="popLayout">
              {offers?.map((offer) => (
                <motion.article 
                  key={offer.id} 
                  layout
                  initial={{ opacity: 0, scale: 0.9 }}
                  animate={{ opacity: 1, scale: 1 }}
                  exit={{ opacity: 0, scale: 0.8 }}
                  className={styles.offerCard}
                >
                  <div className={styles.offerHeader}>
                    <div>
                      <h3 className={styles.offerTitle}>{offer.title}</h3>
                      <p className={styles.offerMerchant}>{offer.merchant}</p>
                    </div>
                    <div className={styles.offerPrice}>${(offer.priceCents / 100).toFixed(2)}</div>
                  </div>
                  <p className={styles.offerDescription}>{offer.description}</p>
                  <div className={styles.offerMeta}>
                    <div style={{ display: 'flex', flexDirection: 'column' }}>
                      <span className={styles.label}>Stock</span>
                      <span style={{ fontWeight: 900, fontSize: '1.2rem' }}>{offer.stock}</span>
                    </div>
                    <div style={{ display: 'flex', flexDirection: 'column', textAlign: 'right' }}>
                      <span className={styles.label}>Ends In</span>
                      <CountdownTimer targetDate={offer.pickupEnd} />
                    </div>
                  </div>
                  <button className={styles.button} onClick={() => startReservation(offer)} disabled={offer.stock < 1}>
                    Reserve Now
                  </button>
                </motion.article>
              ))}
            </AnimatePresence>
          )}
          {!loading && offers?.length === 0 && (
            <p style={{ gridColumn: '1/-1', textAlign: 'center', opacity: 0.5, padding: '80px' }}>No active boxes available right now. Check back later!</p>
          )}
        </div>
      </section>

      {orderResult && (
        <motion.section 
          initial={{ scale: 0.8, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          className={`${styles.panel} ${styles.reservation}`}
        >
          <h2 className={styles.panelTitle}>Active Reservation</h2>
          <div className={styles.reservationGroup}>
            <label className={styles.label}>Pickup Code</label>
            <p className={styles.pickupCode} style={{ fontSize: '2.5rem', color: 'var(--accent)' }}>{mounted ? orderResult.pickupCode : "..."}</p>
          </div>
          <div className={styles.reservationGroup}>
            <label className={styles.label}>Qty</label>
            <p className={styles.orderId}>{orderResult.quantity} Box(es)</p>
          </div>
          <div className={styles.reservationGroup}>
            <label className={styles.label}>Status</label>
            <p className={styles.orderStatus}>{mounted ? orderResult.status : "..."}</p>
          </div>
          <div style={{ display: 'flex', alignItems: 'flex-end', gridColumn: '1/-1' }}>
            {orderResult.status !== "picked_up" && (
              <button className={`${styles.button} ${styles.buttonSecondary}`} onClick={confirmPickup} style={{ width: '100%' }}>
                Confirm Pickup
              </button>
            )}
          </div>
        </motion.section>
      )}

      <AnimatePresence>
        {confirmingOffer && (
          <motion.div 
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className={styles.modalOverlay}
          >
            <motion.div 
              initial={{ scale: 0.5 }}
              animate={{ scale: 1 }}
              exit={{ scale: 0.5 }}
              className={styles.modal}
            >
              <h2 className={styles.panelTitle}>Confirm Order</h2>
              <p>Reserving <strong>{confirmingOffer.title}</strong> from <strong>{confirmingOffer.merchant}</strong>.</p>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '20px' }}>
                <div className={styles.fieldGroup}>
                  <label className={styles.label}>Name</label>
                  <input className={styles.input} placeholder="Your name" value={customerName} onChange={(e) => setCustomerName(e.target.value)} autoFocus />
                </div>
                <div className={styles.fieldGroup}>
                  <label className={styles.label}>Qty</label>
                  <input className={styles.input} type="number" min={1} max={confirmingOffer.stock} value={quantity} onChange={(e) => setQuantity(Math.min(confirmingOffer.stock, Math.max(1, Number(e.target.value))))} />
                </div>
              </div>
              <div style={{ background: '#f0f0f0', padding: '12px', border: 'var(--border-thin)', textAlign: 'right' }}>
                <span className={styles.label}>Total: </span>
                <span style={{ fontSize: '1.5rem', fontWeight: 900 }}>${((confirmingOffer.priceCents * quantity) / 100).toFixed(2)}</span>
              </div>
              <div className={styles.modalActions}>
                <button className={`${styles.button} ${styles.buttonSecondary}`} onClick={() => setConfirmingOffer(null)}>Cancel</button>
                <button className={styles.button} onClick={confirmOrder}>Reserve</button>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

      <AnimatePresence>
        {notification && (
          <motion.div 
            initial={{ x: 100, opacity: 0 }}
            animate={{ x: 0, opacity: 1 }}
            exit={{ x: 100, opacity: 0 }}
            className={`${styles.toast} ${notification.type === "success" ? styles.toastSuccess : styles.toastError}`}
          >
            {notification.text}
          </motion.div>
        )}
      </AnimatePresence>
    </motion.main>
  );
}
