export type Role = "student" | "instructor" | "admin";

export interface User {
  id: string;
  email: string;
  full_name: string;
  phone: string;
  role: Role;
}

export type InstructorVerificationStatus = "pending" | "approved";

export interface InstructorProfile {
  UserID: string;
  VehicleMake: string;
  VehicleModel: string;
  VehiclePlate: string;
  VehicleYear: number | null;
  HourlyRateUYU: number;
  Bio: string;
  Status: "offline" | "online" | "busy";
  VerificationStatus: InstructorVerificationStatus;
  AvgRating: number;
  TotalLessons: number;
  LastKnownLat: number | null;
  LastKnownLon: number | null;
  CreatedAt: string;
}

export interface InstructorSummary extends InstructorProfile {
  FullName: string;
}

export interface NearbyInstructor extends InstructorProfile {
  FullName: string;
  DistanceKm: number;
  Lat: number;
  Lon: number;
}

export type LessonStatus =
  | "requested"
  | "searching"
  | "accepted"
  | "instructor_enroute"
  | "in_progress"
  | "completed"
  | "cancelled"
  | "no_instructor_found";

export interface AvailabilityBlock {
  ID: string;
  InstructorID: string;
  DayOfWeek: number; // 0 = Sunday .. 6 = Saturday
  StartTime: string; // "HH:MM"
  EndTime: string;
  CreatedAt: string;
}

export interface AvailabilitySlot {
  instructor_id: string;
  start: string; // RFC3339
  end: string;
}

export interface SlotWithInstructor extends AvailabilitySlot {
  instructor_name: string;
  vehicle_make: string;
  vehicle_model: string;
  hourly_rate_uyu: number;
}

export interface LessonRequest {
  ID: string;
  StudentID: string;
  InstructorID: string | null;
  Status: LessonStatus;
  PickupLat: number;
  PickupLon: number;
  PickupAddressText: string;
  DurationMinutes: number;
  EstimatedPriceUYU: number;
  FinalPriceUYU: number | null;
  ScheduledAt: string | null;
  RequestedAt: string;
}

export type PaymentStatus = "pending" | "processing" | "approved" | "rejected" | "refunded";

export interface Payment {
  ID: string;
  LessonRequestID: string;
  StudentID: string;
  AmountUYU: number;
  InstructorPriceUYU: number;
  PlatformFeeUYU: number;
  ProviderRef: string;
  ProviderPaymentID: string;
  Status: PaymentStatus;
  CreatedAt: string;
  UpdatedAt: string;
}

export interface BookingResult {
  lesson: LessonRequest;
  payment: Payment;
}

export type ExamType = "theory" | "practice";
export type ExamSubmissionStatus = "pending_review" | "approved" | "rejected";

export interface ExamSubmission {
  ID: string;
  StudentID: string;
  Type: ExamType;
  Status: ExamSubmissionStatus;
  PhotoURL: string;
  RejectionReason: string;
  ReviewedBy: string | null;
  SubmittedAt: string;
  ReviewedAt: string | null;
}

export interface ExamTypeStatus {
  type: ExamType;
  unlocked: boolean;
  progress: number;
  threshold: number;
  submission: ExamSubmission | null;
}

export interface ExamStatusOverview {
  theory: ExamTypeStatus;
  practice: ExamTypeStatus;
}

export interface PendingExam extends ExamSubmission {
  student_name: string;
}

export type AccountType = "caja_ahorro" | "cuenta_corriente";

export interface PayoutMethod {
  InstructorID: string;
  BankName: string;
  AccountType: AccountType;
  AccountNumber: string;
  AccountHolderName: string;
  AccountHolderDocument: string;
  UpdatedAt: string;
}

export interface Payout {
  ID: string;
  InstructorID: string;
  AmountUYU: number;
  PeriodStart: string;
  PeriodEnd: string;
  CreatedBy: string;
  CreatedAt: string;
}

export interface WalletSummary {
  pending_uyu: number;
  total_paid_uyu: number;
  payouts: Payout[] | null;
}

export interface InstructorBalance {
  instructor_id: string;
  instructor_name: string;
  amount_uyu: number;
}

export type DocumentType = "license" | "identity" | "vehicle_registration" | "insurance";
export type DocumentSubmissionStatus = "pending_review" | "approved" | "rejected";

export interface InstructorDocument {
  ID: string;
  InstructorID: string;
  Type: DocumentType;
  PhotoURL: string;
  ExpiryDate: string | null;
  Status: DocumentSubmissionStatus;
  RejectionReason: string;
  ReviewedBy: string | null;
  SubmittedAt: string;
  ReviewedAt: string | null;
}

export interface DocumentStatusOverview {
  verified: boolean;
  documents: Record<DocumentType, InstructorDocument | null>;
}

export interface PendingDocument extends InstructorDocument {
  instructor_name: string;
}

export interface ScheduleOverride {
  ID: string;
  InstructorID: string;
  Date: string; // RFC3339 (time component is always midnight)
  IsDayOff: boolean;
  StartTime: string; // "HH:MM", empty when IsDayOff
  EndTime: string;
  CreatedAt: string;
}
