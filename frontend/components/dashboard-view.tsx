import React from "react";
export function DashboardView({ payload }: { payload: any }) {
  return (
    <main>
      <h1>{payload.team_name} Trends</h1>
      <p>As of {payload.as_of_date}</p>
      <section>
        <h2>Rolling {payload.window}</h2>
        <p>Record: {payload.rolling.wins}-{payload.rolling.losses}</p>
      </section>
    </main>
  );
}
