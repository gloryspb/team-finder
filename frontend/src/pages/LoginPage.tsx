import { FormEvent, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../api/auth";
import { Notice } from "../components/Notice";

export function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  async function submit(event: FormEvent) {
    event.preventDefault();
    setError("");
    try {
      await login(email, password);
      navigate("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось войти");
    }
  }

  return (
    <section className="auth-panel">
      <h1>Вход</h1>
      <form onSubmit={submit} className="form">
        {error && <Notice type="error">{error}</Notice>}
        <label>Email<input value={email} onChange={(e) => setEmail(e.target.value)} type="email" required /></label>
        <label>Пароль<input value={password} onChange={(e) => setPassword(e.target.value)} type="password" required /></label>
        <button className="button" type="submit">Войти</button>
      </form>
      <p className="muted">Нет аккаунта? <Link to="/register">Зарегистрироваться</Link></p>
    </section>
  );
}
