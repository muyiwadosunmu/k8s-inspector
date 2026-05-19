package web

import (
	"net/http"
	"strconv"
)

// PageResponse is the standard form used for paginated responses.
type PageResponse[T any] struct {
	Items    []T `json:"items"`
	Metadata any `json:"metadata,omitempty"`
}

// Request holds standard request information.
type Request struct {
	Page           int
	PageSize       int
	OrderBy        string
	OrderDirection string
}

// ParsePagination parses pagination from the request query string.
// It returns sensible defaults and validates the values.
func ParsePagination(r *http.Request) (page int, pageSize int) {
	qs := r.URL.Query()

	page = readInt(qs.Get("page"), 1)
	pageSize = readInt(qs.Get("page_size"), 20)

	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return page, pageSize
}

func readInt(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return i
}
