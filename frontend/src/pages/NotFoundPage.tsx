import { Link } from "react-router-dom";

export function NotFoundPage() {
  return (
    <section className="not-found">
      <h1>404</h1>
      <p>Страница не найдена.</p>
      <Link to="/listings" className="button">К поиску команд</Link>
    </section>
  );
}
