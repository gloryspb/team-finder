import { FormEvent, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import type { Game } from "../types";
import { Notice } from "../components/Notice";

export function ListingCreatePage() {
  const navigate = useNavigate();
  const [games, setGames] = useState<Game[]>([]);
  const [form, setForm] = useState({
    game_id: "", title: "", mode: "", required_roles: [] as string[], rank_min: "", rank_max: "", region: "", description: ""
  });
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    api.games().then((items) => {
      setGames(items);
      setForm((current) => ({ ...current, game_id: items[0]?.id ?? "" }));
    }).catch((err) => setError(err.message));
  }, []);

  const game = useMemo(() => games.find((item) => item.id === form.game_id), [games, form.game_id]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    setError("");
    setMessage("");
    try {
      const listing = await api.createListing(form);
      setMessage("Объявление создано.");
      navigate(`/listings/${listing.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось создать объявление");
    }
  }

  function toggleRole(role: string) {
    setForm((current) => ({
      ...current,
      required_roles: current.required_roles.includes(role)
        ? current.required_roles.filter((item) => item !== role)
        : [...current.required_roles, role]
    }));
  }

  return (
    <section className="editor">
      <h1>Создать объявление</h1>
      <form className="form wide" onSubmit={submit}>
        {error && <Notice type="error">{error}</Notice>}
        {message && <Notice type="success">{message}</Notice>}
        <label>Игра
          <select value={form.game_id} onChange={(e) => setForm({ ...form, game_id: e.target.value, mode: "", required_roles: [] })} required>
            {games.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </select>
        </label>
        <label>Название<input value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} required /></label>
        <div className="grid-two">
          <label>Режим
            <select value={form.mode} onChange={(e) => setForm({ ...form, mode: e.target.value })}>
              <option value="">Не важно</option>
              {game?.modes.map((mode) => <option key={mode} value={mode}>{mode}</option>)}
            </select>
          </label>
          <label>Регион<input value={form.region} onChange={(e) => setForm({ ...form, region: e.target.value })} placeholder="EU, CIS, NA" /></label>
        </div>
        <div className="grid-two">
          <label>Ранг от<input value={form.rank_min} onChange={(e) => setForm({ ...form, rank_min: e.target.value })} /></label>
          <label>Ранг до<input value={form.rank_max} onChange={(e) => setForm({ ...form, rank_max: e.target.value })} /></label>
        </div>
        <div>
          <span className="label">Нужные роли</span>
          <div className="chips">
            {game?.roles.map((role) => (
              <button type="button" className={form.required_roles.includes(role) ? "chip active" : "chip"} onClick={() => toggleRole(role)} key={role}>{role}</button>
            ))}
          </div>
        </div>
        <label>Описание<textarea value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} maxLength={2000} rows={6} /></label>
        <button className="button" type="submit">Опубликовать</button>
      </form>
    </section>
  );
}
