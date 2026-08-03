package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rexec/rexec/internal/models"
)

// GetAllUsers retrieves all users for the admin dashboard
func (s *PostgresStore) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT id, email, username, tier, COALESCE(is_admin, false),
		       COALESCE(pipeops_id, ''), subscription_active, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var u models.User
		var pipeopsID sql.NullString
		// Assuming SubscriptionActive is a bool in your models.User
		// If it's a pointer/sql.NullBool, adjust accordingly
		err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.Username,
			&u.Tier,
			&u.IsAdmin,
			&pipeopsID,
			&u.SubscriptionActive,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		u.PipeOpsID = pipeopsID.String
		users = append(users, &u)
	}
	return users, nil
}

// GetContainerCountsByUser returns a map of userID -> active container count.
func (s *PostgresStore) GetContainerCountsByUser(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT user_id, COUNT(*)
		FROM containers
		WHERE deleted_at IS NULL
		GROUP BY user_id
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var userID string
		var count int
		if err := rows.Scan(&userID, &count); err != nil {
			return nil, err
		}
		counts[userID] = count
	}
	return counts, nil
}

// GetAllContainersAdmin retrieves all containers for the admin dashboard
// It performs a JOIN with users to get owner details
func (s *PostgresStore) GetAllContainersAdmin(ctx context.Context) ([]*models.AdminContainer, error) {
	query := `
		SELECT
			c.id, c.user_id, c.name, c.image, c.status, c.created_at,
			c.memory_mb, c.cpu_shares, c.disk_mb,
			u.username, u.email
		FROM containers c
		JOIN users u ON c.user_id = u.id
		WHERE c.deleted_at IS NULL
		ORDER BY c.created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var containers []*models.AdminContainer
	for rows.Next() {
		var ac models.AdminContainer
		ac.Resources = models.ResourceLimits{} // Initialize nested struct

		err := rows.Scan(
			&ac.ID,
			&ac.UserID,
			&ac.Name,
			&ac.Image,
			&ac.Status,
			&ac.CreatedAt,
			&ac.Resources.MemoryMB,
			&ac.Resources.CPUShares,
			&ac.Resources.DiskMB,
			&ac.Username,
			&ac.UserEmail,
		)
		if err != nil {
			return nil, err
		}
		containers = append(containers, &ac)
	}
	return containers, nil
}

// GetAllSessionsAdmin retrieves active terminal sessions for the admin dashboard
func (s *PostgresStore) GetAllSessionsAdmin(ctx context.Context) ([]*models.AdminTerminal, error) {
	// Join sessions with users and containers to provide meaningful info
	// Filter by last_ping_at to only show recently active sessions (e.g., last 5 minutes)
	// Although for now, we'll just return all sessions in the table as "active" implies
	// they haven't been deleted yet.
	query := `
		SELECT
			s.id, s.container_id, s.user_id, s.created_at,
			u.username,
			c.name as container_name, c.status
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		JOIN containers c ON s.container_id = c.id
		ORDER BY s.created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var terminals []*models.AdminTerminal
	for rows.Next() {
		var t models.AdminTerminal
		err := rows.Scan(
			&t.ID,
			&t.ContainerID,
			&t.UserID,
			&t.ConnectedAt,
			&t.Username,
			&t.Name,   // Container name
			&t.Status, // Container status
		)
		if err != nil {
			return nil, err
		}
		terminals = append(terminals, &t)
	}
	return terminals, nil
}

