package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/rexec/rexec/internal/models"
)

const adminUserSearchMaxLen = 100

// GetAdminUsers returns a page of users for the admin dashboard.
func (s *PostgresStore) GetAdminUsers(ctx context.Context, params models.AdminUserListParams) (*models.AdminUserList, error) {
	params = normalizeAdminUserListParams(params)

	where := "TRUE"
	args := make([]interface{}, 0, 4)
	arg := 1

	if params.Search != "" {
		where += fmt.Sprintf(" AND (u.email ILIKE $%d ESCAPE '\\' OR u.username ILIKE $%d ESCAPE '\\' OR u.id ILIKE $%d ESCAPE '\\')", arg, arg, arg)
		args = append(args, adminUserSearchPattern(params.Search))
		arg++
	}
	if params.Subscribers {
		where += " AND u.subscription_active = TRUE"
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM users u WHERE ` + where
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	if total > 0 && params.Offset() >= total {
		params.Page = params.TotalPages(total)
	}

	listQuery := fmt.Sprintf(`
		SELECT
			u.id, u.email, u.username, u.tier, COALESCE(u.is_admin, false),
			COALESCE(u.pipeops_id, ''), COALESCE(u.subscription_active, false),
			u.created_at, u.updated_at,
			COALESCE((
				SELECT COUNT(*)::int
				FROM containers c
				WHERE c.user_id = u.id AND c.deleted_at IS NULL
			), 0) AS container_count
		FROM users u
		WHERE %s
		ORDER BY u.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, arg, arg+1)

	listArgs := append(append([]interface{}{}, args...), params.PerPage, params.Offset())
	rows, err := s.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*models.AdminUser, 0, params.PerPage)
	for rows.Next() {
		var u models.AdminUser
		var pipeopsID sql.NullString
		if err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.Username,
			&u.Tier,
			&u.IsAdmin,
			&pipeopsID,
			&u.SubscriptionActive,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.ContainerCount,
		); err != nil {
			return nil, err
		}
		u.PipeOpsID = pipeopsID.String
		users = append(users, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &models.AdminUserList{
		Users:      users,
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      total,
		TotalPages: params.TotalPages(total),
	}, nil
}

func normalizeAdminUserListParams(params models.AdminUserListParams) models.AdminUserListParams {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Page > 10000 {
		params.Page = 10000
	}
	if params.PerPage < 1 {
		params.PerPage = 25
	}
	if params.PerPage > 100 {
		params.PerPage = 100
	}
	params.Search = strings.TrimSpace(params.Search)
	if len(params.Search) > adminUserSearchMaxLen {
		params.Search = params.Search[:adminUserSearchMaxLen]
	}
	return params
}

func adminUserSearchPattern(search string) string {
	return "%" + escapeLikePattern(search) + "%"
}

