import { DashboardView } from "../components/dashboard-view";
import { fetchDashboard } from "../lib/api";

export default async function Home() {
  const payload = await fetchDashboard("arizona-mbb", 5);
  return <DashboardView payload={payload} />;
}
