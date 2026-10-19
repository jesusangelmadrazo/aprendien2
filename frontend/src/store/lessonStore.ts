import { create } from "zustand";

export interface PendingOffer {
  offerId: string;
  lessonRequestId: string;
  pickupAddress: string;
  distanceKm: number;
  expiresInS: number;
}

export interface CurrentLesson {
  lessonRequestId: string;
  status: "searching" | "accepted" | "no_instructor_found";
  instructorName?: string;
  vehicle?: string;
  distanceKm?: number;
}

interface LessonState {
  pendingOffer: PendingOffer | null;
  currentLesson: CurrentLesson | null;
  setPendingOffer: (payload: any) => void;
  clearPendingOfferIfMatches: (offerId: string) => void;
  clearPendingOffer: () => void;
  setCurrentLesson: (lesson: CurrentLesson) => void;
  clearCurrentLesson: () => void;
}

export const useLessonStore = create<LessonState>((set, get) => ({
  pendingOffer: null,
  currentLesson: null,
  setPendingOffer: (payload) =>
    set({
      pendingOffer: {
        offerId: payload.offer_id,
        lessonRequestId: payload.lesson_request_id,
        pickupAddress: payload.pickup_address,
        distanceKm: payload.distance_km,
        expiresInS: payload.expires_in_s,
      },
    }),
  clearPendingOfferIfMatches: (offerId) => {
    if (get().pendingOffer?.offerId === offerId) set({ pendingOffer: null });
  },
  clearPendingOffer: () => set({ pendingOffer: null }),
  setCurrentLesson: (lesson) => set({ currentLesson: lesson }),
  clearCurrentLesson: () => set({ currentLesson: null }),
}));