// DeleteUser permanently deletes a user and their cascading resources
func (s *PostgresStore) DeleteUser(ctx context.Context, id string) error {
	// You might want to implement CASCADE DELETES in your SQL schema
	// or manually delete associated resources (containers, terminals, etc.) here.
	// For simplicity, this example just deletes the user.
	query := `DELETE FROM users WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

// GetAdminUsageStats returns aggregate usage analytics for the admin dashboard.
func (s *PostgresStore) GetAdminUsageStats(ctx context.Context, from, to time.Time, interval string) (*models.AdminUsageStats, error) {
	stats := &models.AdminUsageStats{
		From:     from,
		To:       to,
		Interval: interval,
	}

	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&stats.Totals.Users); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM containers WHERE deleted_at IS NULL`).Scan(&stats.Totals.Containers); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sessions WHERE last_ping_at > NOW() - INTERVAL '5 minutes'`).Scan(&stats.Totals.ActiveSessions); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_sessions`).Scan(&stats.Totals.Logins); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM agents`).Scan(&stats.Totals.Agents); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM agents WHERE last_heartbeat > NOW() - INTERVAL '2 minutes'`).Scan(&stats.Totals.OnlineAgents); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM terminal_recordings`).Scan(&stats.Totals.Recordings); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(ROUND(SUM(duration_ms) / 3600000.0)::INTEGER, 0) FROM terminal_recordings`).Scan(&stats.Totals.RecordingHours); err != nil {
		return nil, err
	}

	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE created_at >= $1 AND created_at < $2`, from, to).Scan(&stats.Activity.NewUsers); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM containers WHERE created_at >= $1 AND created_at < $2`, from, to).Scan(&stats.Activity.NewContainers); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sessions WHERE created_at >= $1 AND created_at < $2`, from, to).Scan(&stats.Activity.NewSessions); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_sessions WHERE created_at >= $1 AND created_at < $2`, from, to).Scan(&stats.Activity.NewLogins); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM agents WHERE created_at >= $1 AND created_at < $2`, from, to).Scan(&stats.Activity.NewAgents); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM terminal_recordings WHERE created_at >= $1 AND created_at < $2`, from, to).Scan(&stats.Activity.NewRecordings); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(ROUND(SUM(duration_ms) / 3600000.0)::INTEGER, 0) FROM terminal_recordings WHERE created_at >= $1 AND created_at < $2`, from, to).Scan(&stats.Activity.RecordingHours); err != nil {
		return nil, err
	}

	buckets, err := buildAdminUsageBuckets(from, to, interval)
	if err != nil {
		return nil, err
	}

	stats.Timeline = buckets

	if err := s.fillAdminUsageSeries(ctx, stats.Timeline, `users`, `created_at`, interval, from, to, func(point *models.AdminUsagePoint, count int) {
		point.NewUsers = count
	}); err != nil {
		return nil, err
	}
	if err := s.fillAdminUsageSeries(ctx, stats.Timeline, `containers`, `created_at`, interval, from, to, func(point *models.AdminUsagePoint, count int) {
		point.NewContainers = count
	}); err != nil {
		return nil, err
	}
	if err := s.fillAdminUsageSeries(ctx, stats.Timeline, `sessions`, `created_at`, interval, from, to, func(point *models.AdminUsagePoint, count int) {
		point.NewSessions = count
	}); err != nil {
		return nil, err
	}
	if err := s.fillAdminUsageSeries(ctx, stats.Timeline, `user_sessions`, `created_at`, interval, from, to, func(point *models.AdminUsagePoint, count int) {
		point.NewLogins = count
	}); err != nil {
		return nil, err
	}
	if err := s.fillAdminUsageSeries(ctx, stats.Timeline, `agents`, `created_at`, interval, from, to, func(point *models.AdminUsagePoint, count int) {
		point.NewAgents = count
	}); err != nil {
		return nil, err
	}
	if err := s.fillAdminUsageSeries(ctx, stats.Timeline, `terminal_recordings`, `created_at`, interval, from, to, func(point *models.AdminUsagePoint, count int) {
		point.NewRecordings = count
	}); err != nil {
		return nil, err
	}

	return stats, nil
}

func buildAdminUsageBuckets(from, to time.Time, interval string) ([]models.AdminUsagePoint, error) {
	var (
		step      time.Duration
		labelFmt  string
		bucketCnt int
	)

	switch interval {
	case "hour":
		step = time.Hour
		labelFmt = "3 PM"
	case "day":
		step = 24 * time.Hour
		labelFmt = "Jan 2"
	case "week":
		step = 7 * 24 * time.Hour
		labelFmt = "Jan 2"
	case "month":
		labelFmt = "Jan 2006"
	default:
		return nil, fmt.Errorf("unsupported interval %q", interval)
	}

	buckets := make([]models.AdminUsagePoint, 0)
	// Normalize to UTC so bucket keys match SQL date_trunc(... AT TIME ZONE 'UTC').
	cursor := from.UTC()
	end := to.UTC()
	for cursor.Before(end) {
		next := cursor.Add(step)
		if interval == "month" {
			next = cursor.AddDate(0, 1, 0)
		}
		buckets = append(buckets, models.AdminUsagePoint{
			BucketStart: cursor,
			BucketLabel: cursor.Format(labelFmt),
		})
		cursor = next
		bucketCnt++
		if bucketCnt > 400 {
			return nil, fmt.Errorf("too many usage buckets requested")
		}
	}

	if len(buckets) == 0 {
		buckets = append(buckets, models.AdminUsagePoint{
			BucketStart: from.UTC(),
			BucketLabel: from.UTC().Format("Jan 2"),
		})
	}

	return buckets, nil
}

func (s *PostgresStore) fillAdminUsageSeries(ctx context.Context, timeline []models.AdminUsagePoint, table, column, interval string, from, to time.Time, assign func(point *models.AdminUsagePoint, count int)) error {
	if len(timeline) == 0 {
		return nil
	}

	// Return epoch seconds so assignment never depends on time.Time map-key
	// equality (lib/pq often scans timestamptz with a non-UTC Location, which
	// makes Go map lookups miss even when Equal() is true — leaving the chart
	// all zeros while totals still look fine).
	query := fmt.Sprintf(`
		SELECT EXTRACT(EPOCH FROM date_trunc('%s', %s AT TIME ZONE 'UTC') AT TIME ZONE 'UTC')::bigint AS bucket_unix,
		       COUNT(*)::int
		FROM %s
		WHERE %s >= $1 AND %s < $2
		GROUP BY 1
		ORDER BY 1
	`, interval, column, table, column, column)

	rows, err := s.db.QueryContext(ctx, query, from.UTC(), to.UTC())
	if err != nil {
		return err
	}
	defer rows.Close()

	points := indexAdminUsageTimeline(timeline)

	for rows.Next() {
		var bucketUnix int64
		var count int
		if err := rows.Scan(&bucketUnix, &count); err != nil {
			return err
		}
		if point, ok := points[bucketUnix]; ok {
			assign(point, count)
		}
	}

	return rows.Err()
}

// indexAdminUsageTimeline maps each bucket's UTC unix second to the point so
// SQL series rows can be merged without time.Location mismatches.
func indexAdminUsageTimeline(timeline []models.AdminUsagePoint) map[int64]*models.AdminUsagePoint {
	points := make(map[int64]*models.AdminUsagePoint, len(timeline))
	for i := range timeline {
		points[timeline[i].BucketStart.UTC().Unix()] = &timeline[i]
	}
	return points
}
