package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

var errIndexCreationFailed = fmt.Errorf("elasticsearch index creation failed")

func (plugin *Search) Index(ctx context.Context, entityName, ref, category, value string) error {
	if value == "" {
		return nil
	}

	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()

	go func() {
		defer pipeWriter.Close()
		err := json.NewEncoder(pipeWriter).Encode(&document{
			EntityName: entityName,
			Category:   category,
			Value:      value,
			Link:       ref,
		})

		if err != nil {
			plugin.logger.Error("error encoding document:", slog.Any("error", err))
		}
	}()

	docID := fmt.Sprintf("%s#%s", entityName, category)
	docID = url.PathEscape(docID)
	req := esapi.IndexRequest{
		Index:      defaultSearchIndexName,
		DocumentID: docID,
		Body:       pipeReader,
		OpType:     "index",
	}

	res, err := req.Do(ctx, plugin.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		msg, _ := io.ReadAll(res.Body)
		return fmt.Errorf("%w: %s", errIndexCreationFailed, string(msg))
	}

	return nil
}

func (plugin *Search) createIndex(ctx context.Context) error {
	indexSettings := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   plugin.cfg.NumberOfShards,
			"number_of_replicas": plugin.cfg.NumberOfReplicas,
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"value": map[string]interface{}{
					"type": "text",
				},
				"entityName": map[string]interface{}{
					"type": "keyword",
				},
				"category": map[string]interface{}{
					"type": "keyword",
				},
				"link": map[string]interface{}{
					"type": "keyword",
				},
			},
		},
	}

	settingsJSON, err := json.Marshal(indexSettings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	req := esapi.IndicesCreateRequest{
		Index: defaultSearchIndexName,
		Body:  bytes.NewReader(settingsJSON),
	}

	res, err := req.Do(ctx, plugin.client)
	if err != nil {
		return fmt.Errorf("create index request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var errResp errResponse
		if decodeErr := json.NewDecoder(res.Body).Decode(&errResp); decodeErr != nil {
			return fmt.Errorf("error parsing error response: %w", decodeErr)
		}

		// Handle "already exists" case
		if errResp.Error.Type == "resource_already_exists_exception" {
			return nil
		}

		return fmt.Errorf("%w: %s", errIndexCreationFailed, errResp.Error.Reason)
	}

	return nil
}
