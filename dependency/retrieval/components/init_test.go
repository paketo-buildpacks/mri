package components_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnit(t *testing.T) {
	suite := spec.New("retrieval-components", spec.Report(report.Terminal{}), spec.Parallel())
	suite("ReleaseFetcher", testReleaseFetcher)
	suite("FindNewVersions", testFindNewVersions)
	suite("MetadataGeneration", testMetadataGeneration)
	suite("DeprecationDate", testGetDeprecationDate)
	suite("LicenseRetrieval", testLicenseRetrieval)
	suite("PurlGeneration", testPurlGeneration)
	suite("DependencyValidation", testDependencyValidation)
	suite.Run(t)
}
