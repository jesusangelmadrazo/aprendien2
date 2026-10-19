import { useEffect, useState } from "react";
import { useLessonStore } from "../../store/lessonStore";
import { respondToOffer } from "../../ws/useWebSocket";

export default function IncomingOfferModal() {
  const pendingOffer = useLessonStore((s) => s.pendingOffer);
  const clearPendingOffer = useLessonStore((s) => s.clearPendingOffer);
  const [secondsLeft, setSecondsLeft] = useState(0);

  useEffect(() => {
    if (!pendingOffer) return;
    setSecondsLeft(pendingOffer.expiresInS);
    const interval = setInterval(() => {
      setSecondsLeft((s) => Math.max(0, s - 1));
    }, 1000);
    return () => clearInterval(interval);
  }, [pendingOffer?.offerId]);

  if (!pendingOffer) return null;

  function respond(accepted: boolean) {
    if (!pendingOffer) return;
    respondToOffer(pendingOffer.offerId, accepted);
    clearPendingOffer();
  }

  return (
    <div className="modal-backdrop">
      <div className="modal-card">
        <h2>Nueva solicitud de clase</h2>
        <p>{pendingOffer.pickupAddress || "Punto de encuentro en el mapa"}</p>
        <p>Distancia: {pendingOffer.distanceKm.toFixed(1)} km</p>
        <p className="countdown">{secondsLeft}s para responder</p>
        <div className="modal-actions">
          <button onClick={() => respond(false)} className="btn-secondary">
            Rechazar
          </button>
          <button onClick={() => respond(true)}>Aceptar</button>
        </div>
      </div>
    </div>
  );
}
