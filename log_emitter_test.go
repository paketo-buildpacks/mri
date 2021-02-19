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
