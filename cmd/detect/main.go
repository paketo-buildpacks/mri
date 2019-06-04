package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"ruby-cnb/ruby"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"gopkg.in/yaml.v2"
)

func main() {
	context, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default detection context: %s", err)
		os.Exit(101)
	}

	code, err := runDetect(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)
}

func runDetect(context detect.Detect) (int, error) {
	buildpackYAMLPath := filepath.Join(context.Application.Root, "buildpack.yml")
	exists, err := helper.FileExists(buildpackYAMLPath)
	if err != nil {
		return detect.FailStatusCode, err
	}

	version := context.BuildPlan[ruby.Dependency].Version
	if exists {
		version, err = readBuildpackYamlVersion(buildpackYAMLPath)
		if err != nil {
			return detect.FailStatusCode, err
		}
	}

	return context.Pass(buildplan.BuildPlan{
		ruby.Dependency: buildplan.Dependency{
			Version:  version,
			Metadata: buildplan.Metadata{"build": true, "launch": true},
		},
	})
}

func readBuildpackYamlVersion(buildpackYAMLPath string) (string, error) {
	buf, err := ioutil.ReadFile(buildpackYAMLPath)
	if err != nil {
		return "", err
	}

	config := struct {
		Ruby struct {
			Version string `yaml:"version"`
		} `yaml:"ruby"`
	}{}
	if err := yaml.Unmarshal(buf, &config); err != nil {
		return "", err
	}

	return config.Ruby.Version, nil
}
