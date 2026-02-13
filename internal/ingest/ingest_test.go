package ingest

import "testing"

func TestParseGame(t *testing.T) {
	rec := []string{"2025-01-01", "UCLA", "W", "82", "74", "69"}
	g, err := parseGame(1, rec)
	if err != nil {
		t.Fatal(err)
	}
	if g.TeamID != 1 || g.Result != "W" || g.PointsFor != 82 {
		t.Fatalf("bad parse: %+v", g)
	}
	if g.OffensiveRating <= 0 || g.DefensiveRating <= 0 {
		t.Fatalf("ratings missing: %+v", g)
	}
}
