package mri_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitMRI(t *testing.T) {
	suite := spec.New("mri", spec.Report(report.Terminal{}))
	suite("BuildpackYMLParser", testBuildpackYMLParser)
	suite("Detect", testDetect)
	suite("Build", testBuild)
	suite.Run(t)
}
