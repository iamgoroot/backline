package elastic

import (
	"bytes"
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/iamgoroot/backline/pkg/core"
	"github.com/segmentio/encoding/json"
)

var errSearchFailed = fmt.Errorf("search failed")

type elasticResult struct {
	Hits struct {
		Hits []struct {
			Source    document `json:"_source"`
			Highlight struct {
				Value []string `json:"value"`
			} `json:"highlight"`
		} `json:"hits"`
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
	} `json:"hits"`
}

func (plugin *Search) Search(ctx context.Context, query string, offset, limit int) ([]core.SearchResult, error) {
	req, err := plugin.buildESQuery(query, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("query build failed: %w", err)
	}

	res, err := req.Do(ctx, plugin.client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("%w %s", errSearchFailed, res.String())
	}

	return processResults(res)
}

func processResults(res *esapi.Response) ([]core.SearchResult, error) {
	var esResult elasticResult
	if err := json.NewDecoder(res.Body).Decode(&esResult); err != nil {
		return nil, fmt.Errorf("failed to decode elasticsearch response: %w", err)
	}

	results := make([]core.SearchResult, 0, len(esResult.Hits.Hits))

	for _, hit := range esResult.Hits.Hits {
		var highlight string
		if len(hit.Highlight.Value) > 0 {
			highlight = hit.Highlight.Value[0]
		}

		result := core.SearchResult{
			EntityName: hit.Source.EntityName,
			Category:   hit.Source.Category,
			Link:       hit.Source.Link,
			Highlight:  highlight,
		}
		results = append(results, result)
	}

	return results, nil
}

func (plugin *Search) buildESQuery(query string, offset, limit int) (esapi.SearchRequest, error) {
	esQuery := esSearchQuery{}

	esQuery.Query.MultiMatch.Query = query
	esQuery.Query.MultiMatch.Fuzziness = 2

	esQuery.Highlight.Fields.Value.NumberOfFragments = 1
	esQuery.Source.Excludes = []string{"value"}
	esQuery.Size = limit
	esQuery.From = offset

	queryJSON, err := json.Marshal(esQuery)
	if err != nil {
		return esapi.SearchRequest{}, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{defaultSearchIndexName},
		Body:  bytes.NewReader(queryJSON),
	}

	return req, nil
}
