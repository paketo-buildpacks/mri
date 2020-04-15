package main

import (
	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/ruby-mri-cnb/ruby"
)

func main() {
	buildpackYMLParser := ruby.NewBuildpackYMLParser()

	packit.Detect(ruby.Detect(buildpackYMLParser))
}
