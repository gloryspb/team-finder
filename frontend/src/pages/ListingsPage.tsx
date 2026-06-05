import { FormEvent, useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api/client";
import type { Game, ListingDetails } from "../types";
import { Notice } from "../components/Notice";

export function ListingsPage() {
  const [games, setGames] = useState<Game[]>([]);
  const [listings, setListings] = useState<ListingDetails[]>([]);
  const [filters, setFilters] = useState({ game_id: "", role: "", region: "", mode: "", search: "" });
  const [error, setError] = useState("");

  useEffect(() => {
    api.games().then(setGames).catch((err) => setError(err.message));
    load();
  }, []);

  async function load(nextFilters = filters) {
    try {
      setError("");
      const items = await api.listings(nextFilters);
      setListings(items);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось загрузить объявления");
    }
  }

  function submit(event: FormEvent) {
    event.preventDefault();
    load();
  }

  const selectedGame = useMemo(() => games.find((game) => game.id === filters.game_id), [games, filters.game_id]);

  return (
    <section>
      <div className="page-title">
        <div>
          <h1>Поиск команды</h1>
          <p className="muted">Подберите игру, роль, регион и режим.</p>
        </div>
        <Link to="/listings/create" className="button">Новое объявление</Link>
      </div>
      {error && <Notice type="error">{error}</Notice>}
      <form className="filters" onSubmit={submit}>
        <select value={filters.game_id} onChange={(e) => setFilters({ ...filters, game_id: e.target.value, role: "", mode: "" })}>
          <option value="">Все игры</option>
          {games.map((game) => <option value={game.id} key={game.id}>{game.name}</option>)}
        </select>
        <select value={filters.role} onChange={(e) => setFilters({ ...filters, role: e.target.value })}>
          <option value="">Любая роль</option>
          {(selectedGame?.roles ?? games.flatMap((game) => game.roles)).map((role) => <option value={role} key={role}>{role}</option>)}
        </select>
        <select value={filters.mode} onChange={(e) => setFilters({ ...filters, mode: e.target.value })}>
          <option value="">Любой режим</option>
          {(selectedGame?.modes ?? games.flatMap((game) => game.modes)).map((mode) => <option value={mode} key={mode}>{mode}</option>)}
        </select>
        <input value={filters.region} onChange={(e) => setFilters({ ...filters, region: e.target.value })} placeholder="Регион" />
        <input value={filters.search} onChange={(e) => setFilters({ ...filters, search: e.target.value })} placeholder="Поиск по названию" />
        <button className="button" type="submit">Найти</button>
      </form>
      <div className="cards">
        {listings.map((listing) => <ListingCard key={listing.id} listing={listing} />)}
        {!listings.length && <p className="empty">Подходящих объявлений не найдено.</p>}
      </div>
    </section>
  );
}

function ListingCard({ listing }: { listing: ListingDetails }) {
  return (
    <article className="listing-card">
      <div className="card-head">
        <span className="game-name">{listing.game.name}</span>
        <span className={`badge ${listing.status}`}>{listing.status === "open" ? "открыто" : "закрыто"}</span>
      </div>
      <h2><Link to={`/listings/${listing.id}`}>{listing.title}</Link></h2>
      <p>{listing.description || "Описание не указано."}</p>
      <div className="tags">
        {listing.mode && <span>{listing.mode}</span>}
        {listing.region && <span>{listing.region}</span>}
        {listing.required_roles.map((role) => <span key={role}>{role}</span>)}
      </div>
      <footer>Автор: {listing.owner_profile.nickname} · заявок: {listing.applications_count}</footer>
    </article>
  );
}
