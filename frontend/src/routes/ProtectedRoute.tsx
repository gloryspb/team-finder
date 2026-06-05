import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "../api/auth";

export function ProtectedRoute() {
  const { loading, isAuthenticated } = useAuth();
  if (loading) {
    return <div className="empty">Загрузка...</div>;
  }
  return isAuthenticated ? <Outlet /> : <Navigate to="/login" replace />;
}
