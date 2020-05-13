package main

import (
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-community/mri/mri"
)

func main() {
	buildpackYMLParser := mri.NewBuildpackYMLParser()

	packit.Detect(mri.Detect(buildpackYMLParser))
}
