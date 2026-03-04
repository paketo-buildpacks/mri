package mri

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type BuildpackYMLParser struct{}

func NewBuildpackYMLParser() BuildpackYMLParser {
	return BuildpackYMLParser{}
}

func (p BuildpackYMLParser) ParseVersion(path string) (string, error) {
	var buildpack struct {
		MRI struct {
			Version string `yaml:"version"`
		} `yaml:"mri"`
	}

	file, err := os.Open(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close file: %v\n", err)
		}
	}()

	if !os.IsNotExist(err) {
		err = yaml.NewDecoder(file).Decode(&buildpack)
		if err != nil {
			return "", err
		}
	}

	return buildpack.MRI.Version, nil
}
