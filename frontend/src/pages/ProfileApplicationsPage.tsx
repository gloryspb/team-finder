import { FormEvent, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../api/auth";
import type { ApplicationDetails } from "../types";
import { Notice } from "../components/Notice";

export function ProfileApplicationsPage() {
  const { profile, refreshProfile } = useAuth();
  const [incoming, setIncoming] = useState<ApplicationDetails[]>([]);
  const [outgoing, setOutgoing] = useState<ApplicationDetails[]>([]);
  const [form, setForm] = useState({
    nickname: "", region: "", languages: "", voice_chat: false, bio: ""
  });
  const [notice, setNotice] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    if (profile) {
      setForm({
        nickname: profile.nickname,
        region: profile.region,
        languages: profile.languages.join(", "),
        voice_chat: profile.voice_chat,
        bio: profile.bio
      });
    }
    loadApplications();
  }, [profile?.id]);

  async function loadApplications() {
    try {
      const [inApps, outApps] = await Promise.all([api.incoming(), api.outgoing()]);
      setIncoming(inApps);
      setOutgoing(outApps);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось загрузить заявки");
    }
  }

  async function saveProfile(event: FormEvent) {
    event.preventDefault();
    setError("");
    setNotice("");
    try {
      await api.updateProfile({
        nickname: form.nickname,
        region: form.region,
        languages: form.languages.split(",").map((item) => item.trim()).filter(Boolean),
        voice_chat: form.voice_chat,
        bio: form.bio
      });
      await refreshProfile();
      setNotice("Профиль обновлён.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось сохранить профиль");
    }
  }

  async function setStatus(id: string, status: "accepted" | "rejected") {
    setError("");
    try {
      await api.setApplicationStatus(id, status);
      setNotice(status === "accepted" ? "Заявка принята." : "Заявка отклонена.");
      await loadApplications();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось изменить статус");
    }
  }

  return (
    <section>
      <h1>Профиль и заявки</h1>
      {error && <Notice type="error">{error}</Notice>}
      {notice && <Notice type="success">{notice}</Notice>}
      <div className="split">
        <form className="form" onSubmit={saveProfile}>
          <h2>Профиль игрока</h2>
          <label>Никнейм<input value={form.nickname} onChange={(e) => setForm({ ...form, nickname: e.target.value })} required /></label>
          <label>Регион<input value={form.region} onChange={(e) => setForm({ ...form, region: e.target.value })} /></label>
          <label>Языки<input value={form.languages} onChange={(e) => setForm({ ...form, languages: e.target.value })} placeholder="ru, en" /></label>
          <label className="checkbox"><input type="checkbox" checked={form.voice_chat} onChange={(e) => setForm({ ...form, voice_chat: e.target.checked })} /> Голосовой чат</label>
          <label>О себе<textarea value={form.bio} onChange={(e) => setForm({ ...form, bio: e.target.value })} maxLength={1000} rows={5} /></label>
          <button className="button" type="submit">Сохранить</button>
        </form>
        <div>
          <h2>Исходящие заявки</h2>
          <ApplicationList applications={outgoing} />
        </div>
      </div>
      <section>
        <h2>Входящие заявки</h2>
        <div className="list">
          {incoming.map((app) => (
            <article className="application" key={app.id}>
              <div>
                <strong>{app.applicant_profile.nickname}</strong>
                <p>{app.message || "Без сообщения."}</p>
                <Link to={`/listings/${app.listing.id}`}>{app.listing.title} · {app.game.name}</Link>
              </div>
              <div className="actions">
                <span className={`badge ${app.status}`}>{statusLabel(app.status)}</span>
                {app.status === "pending" && (
                  <>
                    <button className="button small" onClick={() => setStatus(app.id, "accepted")}>Принять</button>
                    <button className="ghost danger" onClick={() => setStatus(app.id, "rejected")}>Отклонить</button>
                  </>
                )}
              </div>
            </article>
          ))}
          {!incoming.length && <p className="empty">Входящих заявок пока нет.</p>}
        </div>
      </section>
    </section>
  );
}

function ApplicationList({ applications }: { applications: ApplicationDetails[] }) {
  return (
    <div className="list">
      {applications.map((app) => (
        <article className="row" key={app.id}>
          <Link to={`/listings/${app.listing.id}`}>{app.listing.title}</Link>
          <span>{app.game.name}</span>
          <span className={`badge ${app.status}`}>{statusLabel(app.status)}</span>
        </article>
      ))}
      {!applications.length && <p className="empty">Исходящих заявок пока нет.</p>}
    </div>
  );
}

function statusLabel(status: string) {
  return status === "accepted" ? "принята" : status === "rejected" ? "отклонена" : "ожидает";
}
