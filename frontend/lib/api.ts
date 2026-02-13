export async function fetchDashboard(teamSlug: string, window: number) {
  const base = process.env.DASHBOARD_API_BASE_URL ?? "http://localhost:8080";
  const res = await fetch(`${base}/api/v1/teams/${teamSlug}/dashboard?window=${window}`, { cache: "no-store" });
  if (!res.ok) throw new Error(`failed: ${res.status}`);
  return res.json();
}
