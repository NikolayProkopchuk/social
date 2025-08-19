package store

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type PaginatedFeedQuery struct {
	Limit  int       `json:"limit" validate:"gte=1,lte=100"`
	Offset int       `json:"offset" validate:"gte=0"`
	Sort   string    `json:"sort" validate:"omitempty,oneof=asc desc"`
	Tags   []string  `json:"tags" validate:"max=5"`
	Search string    `json:"search" validate:"omitempty,max=50"`
	Since  time.Time `json:"since"`
	Until  time.Time `json:"until"`
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

	if tagsParam := query.Get("tags"); tagsParam != "" {
		paginatedFeedQuery.Tags = strings.Split(tagsParam, ",")
	}
	paginatedFeedQuery.Search = query.Get("search")
	if sinceParam := query.Get("since"); sinceParam != "" {
		since, timeParsErr := time.Parse(time.RFC3339, sinceParam)
		if timeParsErr != nil {
			return nil, timeParsErr
		}
		paginatedFeedQuery.Since = since
	}
	if untilParam := query.Get("until"); untilParam != "" {
		until, timeParsErr := time.Parse(time.RFC3339, untilParam)
		if timeParsErr != nil {
			return nil, timeParsErr
		}
		paginatedFeedQuery.Until = until
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
