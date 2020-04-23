package main

import (
	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/mri-cnb/mri"
)

func main() {
	buildpackYMLParser := mri.NewBuildpackYMLParser()

	packit.Detect(mri.Detect(buildpackYMLParser))
}
