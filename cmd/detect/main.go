package main

import (
	"github.com/cloudfoundry/packit"
	"github.com/paketo-community/mri/mri"
)

func main() {
	buildpackYMLParser := mri.NewBuildpackYMLParser()

	packit.Detect(mri.Detect(buildpackYMLParser))
}
