package ruby

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Version string `yaml:"version"`
}

type BuildpackYMLParser struct{}

func NewBuildpackYMLParser() BuildpackYMLParser {
	return BuildpackYMLParser{}
}

func (p BuildpackYMLParser) ParseVersion(path string) (string, error) {
	var buildpack struct {
		Ruby Config `yaml:"ruby"`
	}

	file, err := os.Open(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	defer file.Close()

	if !os.IsNotExist(err) {
		err = yaml.NewDecoder(file).Decode(&buildpack)
		if err != nil {
			return "", err
		}
	}

	return buildpack.Ruby.Version, nil
}
