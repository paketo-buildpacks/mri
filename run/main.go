package main

import (
	"os"

	"github.com/paketo-buildpacks/mri"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/draft"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
)

func main() {
	packit.Run(
		mri.Detect(mri.NewBuildpackYMLParser()),
		mri.Build(
			draft.NewPlanner(),
			postal.NewService(cargo.NewTransport()),
			scribe.NewEmitter(os.Stdout),
			chronos.DefaultClock,
			pexec.NewExecutable("gem"),
		),
	)
}
