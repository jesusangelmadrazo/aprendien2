import { useEffect } from "react";
import { useAuthStore } from "../store/authStore";
import { useLessonStore } from "../store/lessonStore";
import * as socketManager from "./socketManager";
import { apiBaseUrl } from "../api/client";

const wsBaseUrl = apiBaseUrl.replace(/^http/, "ws");

/** Mounted once at the app root; keeps a single WS connection alive for the
 * logged-in user and routes incoming events into the lesson store. */
export function useRealtimeConnection() {
  const token = useAuthStore((s) => s.token);

  useEffect(() => {
    if (!token) {
      socketManager.disconnect();
      return;
    }

    socketManager.connect(token, wsBaseUrl);

    const unsubscribe = socketManager.subscribe((event) => {
      const store = useLessonStore.getState();
      switch (event.type) {
        case "lesson_offer":
          store.setPendingOffer(event.payload);
          break;
        case "offer_expired":
          store.clearPendingOfferIfMatches(event.payload.offer_id);
          break;
        case "instructor_assigned":
          store.setCurrentLesson({
            lessonRequestId: event.payload.lesson_request_id,
            status: "accepted",
            instructorName: event.payload.instructor_name,
            vehicle: event.payload.vehicle,
            distanceKm: event.payload.distance_km,
          });
          break;
        case "no_instructor_found":
          store.setCurrentLesson({
            lessonRequestId: event.payload.lesson_request_id,
            status: "no_instructor_found",
          });
          break;
      }
    });

    return () => {
      unsubscribe();
      socketManager.disconnect();
    };
  }, [token]);
}

export function sendLocationUpdate(lat: number, lon: number) {
  socketManager.send("location_update", { lat, lon });
}

/** Fires whenever the WS (re)connects — useful to immediately re-send state
 * (like the instructor's location) that a dropped connection may have missed. */
export function onSocketReconnect(listener: () => void) {
  return socketManager.subscribeOpen(listener);
}

export function respondToOffer(offerId: string, accepted: boolean) {
  socketManager.send("lesson_offer_response", { offer_id: offerId, accepted });
}
