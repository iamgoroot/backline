package indexer

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	errPathIsEmpty        = fmt.Errorf("path is empty")
	errMapIsEmpty         = fmt.Errorf("pathMap is nil")
	errKeyMissing         = fmt.Errorf("key is missing")
	errNotSupportedFormat = fmt.Errorf("failed to parse spec as YAML or JSON")
)

type runFunc[T any] func([]string, T) error

func handleSpec[T any](spec map[string]any, do runFunc[T], path ...string) error {
	return handleSpecWithRef(spec, spec, do, nil, 0, path...)
}

func handleSpecWithRef[T any](root, pathMap map[string]any, run runFunc[T], pathLog []string, skipLog int, path ...string) error {
	if len(path) == 0 {
		return errPathIsEmpty
	}

	if pathMap == nil {
		return errMapIsEmpty
	}

	key := path[0] // handle wildcard. process every item. useful to travers opanapi paths
	if key == "*" {
		for key, val := range pathMap {
			paths, ok := val.(map[string]any)
			if !ok {
				continue
			}

			err := handleSpecWithRef[T](root, paths, run, append(pathLog, key), skipLog, path[1:]...)
			if err != nil {
				return err
			}
		}
	}

	val, ok := pathMap[key]

	if !ok {
		return fmt.Errorf("%w : %s", errKeyMissing, key)
	}

	if skipLog > 0 {
		skipLog--
	} else {
		pathLog = append(pathLog, key)
	}

	return processVal[T](root, val, run, pathLog, skipLog, path)
}

func processVal[T any](root map[string]any, val any, run runFunc[T], pathLog []string, skipLog int, path []string) error {
	switch value := val.(type) {
	case map[string]any:
		err := handleSpecWithRef[T](root, value, run, pathLog, skipLog, path[1:]...)
		if err != nil {
			return err
		}

		err = handleRef(root, value, run, pathLog, skipLog, path)
		if err != nil {
			return err
		}
	case []any:
		for _, item := range value {
			item, ok := item.(map[string]any)
			if !ok {
				continue
			}

			err := handleSpecWithRef[T](root, item, run, pathLog, skipLog, path[1:]...)
			if err != nil {
				return err
			}
		}
	case T:
		err := run(pathLog, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func handleRef[T any](root, value map[string]any, run runFunc[T], pathLog []string, skipPathLog int, path []string) error {
	if ref, ok := value["$ref"]; ok {
		refStr, ok := ref.(string)
		if !ok || !strings.HasPrefix(refStr, "#/") {
			return nil
		}

		paths := strings.Split(refStr[2:], "/")

		err := handleSpecWithRef[T](root, root, run, pathLog, skipPathLog+len(paths), append(paths, path[1:]...)...)
		if err != nil {
			return err
		}
	}

	return nil
}

func parseYAMLOrJSON(data []byte) (map[string]any, error) {
	var yamlResult map[string]any
	if err := yaml.Unmarshal(data, &yamlResult); err == nil {
		return yamlResult, nil
	}

	var jsonResult map[string]any
	if err := json.Unmarshal(data, &jsonResult); err != nil {
		return nil, fmt.Errorf("%w : %s", errNotSupportedFormat, err)
	}

	return jsonResult, nil
}
