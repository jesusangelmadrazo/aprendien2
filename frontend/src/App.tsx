import { Navigate, Route, Routes, Link } from "react-router-dom";
import { useAuthStore } from "./store/authStore";
import { useRealtimeConnection } from "./ws/useWebSocket";
import LoginPage from "./features/auth/LoginPage";
import RegisterPage from "./features/auth/RegisterPage";
import StudentHome from "./features/student/StudentHome";
import StudentLayout from "./features/student/StudentLayout";
import ReservationsPage from "./features/student/ReservationsPage";
import LessonHistoryPage from "./features/student/LessonHistoryPage";
import PaymentsPage from "./features/student/PaymentsPage";
import ExamsPage from "./features/student/ExamsPage";
import InstructorProfilePage from "./features/instructor/ProfilePage";
import InstructorLayout from "./features/instructor/InstructorLayout";
import InstructorReservationsPage from "./features/instructor/InstructorReservationsPage";
import InstructorLessonsGivenPage from "./features/instructor/InstructorLessonsGivenPage";
import WalletPage from "./features/instructor/WalletPage";
import PayoutMethodPage from "./features/instructor/PayoutMethodPage";
import DocumentsPage from "./features/instructor/DocumentsPage";
import IncomingOfferModal from "./features/instructor/IncomingOfferModal";
import AdminLayout from "./features/admin/AdminLayout";
import ExamsReviewPage from "./features/admin/ExamsReviewPage";
import PayoutsPage from "./features/admin/PayoutsPage";
import DocumentsReviewPage from "./features/admin/DocumentsReviewPage";
import type { Role } from "./types";

function ProtectedRoute({ children, role }: { children: React.ReactNode; role?: Role }) {
  const { token, user } = useAuthStore();
  if (!token) return <Navigate to="/login" replace />;
  if (role && user?.role !== role) return <Navigate to="/" replace />;
  return <>{children}</>;
}

function Header() {
  const { user, logout } = useAuthStore();
  return (
    <header className="app-header">
      <Link to="/" className="brand">Clase Drive</Link>
      {user && (
        <nav>
          <span>{user.full_name}</span>
          <button onClick={logout}>Salir</button>
        </nav>
      )}
    </header>
  );
}

function Home() {
  const user = useAuthStore((s) => s.user);
  if (!user) return <Navigate to="/login" replace />;
  if (user.role === "instructor") return <Navigate to="/instructor" replace />;
  if (user.role === "admin") return <Navigate to="/admin" replace />;
  return <Navigate to="/student" replace />;
}

export default function App() {
  useRealtimeConnection();
  const user = useAuthStore((s) => s.user);

  return (
    <>
      <Header />
      <main>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route
            path="/student"
            element={
              <ProtectedRoute role="student">
                <StudentLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<StudentHome />} />
            <Route path="reservas" element={<ReservationsPage />} />
            <Route path="clases" element={<LessonHistoryPage />} />
            <Route path="pagos" element={<PaymentsPage />} />
            <Route path="examenes" element={<ExamsPage />} />
          </Route>
          <Route
            path="/instructor"
            element={
              <ProtectedRoute role="instructor">
                <InstructorLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<InstructorProfilePage />} />
            <Route path="documentos" element={<DocumentsPage />} />
            <Route path="reservas" element={<InstructorReservationsPage />} />
            <Route path="clases" element={<InstructorLessonsGivenPage />} />
            <Route path="wallet" element={<WalletPage />} />
            <Route path="cobro" element={<PayoutMethodPage />} />
          </Route>
          <Route
            path="/admin"
            element={
              <ProtectedRoute role="admin">
                <AdminLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<ExamsReviewPage />} />
            <Route path="documentos" element={<DocumentsReviewPage />} />
            <Route path="payouts" element={<PayoutsPage />} />
          </Route>
        </Routes>
      </main>
      {user?.role === "instructor" && <IncomingOfferModal />}
    </>
  );
}
