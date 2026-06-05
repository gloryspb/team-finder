import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../api/auth";

export function Layout() {
  const { isAuthenticated, profile, logout } = useAuth();
  const navigate = useNavigate();

  function handleLogout() {
    logout();
    navigate("/login");
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <NavLink to="/" className="brand">Team Finder</NavLink>
        <nav className="nav">
          <NavLink to="/listings">Поиск команд</NavLink>
          {isAuthenticated && <NavLink to="/listings/create">Создать объявление</NavLink>}
          {isAuthenticated && <NavLink to="/profile">Профиль и заявки</NavLink>}
        </nav>
        <div className="account">
          {isAuthenticated ? (
            <>
              <span>{profile?.nickname ?? "Игрок"}</span>
              <button className="ghost" onClick={handleLogout}>Выйти</button>
            </>
          ) : (
            <>
              <NavLink to="/login">Вход</NavLink>
              <NavLink to="/register" className="button small">Регистрация</NavLink>
            </>
          )}
        </div>
      </header>
      <main className="page">
        <Outlet />
      </main>
    </div>
  );
}
