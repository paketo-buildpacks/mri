package ruby_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitRuby(t *testing.T) {
	suite := spec.New("ruby", spec.Report(report.Terminal{}))
	suite("BuildpackYMLParser", testBuildpackYMLParser)
	suite("Detect", testDetect)
	suite("LogEmitter", testLogEmitter)
	suite("Clock", testClock)
	suite("PlanEntryResolver", testPlanEntryResolver)
	suite("PlanRefinery", testPlanRefinery)
	suite("Build", testBuild)
	suite.Run(t)
}
