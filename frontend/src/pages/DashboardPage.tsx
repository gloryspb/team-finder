import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../api/auth";
import type { ApplicationDetails, ListingDetails } from "../types";
import { Notice } from "../components/Notice";

export function DashboardPage() {
  const { user, profile } = useAuth();
  const [listings, setListings] = useState<ListingDetails[]>([]);
  const [incoming, setIncoming] = useState<ApplicationDetails[]>([]);
  const [outgoing, setOutgoing] = useState<ApplicationDetails[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    async function load() {
      try {
        const [all, inApps, outApps] = await Promise.all([api.listings(), api.incoming(), api.outgoing()]);
        setListings(all);
        setIncoming(inApps);
        setOutgoing(outApps);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Не удалось загрузить сводку");
      }
    }
    load();
  }, []);

  const myListings = useMemo(() => listings.filter((item) => item.owner_id === user?.id), [listings, user]);

  return (
    <section>
      <div className="page-title">
        <div>
          <h1>Панель игрока</h1>
          <p className="muted">Добро пожаловать, {profile?.nickname ?? "игрок"}.</p>
        </div>
        <Link to="/listings/create" className="button">Создать объявление</Link>
      </div>
      {error && <Notice type="error">{error}</Notice>}
      <div className="stats">
        <article><strong>{listings.length}</strong><span>открытых объявлений</span></article>
        <article><strong>{myListings.length}</strong><span>моих объявлений</span></article>
        <article><strong>{incoming.length}</strong><span>входящих заявок</span></article>
        <article><strong>{outgoing.length}</strong><span>исходящих заявок</span></article>
      </div>
      <div className="split">
        <section>
          <h2>Мои объявления</h2>
          <div className="list">
            {myListings.map((listing) => <ListingRow key={listing.id} listing={listing} />)}
            {!myListings.length && <p className="empty">Вы ещё не создавали объявления.</p>}
          </div>
        </section>
        <section>
          <h2>Последние заявки</h2>
          <div className="list">
            {incoming.slice(0, 4).map((app) => (
              <article className="row" key={app.id}>
                <span>{app.applicant_profile.nickname}</span>
                <strong>{app.listing.title}</strong>
                <span className={`badge ${app.status}`}>{statusLabel(app.status)}</span>
              </article>
            ))}
            {!incoming.length && <p className="empty">Входящих заявок пока нет.</p>}
          </div>
        </section>
      </div>
    </section>
  );
}

function ListingRow({ listing }: { listing: ListingDetails }) {
  return (
    <article className="row">
      <span>{listing.game.name}</span>
      <Link to={`/listings/${listing.id}`}>{listing.title}</Link>
      <span className="badge">{listing.applications_count} заявок</span>
    </article>
  );
}

function statusLabel(status: string) {
  return status === "accepted" ? "принята" : status === "rejected" ? "отклонена" : "ожидает";
}
