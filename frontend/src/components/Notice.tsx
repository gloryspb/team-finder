import type { ReactNode } from "react";

export function Notice({ type = "info", children }: { type?: "info" | "success" | "error"; children: ReactNode }) {
  return <div className={`notice ${type}`}>{children}</div>;
}
