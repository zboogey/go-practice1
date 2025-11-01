package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go-practice5/internal/models"
)

type MovieFilters struct {
	YearMin *int
	YearMax *int
	Limit   int
	Offset  int
}

func QueryMovies(ctx context.Context, db *sql.DB, f MovieFilters) ([]models.Movie, time.Duration, error) {
	var sb strings.Builder
	var args []any

	sb.WriteString(`
SELECT
  m.id, m.title, m.year,
  COALESCE(ac.cnt, 0) AS actor_count
FROM movies m
LEFT JOIN (
  SELECT movie_id, COUNT(*) AS cnt
  FROM actors
  GROUP BY movie_id
) ac ON ac.movie_id = m.id
WHERE 1=1
`)

	if f.YearMin != nil {
		args = append(args, *f.YearMin)
		sb.WriteString(fmt.Sprintf(" AND m.year >= $%d", len(args)))
	}
	if f.YearMax != nil {
		args = append(args, *f.YearMax)
		sb.WriteString(fmt.Sprintf(" AND m.year <= $%d", len(args)))
	}

	if f.Limit <= 0 {
		f.Limit = 100
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	args = append(args, f.Limit, f.Offset)
	sb.WriteString(fmt.Sprintf(`
ORDER BY m.year DESC, m.id ASC
LIMIT $%d OFFSET $%d
`, len(args)-1, len(args)))

	query := sb.String()

	start := time.Now()
	rows, err := db.QueryContext(ctx, query, args...)
	elapsed := time.Since(start)
	if err != nil {
		return nil, elapsed, err
	}
	defer rows.Close()

	var out []models.Movie
	for rows.Next() {
		var m models.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Year, &m.ActorCount); err != nil {
			return nil, elapsed, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, elapsed, err
	}

	return out, elapsed, nil
}
