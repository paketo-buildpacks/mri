package mri_test

import (
	"bytes"
	"testing"

	"github.com/paketo-buildpacks/mri"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testLogEmitter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer  *bytes.Buffer
		emitter mri.LogEmitter
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
		emitter = mri.NewLogEmitter(buffer)
	})

	context("Candidates", func() {
		it("prints a formatted map of version source inputs", func() {
			emitter.Candidates([]packit.BuildpackPlanEntry{
				{
					Name: "mri",
					Metadata: map[string]interface{}{
						"version-source": "package.json",
						"version":        "package-json-version",
					},
				},
				{
					Name: "mri",
					Metadata: map[string]interface{}{
						"version": "other-version",
					},
				},
				{
					Name: "mri",
					Metadata: map[string]interface{}{
						"version-source": "buildpack.yml",
						"version":        "buildpack-yml-version",
					},
				},
				{
					Name: "mri",
				},
				{
					Name: "mri",
					Metadata: map[string]interface{}{
						"version-source": "BP_MRI_VERSION",
						"version":        "env-var-version",
					},
				},
			})

			Expect(buffer.String()).To(ContainSubstring("    Candidate version sources (in priority order):"))
			Expect(buffer.String()).To(ContainSubstring("BP_MRI_VERSION -> \"env-var-version\""))
			Expect(buffer.String()).To(ContainSubstring("buildpack.yml  -> \"buildpack-yml-version\""))
			Expect(buffer.String()).To(ContainSubstring("<unknown>      -> \"other-version\""))
			Expect(buffer.String()).To(ContainSubstring("<unknown>      -> \"*\""))
		})
	})

	context("Environment", func() {
		it("prints details about the environment", func() {
			emitter.Environment(packit.Environment{
				"GEM_PATH.override": "/some/path",
			})

			Expect(buffer.String()).To(ContainSubstring("  Configuring environment"))
			Expect(buffer.String()).To(ContainSubstring("    GEM_PATH -> \"/some/path\""))
		})
	})
}
