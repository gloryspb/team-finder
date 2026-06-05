import { FormEvent, useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../api/auth";
import type { ListingDetails } from "../types";
import { Notice } from "../components/Notice";

export function ListingDetailsPage() {
  const { id = "" } = useParams();
  const { user, isAuthenticated } = useAuth();
  const [listing, setListing] = useState<ListingDetails | null>(null);
  const [message, setMessage] = useState("");
  const [notice, setNotice] = useState("");
  const [error, setError] = useState("");

  async function load() {
    try {
      setListing(await api.listing(id));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось загрузить объявление");
    }
  }

  useEffect(() => {
    load();
  }, [id]);

  async function apply(event: FormEvent) {
    event.preventDefault();
    setError("");
    setNotice("");
    try {
      await api.apply(id, message);
      setMessage("");
      setNotice("Заявка отправлена.");
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось отправить заявку");
    }
  }

  async function close() {
    if (!listing) return;
    try {
      await api.closeListing(listing.id);
      setNotice("Объявление закрыто.");
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось закрыть объявление");
    }
  }

  if (!listing) {
    return <p className="empty">{error || "Загрузка..."}</p>;
  }

  const isOwner = user?.id === listing.owner_id;

  return (
    <section className="details">
      {error && <Notice type="error">{error}</Notice>}
      {notice && <Notice type="success">{notice}</Notice>}
      <div className="details-head">
        <div>
          <span className="game-name">{listing.game.name}</span>
          <h1>{listing.title}</h1>
          <p className="muted">Автор: {listing.owner_profile.nickname} · заявок: {listing.applications_count}</p>
        </div>
        <span className={`badge ${listing.status}`}>{listing.status === "open" ? "открыто" : "закрыто"}</span>
      </div>
      <div className="tags">
        {listing.mode && <span>{listing.mode}</span>}
        {listing.region && <span>{listing.region}</span>}
        {listing.rank_min && <span>от {listing.rank_min}</span>}
        {listing.rank_max && <span>до {listing.rank_max}</span>}
        {listing.required_roles.map((role) => <span key={role}>{role}</span>)}
      </div>
      <p className="description">{listing.description || "Автор не добавил описание."}</p>
      {isOwner && listing.status === "open" && <button className="ghost danger" onClick={close}>Закрыть объявление</button>}
      {!isOwner && listing.status === "open" && (
        isAuthenticated ? (
          <form className="form apply" onSubmit={apply}>
            <label>Сообщение к заявке<textarea value={message} onChange={(e) => setMessage(e.target.value)} rows={4} /></label>
            <button className="button" type="submit">Отправить заявку</button>
          </form>
        ) : (
          <Notice>Чтобы отправить заявку, нужно <Link to="/login">войти</Link>.</Notice>
        )
      )}
    </section>
  );
}
