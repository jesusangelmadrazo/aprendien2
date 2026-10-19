import { useEffect, useMemo, useState } from "react";
import { getSlotsForDate, bookSlot } from "../../api/availability";
import type { InstructorSummary, SlotWithInstructor } from "../../types";

interface Props {
  instructors: InstructorSummary[];
  pickup: { lat: number; lon: number } | null;
}

const WEEKDAY_SHORT = ["Dom", "Lun", "Mar", "Mié", "Jue", "Vie", "Sáb"];
const DAYS_AHEAD = 6;
const TIME_FORMATTER = new Intl.DateTimeFormat("es-UY", { hour: "2-digit", minute: "2-digit" });
const AVATAR_COLORS = ["var(--color-avatar-1)", "var(--color-avatar-2)", "var(--color-avatar-3)"];
// Preview only — the backend computes the real fee from PLATFORM_FEE_PERCENT
// and is the source of truth for what actually gets charged.
const FEE_PREVIEW_PERCENT = 10;

function isoDate(d: Date) {
  const offsetMs = d.getTimezoneOffset() * 60000;
  return new Date(d.getTime() - offsetMs).toISOString().slice(0, 10);
}

function dateForOffset(offset: number) {
  const d = new Date();
  d.setDate(d.getDate() + offset);
  return d;
}

