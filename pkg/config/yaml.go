package config

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/goccy/go-yaml/ast"
	"github.com/iamgoroot/backline/pkg/core"

	"github.com/goccy/go-yaml"
)

var errEnvVarNotFound = fmt.Errorf("environment variable not found")

type yamlCfgReader struct {
	node    ast.Node
	decoder *yaml.Decoder
}

var envPrefix = []byte("env:")

type ref struct {
	Ref string `yaml:"$ref"`
}

func NewYamlCfgReader(reader io.Reader) (CfgReader, error) {
	dec := yaml.NewDecoder(reader,
		yaml.CustomUnmarshaler(makeEnvAwareUnmarshaller[string]()),
		yaml.CustomUnmarshaler(makeEnvAwareUnmarshaller[int]()),
		yaml.CustomUnmarshaler(makeEnvAwareUnmarshaller[bool]()),
	)
	cfg := yamlCfgReader{decoder: dec}
	err := dec.Decode(&cfg.node)

	return cfg, err
}

func (cfg yamlCfgReader) ReadAt(key string, cfgNode any) error {
	path, err := yaml.PathString(key)
	if err != nil {
		return err
	}

	node, err := path.FilterNode(cfg.node)
	if err != nil {
		return err
	}

	if node == nil {
		return core.ConfigurationError(key)
	}

	var reference ref

	err = cfg.decoder.DecodeFromNode(node, &reference)
	if err != nil && reference.Ref != "" {
		return cfg.ReadAt(reference.Ref, cfgNode)
	}

	err = cfg.decoder.DecodeFromNode(node, cfgNode)

	return err
}

// e.g. env:PORT|8080.
func makeEnvAwareUnmarshaller[T any]() func(*T, []byte) error {
	return func(typedVal *T, rawValue []byte) error {
		if bytes.HasPrefix(rawValue, envPrefix) {
			envReference := rawValue[4:]

			var defaultVal []byte

			sptlitIndex := bytes.IndexByte(envReference, '|')
			if sptlitIndex > 0 {
				defaultVal = envReference[sptlitIndex+1:]
				envReference = envReference[:sptlitIndex]
			} else {
				envReference = bytes.TrimSuffix(envReference, []byte("\n"))
			}

			envVal := os.Getenv(string(envReference))
			if envVal != "" {
				return yaml.Unmarshal([]byte(envVal), typedVal)
			}

			if len(defaultVal) != 0 {
				return yaml.Unmarshal(defaultVal, typedVal)
			}

			return fmt.Errorf("%w: %s", errEnvVarNotFound, string(envReference))
		}

		return yaml.Unmarshal(rawValue, typedVal)
	}
}
