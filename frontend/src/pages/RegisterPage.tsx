import { FormEvent, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../api/auth";
import { Notice } from "../components/Notice";

export function RegisterPage() {
  const { register } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [nickname, setNickname] = useState("");
  const [error, setError] = useState("");

  async function submit(event: FormEvent) {
    event.preventDefault();
    setError("");
    try {
      await register(email, password, nickname);
      navigate("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось зарегистрироваться");
    }
  }

  return (
    <section className="auth-panel">
      <h1>Регистрация</h1>
      <form onSubmit={submit} className="form">
        {error && <Notice type="error">{error}</Notice>}
        <label>Никнейм<input value={nickname} onChange={(e) => setNickname(e.target.value)} required /></label>
        <label>Email<input value={email} onChange={(e) => setEmail(e.target.value)} type="email" required /></label>
        <label>Пароль<input value={password} onChange={(e) => setPassword(e.target.value)} type="password" minLength={6} required /></label>
        <button className="button" type="submit">Создать аккаунт</button>
      </form>
      <p className="muted">Уже есть аккаунт? <Link to="/login">Войти</Link></p>
    </section>
  );
}
