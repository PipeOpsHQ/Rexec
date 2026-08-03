package storage

import (
	"testing"
	"time"

	"github.com/rexec/rexec/internal/models"
)

func TestBuildAdminUsageBucketsDay(t *testing.T) {
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)

	buckets, err := buildAdminUsageBuckets(from, to, "day")
	if err != nil {
		t.Fatalf("buildAdminUsageBuckets: %v", err)
	}
	if len(buckets) != 3 {
		t.Fatalf("expected 3 day buckets, got %d", len(buckets))
	}
	if !buckets[0].BucketStart.Equal(from) {
		t.Fatalf("first bucket start = %v, want %v", buckets[0].BucketStart, from)
	}
	if buckets[0].BucketStart.Location() != time.UTC {
		t.Fatalf("bucket location = %v, want UTC", buckets[0].BucketStart.Location())
	}
}

func TestBuildAdminUsageBucketsMonth(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	buckets, err := buildAdminUsageBuckets(from, to, "month")
	if err != nil {
		t.Fatalf("buildAdminUsageBuckets: %v", err)
	}
	if len(buckets) != 3 {
		t.Fatalf("expected 3 month buckets, got %d (%v)", len(buckets), buckets)
	}
	if buckets[1].BucketStart.Month() != time.February {
		t.Fatalf("second bucket month = %v, want February", buckets[1].BucketStart.Month())
	}
}

func TestIndexAdminUsageTimelineMatchesDriverLocations(t *testing.T) {
	// Reproduces the production bug: SQL buckets scanned by lib/pq often have a
	// different *time.Location than Go-built UTC buckets, so map[time.Time]
	// lookups miss even when the instants are Equal.
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	timeline := []models.AdminUsagePoint{
		{BucketStart: start, BucketLabel: "Aug 1"},
		{BucketStart: start.Add(24 * time.Hour), BucketLabel: "Aug 2"},
	}
	index := indexAdminUsageTimeline(timeline)

	// Same wall clock, empty location offset 0 — common after some drivers.
	driverBucket := time.Date(2026, 8, 1, 0, 0, 0, 0, time.FixedZone("", 0))
	if !driverBucket.Equal(start) {
		t.Fatal("test setup: driver bucket should Equal start")
	}
	if _, ok := map[time.Time]struct{}{start: {}}[driverBucket]; ok {
		t.Fatal("expected time.Time map key mismatch for FixedZone(\"\", 0) vs UTC")
	}

	point, ok := index[driverBucket.UTC().Unix()]
	if !ok || point == nil {
		t.Fatalf("unix index miss for driver bucket %v", driverBucket)
	}
	if point.BucketLabel != "Aug 1" {
		t.Fatalf("assigned wrong point: %q", point.BucketLabel)
	}

	// Local-zone equivalent instant should also hit via Unix seconds.
	localEq := start.In(time.Local)
	if _, ok := index[localEq.UTC().Unix()]; !ok {
		t.Fatalf("unix index miss for local equivalent %v", localEq)
	}
}

func TestAssignSeriesViaUnixIndex(t *testing.T) {
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	timeline := []models.AdminUsagePoint{
		{BucketStart: from},
		{BucketStart: from.Add(24 * time.Hour)},
	}
	index := indexAdminUsageTimeline(timeline)

	// Simulate two SQL rows with non-UTC locations.
	rows := []struct {
		bucket time.Time
		count  int
	}{
		{time.Date(2026, 8, 1, 0, 0, 0, 0, time.FixedZone("UTC", 0)), 5},
		{from.Add(24 * time.Hour).In(time.Local), 3},
	}

	for _, row := range rows {
		if point, ok := index[row.bucket.UTC().Unix()]; ok {
			point.NewUsers = row.count
		}
	}

	if timeline[0].NewUsers != 5 {
		t.Fatalf("day0 users = %d, want 5", timeline[0].NewUsers)
	}
	if timeline[1].NewUsers != 3 {
		t.Fatalf("day1 users = %d, want 3", timeline[1].NewUsers)
	}
}
