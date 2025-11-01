package handlers

import (
	"encoding/json"
	"fmt"
	"go-practice5/internal/repository"
	"net/http"
	"strconv"
	"time"

	"database/sql"
)

func GetMovies(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		var yearMinPtr, yearMaxPtr *int
		if v := q.Get("year_min"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				yearMinPtr = &n
			}
		}
		if v := q.Get("year_max"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				yearMaxPtr = &n
			}
		}

		limit := 100
		if v := q.Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				limit = n
			}
		}
		offset := 0
		if v := q.Get("offset"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n >= 0 {
				offset = n
			}
		}

		f := repository.MovieFilters{
			YearMin: yearMinPtr,
			YearMax: yearMaxPtr,
			Limit:   limit,
			Offset:  offset,
		}

		// 3s deadline to the DB query
		ctx := r.Context()
		type result struct {
			data    any
			elapsed time.Duration
			err     error
		}

		movies, elapsed, err := repository.QueryMovies(ctx, db, f)
		w.Header().Set("X-Query-Time", fmt.Sprintf("%.3fms", float64(elapsed.Microseconds())/1000.0))
		if err != nil {
			http.Error(w, "query error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(movies)
	}
}
