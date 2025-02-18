package fs

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/model"
)

var _ core.Discovery = &Discovery{}

const configKey = "$.locations.dir"

type Discovery struct {
	core.NoOpShutdown
	locations []string
}

func (d *Discovery) Setup(_ context.Context, _ core.Dependencies) error { return nil }

func (d *Discovery) Discover(_ context.Context, deps core.Dependencies, register core.RegistrationFunc) error {
	err := deps.CfgReader().ReadAt(configKey, &d.locations)
	if err != nil {
		deps.Logger().Error("missing config for entity discovery on file system. skipping discovery", slog.Any("err", err))
		return nil
	}

	for _, location := range d.locations {
		discoverErr := discover(location, register)
		err = errors.Join(err, discoverErr)
	}

	return err
}

func (d *Discovery) TryDownload(_ context.Context, _ core.Dependencies, locMeta *model.LocationMetadata, ref string) (string, error) {
	loc := path.Join(locMeta.Location, ref)

	data, err := os.ReadFile(loc)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func discover(location string, register core.RegistrationFunc) error {
	stat, err := os.Stat(location)
	if err != nil {
		return err
	}

	if !stat.IsDir() {
		return readFile(location, register)
	}

	dir, err := os.ReadDir(location)
	if err != nil {
		return err
	}

	for _, dirEntry := range dir {
		subLocation := path.Join(location, dirEntry.Name())
		discoverErr := discover(subLocation, register)
		err = errors.Join(err, discoverErr)
	}

	return err
}

func readFile(location string, register core.RegistrationFunc) error {
	file, err := os.Open(location)
	if err != nil {
		return nil
	}

	defer file.Close()

	locationMeta := &model.LocationMetadata{
		DownloadedBy: "dir-discovery",
		Location:     location,
	}

	return ReadSpecs(file, locationMeta, register)
}
