package main

import (
	"fmt"
	"os"
	"path/filepath"
	"ruby-cnb/ruby"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/helper"
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
		bpYml := &BuildpackYaml{}
		err = helper.ReadBuildpackYaml(buildpackYAMLPath, bpYml)
		if err != nil {
			return detect.FailStatusCode, err
		}
		version = bpYml.Ruby.Version

	}

	return context.Pass(buildplan.BuildPlan{
		ruby.Dependency: buildplan.Dependency{
			Version:  version,
			Metadata: buildplan.Metadata{"build": true, "launch": true},
		},
	})
}

type BuildpackYaml struct {
	Ruby struct {
		Version string `yaml:"version"`
	} `yaml:"ruby"`
}