function initialsOf(name: string) {
  return name
    .split(" ")
    .map((p) => p[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();
}

function avatarColorFor(name: string) {
  let hash = 0;
  for (let i = 0; i < name.length; i++) hash = (hash * 31 + name.charCodeAt(i)) >>> 0;
  return AVATAR_COLORS[hash % AVATAR_COLORS.length];
}

function slotKey(slot: SlotWithInstructor) {
  return slot.instructor_id + "|" + slot.start;
}

export default function SlotSearch({ instructors, pickup }: Props) {
  const [dayOffset, setDayOffset] = useState(0);
  const [instructorFilter, setInstructorFilter] = useState("");
  const [slots, setSlots] = useState<SlotWithInstructor[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [cart, setCart] = useState<SlotWithInstructor[]>([]);
  const [modalOpen, setModalOpen] = useState(false);
  const [confirming, setConfirming] = useState(false);
  const [toast, setToast] = useState<string | null>(null);

  useEffect(() => {
    if (!toast) return;
    const t = setTimeout(() => setToast(null), 3000);
    return () => clearTimeout(t);
  }, [toast]);

  useEffect(() => {
    setLoading(true);
    setError(null);
    getSlotsForDate(isoDate(dateForOffset(dayOffset)), instructorFilter || undefined)
      .then(setSlots)
      .catch(() => setError("No se pudieron cargar los horarios"))
      .finally(() => setLoading(false));
  }, [dayOffset, instructorFilter]);

  const ratingByInstructor = useMemo(() => {
    const map = new Map<string, number>();
    instructors.forEach((i) => map.set(i.UserID, i.AvgRating));
    return map;
  }, [instructors]);

  const groups = useMemo(() => {
    const byInstructor = new Map<string, SlotWithInstructor[]>();
    for (const slot of slots) {
      if (!byInstructor.has(slot.instructor_id)) byInstructor.set(slot.instructor_id, []);
      byInstructor.get(slot.instructor_id)!.push(slot);
    }
    return [...byInstructor.entries()];
  }, [slots]);

  function inCart(slot: SlotWithInstructor) {
    return cart.some((c) => slotKey(c) === slotKey(slot));
  }

  function toggleSlot(slot: SlotWithInstructor) {
    setCart((prev) =>
      inCart(slot) ? prev.filter((c) => slotKey(c) !== slotKey(slot)) : [...prev, slot]
    );
  }

  function removeFromCart(slot: SlotWithInstructor) {
    setCart((prev) => prev.filter((c) => slotKey(c) !== slotKey(slot)));
  }

  const subtotal = cart.reduce((sum, c) => sum + c.hourly_rate_uyu, 0);
  const feePreview = Math.round(subtotal * FEE_PREVIEW_PERCENT) / 100;
  const totalPreview = subtotal + feePreview;

  async function handleConfirm() {
    if (!pickup || cart.length === 0) return;
    setConfirming(true);
    const results = await Promise.allSettled(
      cart.map((slot) =>
        bookSlot(slot.instructor_id, {
          scheduled_at: slot.start,
          pickup_lat: pickup.lat,
          pickup_lon: pickup.lon,
          duration_minutes: 60,
        })
      )
    );

    const failedKeys = new Set(
      cart.filter((_, i) => results[i].status === "rejected").map((c) => slotKey(c))
    );
    const succeededCount = cart.length - failedKeys.size;
    const chargedTotal = results.reduce(
      (sum, r) => (r.status === "fulfilled" ? sum + r.value.payment.AmountUYU : sum),
      0
    );

    setCart((prev) => prev.filter((c) => failedKeys.has(slotKey(c))));
    setConfirming(false);

    if (failedKeys.size === 0) {
      setModalOpen(false);
      setToast(
        succeededCount === 1
          ? `¡Clase confirmada! Se cobraron $${chargedTotal} UYU.`
          : `¡${succeededCount} clases confirmadas! Se cobraron $${chargedTotal} UYU en total.`
      );
    } else if (succeededCount > 0) {
      setToast(`${succeededCount} confirmadas ($${chargedTotal} UYU), ${failedKeys.size} ya no estaban disponibles`);
    } else {
      setError("Esos horarios ya no están disponibles. Sacalos de la selección e intentá con otros.");
    }

    getSlotsForDate(isoDate(dateForOffset(dayOffset)), instructorFilter || undefined)
      .then(setSlots)
      .catch(() => {});
  }

  return (
    <section className="slot-search-card">
      <h2>Reservar horario fijo</h2>
      <p className="hint">Elegí el día, mirá quién tiene lugar, y sumá horas siempre con el mismo instructor.</p>

      <div className="day-strip">
        {Array.from({ length: DAYS_AHEAD }, (_, i) => i).map((offset) => {
          const d = dateForOffset(offset);
          return (
            <button
              key={offset}
              className={"day-chip" + (offset === dayOffset ? " selected" : "")}
              onClick={() => setDayOffset(offset)}
            >
              <span className="label">{offset === 0 ? "Hoy" : offset === 1 ? "Mañana" : WEEKDAY_SHORT[d.getDay()]}</span>
              <span className="num">{d.getDate()}</span>
            </button>
          );
        })}
      </div>

      <div className="filter-row">
        <label>
          <div className="field-label">Instructor</div>
          <select value={instructorFilter} onChange={(e) => setInstructorFilter(e.target.value)}>
            <option value="">Todos los instructores</option>
            {instructors.map((i) => (
              <option key={i.UserID} value={i.UserID}>
                {i.FullName}
              </option>
            ))}
          </select>
        </label>
      </div>

      {error && <p className="form-error">{error}</p>}
      {loading && <p>Buscando...</p>}
      {!loading && groups.length === 0 && <p className="hint">No hay horarios disponibles para ese filtro.</p>}

      {groups.map(([instructorId, instructorSlots]) => {
        const meta = instructorSlots[0];
        const rating = ratingByInstructor.get(instructorId) ?? 0;
        return (
          <div key={instructorId} className="instructor-group">
            <div className="instructor-header">
              <div className="avatar" style={{ background: avatarColorFor(meta.instructor_name) }}>
                {initialsOf(meta.instructor_name)}
              </div>
              <div className="instructor-meta">
                <h3>{meta.instructor_name}</h3>
                <p className="line">
                  {rating > 0 && <span className="rating">★ {rating.toFixed(1)}</span>}
                  {rating > 0 && " · "}
                  {meta.vehicle_make} {meta.vehicle_model}
                </p>
              </div>
              <div className="price-tag">${meta.hourly_rate_uyu} UYU/h</div>
            </div>
            <div className="chip-grid">
              {instructorSlots.map((slot) => (
                <button
                  key={slotKey(slot)}
                  className={"time-chip" + (inCart(slot) ? " selected" : "")}
                  onClick={() => toggleSlot(slot)}
                >
                  {TIME_FORMATTER.format(new Date(slot.start))}
                </button>
              ))}
            </div>
          </div>
        );
      })}

      {cart.length > 0 && (
        <div className="cart-bar">
          <div className="cart-bar-inner show">
            <span className="cart-bar-count">
              <b>{cart.length}</b> {cart.length === 1 ? "horario seleccionado" : "horarios seleccionados"}
            </span>
            <button onClick={() => setModalOpen(true)}>Revisar y confirmar</button>
          </div>
        </div>
      )}

      {modalOpen && (
        <div className="modal-backdrop show" onClick={() => setModalOpen(false)}>
          <div className="modal-dialog" role="dialog" aria-modal="true" onClick={(e) => e.stopPropagation()}>
            <h2>Confirmá tus clases</h2>
            <p className="modal-sub">Revisá los horarios elegidos. Podés sacar alguno antes de confirmar.</p>

            {cart.length === 0 ? (
              <p className="empty-cart-hint">Todavía no elegiste ningún horario.</p>
            ) : (
              <ul className="review-list">
                {cart.map((slot) => {
                  const d = new Date(slot.start);
                  return (
                    <li key={slotKey(slot)} className="review-row">
                      <div className="date-badge">
                        <span className="weekday">{WEEKDAY_SHORT[d.getDay()]}</span>
                        {d.getDate()}
                      </div>
                      <div className="details">
                        <div className="time">{TIME_FORMATTER.format(d)}</div>
                        <div className="who">
                          {slot.instructor_name} — {slot.vehicle_make} {slot.vehicle_model}
                        </div>
                      </div>
                      <div className="row-price">${slot.hourly_rate_uyu}</div>
                      <button className="remove-btn" aria-label="Quitar" onClick={() => removeFromCart(slot)}>
                        ×
                      </button>
                    </li>
                  );
                })}
              </ul>
            )}

            <div className="price-breakdown">
              <div className="price-breakdown-row">
                <span>Subtotal instructores</span>
                <span>${subtotal} UYU</span>
              </div>
              <div className="price-breakdown-row hint">
                <span>Fee de servicio ({FEE_PREVIEW_PERCENT}%)</span>
                <span>${feePreview} UYU</span>
              </div>
            </div>

            <div className="modal-total">
              <span>Total a pagar</span>
              <b>${totalPreview} UYU</b>
            </div>

            <div className="modal-actions">
              <button className="btn-ghost" onClick={() => setModalOpen(false)}>
                Seguir eligiendo
              </button>
              <button className="btn-confirm" disabled={confirming || cart.length === 0} onClick={handleConfirm}>
                {confirming ? "Procesando pago..." : `Confirmar y pagar $${totalPreview}`}
              </button>
            </div>
          </div>
        </div>
      )}

      {toast && <div className="toast show">{toast}</div>}
    </section>
  );
}
