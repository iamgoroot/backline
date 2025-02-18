package fs

import (
	"bufio"
	"bytes"
	"errors"
	"io"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/model"
	"gopkg.in/yaml.v3"
)

func ReadSpecs(reader io.Reader, meta *model.LocationMetadata, register core.RegistrationFunc) error {
	data := &bytes.Buffer{}
	scanner := bufio.NewScanner(reader)

	var errs error

	for scanner.Scan() {
		txt := scanner.Text()
		if txt == "---" {
			entity, err := parse(data)
			if err == nil {
				register(meta, entity)
			}

			errs = errors.Join(errs, err)

			data.Reset()

			continue
		}

		data.WriteString(txt)
		data.WriteByte('\n')
	}

	entity, err := parse(data)
	if err == nil {
		register(meta, entity)
	}

	return errors.Join(errs, err, scanner.Err())
}

func parse(data *bytes.Buffer) (*model.Entity, error) {
	entity := &model.Entity{}
	err := yaml.Unmarshal(data.Bytes(), entity)

	return entity, err
}
