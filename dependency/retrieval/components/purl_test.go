package components_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/mri/dependency/retrieval/components"
	"github.com/sclevine/spec"
)

func testPurlGeneration(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("Generate", func() {
		it("returns a PURL", func() {
			purl := components.GeneratePurl("dependencyName", "dependencyVersion", "dependencySourceSHA", "http://dependencySource")
			Expect(purl).To(Equal("pkg:generic/dependencyName@dependencyVersion?checksum=dependencySourceSHA&download_url=http://dependencySource"))
		})

	})
}
