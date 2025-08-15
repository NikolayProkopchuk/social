package store

import (
	"net/http"
	"net/url"
	"strconv"
)

type PaginatedFeedQuery struct {
	Limit  int    `json:"limit" validate:"gte=1,lte=100"`
	Offset int    `json:"offset" validate:"gte=0"`
	Sort   string `json:"sort" validate:"omitempty,oneof=asc desc"`
}

func ParsePaginatedFeedQuery(r *http.Request) (*PaginatedFeedQuery, error) {
	query := r.URL.Query()
	limit, err := getDefaultQueryIntParam(&query, "limit", 10)
	if err != nil {
		return nil, err
	}
	offset, err := getDefaultQueryIntParam(&query, "offset", 0)
	if err != nil {
		return nil, err
	}
	sort := query.Get("sort")
	paginatedFeedQuery := PaginatedFeedQuery{
		Limit:  limit,
		Offset: offset,
		Sort:   sort,
	}

	return &paginatedFeedQuery, nil
}

func getDefaultQueryIntParam(values *url.Values, key string, defaultValue int) (int, error) {
	urlParam := values.Get(key)
	if urlParam == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(urlParam)
}
