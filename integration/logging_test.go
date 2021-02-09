package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testLogging(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when the buildpack is run with pack build", func() {
		var (
			image occam.Image

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("logs useful information for the user", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "simple_app"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.MRI.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithEnv(map[string]string{
					"BP_MRI_VERSION": "2.7.x",
				}).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
				"  Resolving MRI version",
				"    Candidate version sources (in priority order):",
				"      BP_MRI_VERSION -> \"2.7.x\"",
				"      <unknown>      -> \"*\"",
				"",
				MatchRegexp(`    Selected MRI version \(using BP_MRI_VERSION\): 2\.7\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing MRI 2\.7\.\d+`),
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
				"",
				"  Configuring environment",
				MatchRegexp(fmt.Sprintf(`    GEM_PATH -> "/home/cnb/.gem/ruby/2\.7\.\d+:/layers/%s/mri/lib/ruby/gems/2\.7\.\d+"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
			))
		})

		context("when the app contains a buildpack.yml", func() {
			it("logs that the buildpack.yml is deprecated", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "simple_app"))
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(filepath.Join(source, "buildpack.yml"), []byte(`{ "mri": { "version": "2.7.x" } }`), 0600)
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy("never").
					WithBuildpacks(
						settings.Buildpacks.MRI.Online,
						settings.Buildpacks.BuildPlan.Online,
					).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				Expect(logs).To(ContainLines(
					MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
					"  Resolving MRI version",
					"    Candidate version sources (in priority order):",
					"      buildpack.yml -> \"2.7.x\"",
					"      <unknown>     -> \"*\"",
					"",
					MatchRegexp(`    Selected MRI version \(using buildpack\.yml\): 2\.7\.\d+`),
					"",
					"    WARNING: Setting the MRI version through buildpack.yml will be deprecated soon in MRI Buildpack v2.0.0.",
					"    Please specify the version through the $BP_MRI_VERSION environment variable instead. See README.md for more information.",
					"",
					"  Executing build process",
					MatchRegexp(`    Installing MRI 2\.7\.\d+`),
					MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
					"",
					"  Configuring environment",
					MatchRegexp(fmt.Sprintf(`    GEM_PATH -> "/home/cnb/.gem/ruby/2\.7\.\d+:/layers/%s/mri/lib/ruby/gems/2\.7\.\d+"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
				))
			})
		})
	})
}
