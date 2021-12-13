package main

import (
	"os"

	"github.com/paketo-buildpacks/mri"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

func main() {
	logger := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))

	packit.Run(
		mri.Detect(mri.NewBuildpackYMLParser()),
		mri.Build(
			draft.NewPlanner(),
			postal.NewService(cargo.NewTransport()),
			logger,
			chronos.DefaultClock,
			pexec.NewExecutable("gem"),
		),
	)
}
