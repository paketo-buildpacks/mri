package main

import (
	"os"

	"github.com/paketo-buildpacks/mri"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/postal"
)

func main() {
	buildpackYMLParser := mri.NewBuildpackYMLParser()
	logEmitter := mri.NewLogEmitter(os.Stdout)
	entryResolver := mri.NewPlanEntryResolver(logEmitter)
	dependencyManager := postal.NewService(cargo.NewTransport())
	planRefinery := mri.NewPlanRefinery()
	gem := pexec.NewExecutable("gem")

	packit.Run(
		mri.Detect(buildpackYMLParser),
		mri.Build(
			entryResolver,
			dependencyManager,
			planRefinery,
			logEmitter,
			chronos.DefaultClock,
			gem,
		),
	)
}
