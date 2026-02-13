import React from "react";
import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { DashboardView } from "../components/dashboard-view";

describe("DashboardView", () => {
  it("renders summary", () => {
    render(<DashboardView payload={{ team_name: "Arizona Wildcats", as_of_date: "2025-02-14", window: 5, rolling: { wins: 4, losses: 1 } }} />);
    expect(screen.getByText("Arizona Wildcats Trends")).toBeTruthy();
  });
});
