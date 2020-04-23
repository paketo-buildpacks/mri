package main

import (
	"os"
	"time"

	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/pexec"
	"github.com/cloudfoundry/packit/postal"
	"github.com/cloudfoundry/mri-cnb/mri"
)

func main() {
	logEmitter := mri.NewLogEmitter(os.Stdout)
	entryResolver := mri.NewPlanEntryResolver(logEmitter)
	dependencyManager := postal.NewService(cargo.NewTransport())
	planRefinery := mri.NewPlanRefinery()
	clock := mri.NewClock(time.Now)
	gem := pexec.NewExecutable("gem")

	packit.Build(mri.Build(entryResolver, dependencyManager, planRefinery, logEmitter, clock, gem))
}