func escapeLikePattern(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
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
// Totals/activity are one round trip; the timeline is a second UNION ALL query.
func (s *PostgresStore) GetAdminUsageStats(ctx context.Context, from, to time.Time, interval string) (*models.AdminUsageStats, error) {
	stats := &models.AdminUsageStats{
		From:     from,
		To:       to,
		Interval: interval,
	}

	if err := s.scanAdminUsageScalars(ctx, stats, from, to); err != nil {
		return nil, err
	}

	buckets, err := buildAdminUsageBuckets(from, to, interval)
	if err != nil {
		return nil, err
	}
	stats.Timeline = buckets

	if err := s.fillAdminUsageTimeline(ctx, stats.Timeline, interval, from, to); err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *PostgresStore) scanAdminUsageScalars(ctx context.Context, stats *models.AdminUsageStats, from, to time.Time) error {
	const query = `
		SELECT
			(SELECT COUNT(*) FROM users),
			(SELECT COUNT(*) FROM users WHERE subscription_active = TRUE),
			(SELECT COUNT(*) FROM containers WHERE deleted_at IS NULL),
			(SELECT COUNT(*) FROM sessions WHERE last_ping_at > NOW() - INTERVAL '5 minutes'),
			(SELECT COUNT(*) FROM user_sessions),
			(SELECT COUNT(*) FROM agents),
			(SELECT COUNT(*) FROM agents WHERE last_heartbeat > NOW() - INTERVAL '2 minutes'),
			(SELECT COUNT(*) FROM terminal_recordings),
			(SELECT COALESCE(ROUND(SUM(duration_ms) / 3600000.0)::INTEGER, 0) FROM terminal_recordings),
			(SELECT COUNT(*) FROM users WHERE created_at >= $1 AND created_at < $2),
			(SELECT COUNT(*) FROM containers WHERE created_at >= $1 AND created_at < $2),
			(SELECT COUNT(*) FROM sessions WHERE created_at >= $1 AND created_at < $2),
			(SELECT COUNT(*) FROM user_sessions WHERE created_at >= $1 AND created_at < $2),
			(SELECT COUNT(*) FROM agents WHERE created_at >= $1 AND created_at < $2),
			(SELECT COUNT(*) FROM terminal_recordings WHERE created_at >= $1 AND created_at < $2),
			(SELECT COALESCE(ROUND(SUM(duration_ms) / 3600000.0)::INTEGER, 0) FROM terminal_recordings WHERE created_at >= $1 AND created_at < $2)
	`
	return s.db.QueryRowContext(ctx, query, from.UTC(), to.UTC()).Scan(
		&stats.Totals.Users,
		&stats.Totals.Subscribers,
		&stats.Totals.Containers,
		&stats.Totals.ActiveSessions,
		&stats.Totals.Logins,
		&stats.Totals.Agents,
		&stats.Totals.OnlineAgents,
		&stats.Totals.Recordings,
		&stats.Totals.RecordingHours,
		&stats.Activity.NewUsers,
		&stats.Activity.NewContainers,
		&stats.Activity.NewSessions,
		&stats.Activity.NewLogins,
		&stats.Activity.NewAgents,
		&stats.Activity.NewRecordings,
		&stats.Activity.RecordingHours,
	)
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

func (s *PostgresStore) fillAdminUsageTimeline(ctx context.Context, timeline []models.AdminUsagePoint, interval string, from, to time.Time) error {
	if len(timeline) == 0 {
		return nil
	}

	// Return epoch seconds so assignment never depends on time.Time map-key
	// equality (lib/pq often scans timestamptz with a non-UTC Location, which
	// makes Go map lookups miss even when Equal() is true — leaving the chart
	// all zeros while totals still look fine).
	const query = `
		SELECT series,
		       EXTRACT(EPOCH FROM date_trunc($3, ts AT TIME ZONE 'UTC') AT TIME ZONE 'UTC')::bigint AS bucket_unix,
		       COUNT(*)::int
		FROM (
			SELECT 'users'::text AS series, created_at AS ts FROM users WHERE created_at >= $1 AND created_at < $2
			UNION ALL
			SELECT 'containers', created_at FROM containers WHERE created_at >= $1 AND created_at < $2
			UNION ALL
			SELECT 'sessions', created_at FROM sessions WHERE created_at >= $1 AND created_at < $2
			UNION ALL
			SELECT 'logins', created_at FROM user_sessions WHERE created_at >= $1 AND created_at < $2
			UNION ALL
			SELECT 'agents', created_at FROM agents WHERE created_at >= $1 AND created_at < $2
			UNION ALL
			SELECT 'recordings', created_at FROM terminal_recordings WHERE created_at >= $1 AND created_at < $2
		) events
		GROUP BY 1, 2
	`

	rows, err := s.db.QueryContext(ctx, query, from.UTC(), to.UTC(), interval)
	if err != nil {
		return err
	}
	defer rows.Close()

	points := indexAdminUsageTimeline(timeline)

	for rows.Next() {
		var series string
		var bucketUnix int64
		var count int
		if err := rows.Scan(&series, &bucketUnix, &count); err != nil {
			return err
		}
		point, ok := points[bucketUnix]
		if !ok {
			continue
		}
		assignAdminUsageCount(point, series, count)
	}

	return rows.Err()
}

func assignAdminUsageCount(point *models.AdminUsagePoint, series string, count int) {
	if point == nil {
		return
	}
	switch series {
	case "users":
		point.NewUsers = count
	case "containers":
		point.NewContainers = count
	case "sessions":
		point.NewSessions = count
	case "logins":
		point.NewLogins = count
	case "agents":
		point.NewAgents = count
	case "recordings":
		point.NewRecordings = count
	}
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
