package config

import (
	"os"
)

type CfgReader interface {
	ReadAt(path string, cfg any) error
}

func ReadYamlCfgFile(configLocation string) (CfgReader, error) {
	file, err := os.Open(configLocation)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return NewYamlCfgReader(file)
}
