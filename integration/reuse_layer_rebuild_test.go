package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testReusingLayerRebuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		docker occam.Docker
		pack   occam.Pack

		imageIDs     map[string]struct{}
		containerIDs map[string]struct{}

		name   string
		source string
	)

	it.Before(func() {
		var err error
		name, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())

		docker = occam.NewDocker()
		pack = occam.NewPack().WithNoColor()
		imageIDs = map[string]struct{}{}
		containerIDs = map[string]struct{}{}
	})

	it.After(func() {
		for id := range containerIDs {
			Expect(docker.Container.Remove.Execute(id)).To(Succeed())
		}

		for id := range imageIDs {
			Expect(docker.Image.Remove.Execute(id)).To(Succeed())
		}

		Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
		Expect(os.RemoveAll(source)).To(Succeed())
	})

	context("when an app is rebuilt and does not change", func() {
		it("reuses a layer from a previous build", func() {
			var (
				err         error
				logs        fmt.Stringer
				firstImage  occam.Image
				secondImage occam.Image

				firstContainer  occam.Container
				secondContainer occam.Container
			)

			source, err = occam.Source(filepath.Join("testdata", "simple_app"))
			Expect(err).NotTo(HaveOccurred())

			build := pack.Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.MRI.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithEnv(map[string]string{"BP_MRI_VERSION": "3.4.x"})

			firstImage, logs, err = build.Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String)

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(2))
			Expect(firstImage.Buildpacks[0].Key).To(Equal(settings.Buildpack.ID))
			Expect(firstImage.Buildpacks[0].Layers).To(HaveKey("mri"))

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
				"  Resolving MRI version",
				"    Candidate version sources (in priority order):",
				"      BP_MRI_VERSION -> \"3.4.x\"",
				"      <unknown>      -> \"\"",
			))

			Expect(logs).To(ContainLines(
				MatchRegexp(`    Selected MRI version \(using BP_MRI_VERSION\): 3\.4\.\d+`),
			))

			Expect(logs).To(ContainLines(
				"  Executing build process",
				MatchRegexp(`    Installing MRI 3\.\d+\.\d+`),
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
			))

			Expect(logs).To(ContainLines(
				"  Configuring build environment",
				MatchRegexp(fmt.Sprintf(`    GEM_PATH         -> "/home/cnb/.local/share/gem/ruby/3\.4\.\d+:/layers/%s/mri/lib/ruby/gems/3\.4\.\d+"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
				`    MALLOC_ARENA_MAX -> "2"`,
			))

			Expect(logs).To(ContainLines(
				"  Configuring launch environment",
				MatchRegexp(fmt.Sprintf(`    GEM_PATH         -> "/home/cnb/.local/share/gem/ruby/3\.4\.\d+:/layers/%s/mri/lib/ruby/gems/3\.4\.\d+"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
				`    MALLOC_ARENA_MAX -> "2"`,
			))

			firstContainer, err = docker.Container.Run.
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				WithPublishAll().
				WithCommand("ruby run.rb").
				Execute(firstImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			Eventually(firstContainer).Should(BeAvailable())

			// Second pack build
			secondImage, logs, err = build.Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(2))
			Expect(secondImage.Buildpacks[0].Key).To(Equal(settings.Buildpack.ID))
			Expect(secondImage.Buildpacks[0].Layers).To(HaveKey("mri"))

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
				"  Resolving MRI version",
				"    Candidate version sources (in priority order):",
				"      BP_MRI_VERSION -> \"3.4.x\"",
				"      <unknown>      -> \"\"",
			))

			Expect(logs).To(ContainLines(
				MatchRegexp(`    Selected MRI version \(using BP_MRI_VERSION\): 3\.4\.\d+`),
			))

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf("  Reusing cached layer /layers/%s/mri", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
			))

			secondContainer, err = docker.Container.Run.
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				WithPublishAll().
				WithCommand("ruby run.rb").
				Execute(secondImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(secondContainer).Should(BeAvailable())
			Eventually(secondContainer).Should(Serve(MatchRegexp(`Hello from Ruby 3\.4\.\d+`)).OnPort(8080))

			Expect(secondImage.Buildpacks[0].Layers["mri"].SHA).To(Equal(firstImage.Buildpacks[0].Layers["mri"].SHA))
		})
	})

	context("when an app is rebuilt and there is a change", func() {
		it("rebuilds the layer", func() {
			var (
				err         error
				logs        fmt.Stringer
				firstImage  occam.Image
				secondImage occam.Image

				firstContainer  occam.Container
				secondContainer occam.Container
			)

			source, err = occam.Source(filepath.Join("testdata", "simple_app"))
			Expect(err).NotTo(HaveOccurred())

			build := pack.Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.MRI.Online,
					settings.Buildpacks.BuildPlan.Online,
				)

			firstImage, logs, err = build.
				WithEnv(map[string]string{"BP_MRI_VERSION": "3.4.x"}).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String)

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(2))
			Expect(firstImage.Buildpacks[0].Key).To(Equal(settings.Buildpack.ID))
			Expect(firstImage.Buildpacks[0].Layers).To(HaveKey("mri"))

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
				"  Resolving MRI version",
				"    Candidate version sources (in priority order):",
				"      BP_MRI_VERSION -> \"3.4.x\"",
				"      <unknown>      -> \"\"",
			))

			Expect(logs).To(ContainLines(
				MatchRegexp(`    Selected MRI version \(using BP_MRI_VERSION\): 3\.4\.\d+`),
			))

			Expect(logs).To(ContainLines(
				"  Executing build process",
				MatchRegexp(`    Installing MRI 3\.4\.\d+`),
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
			))

			Expect(logs).To(ContainLines(
				"  Configuring build environment",
				MatchRegexp(fmt.Sprintf(`    GEM_PATH         -> "/home/cnb/.local/share/gem/ruby/3\.4\.\d+:/layers/%s/mri/lib/ruby/gems/3\.4\.\d+"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
				`    MALLOC_ARENA_MAX -> "2"`,
			))

			Expect(logs).To(ContainLines(
				"  Configuring launch environment",
				MatchRegexp(fmt.Sprintf(`    GEM_PATH         -> "/home/cnb/.local/share/gem/ruby/3\.4\.\d+:/layers/%s/mri/lib/ruby/gems/3\.4\.\d+"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
				`    MALLOC_ARENA_MAX -> "2"`,
			))

			firstContainer, err = docker.Container.Run.
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				WithPublishAll().
				WithCommand("ruby run.rb").Execute(firstImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			Eventually(firstContainer).Should(BeAvailable())

			// Second pack build
			secondImage, logs, err = build.
				WithEnv(map[string]string{"BP_MRI_VERSION": "3.2.x"}).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(2))
			Expect(secondImage.Buildpacks[0].Key).To(Equal(settings.Buildpack.ID))
			Expect(secondImage.Buildpacks[0].Layers).To(HaveKey("mri"))

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
				"  Resolving MRI version",
				"    Candidate version sources (in priority order):",
				"      BP_MRI_VERSION -> \"3.2.x\"",
				"      <unknown>      -> \"\"",
			))

			Expect(logs).To(ContainLines(
				MatchRegexp(`    Selected MRI version \(using BP_MRI_VERSION\): 3\.2\.\d+`),
			))

			Expect(logs).To(ContainLines(
				"  Executing build process",
				MatchRegexp(`    Installing MRI 3\.2\.\d+`),
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
			))

			Expect(logs).To(ContainLines(
				"  Configuring build environment",
				MatchRegexp(fmt.Sprintf(`    GEM_PATH         -> "/home/cnb/.local/share/gem/ruby/3\.2\.\d+:/layers/%s/mri/lib/ruby/gems/3\.2\.\d+"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
				`    MALLOC_ARENA_MAX -> "2"`,
			))

			Expect(logs).To(ContainLines(
				"  Configuring launch environment",
				MatchRegexp(fmt.Sprintf(`    GEM_PATH         -> "/home/cnb/.local/share/gem/ruby/3\.2\.\d+:/layers/%s/mri/lib/ruby/gems/3\.2\.\d+"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
				`    MALLOC_ARENA_MAX -> "2"`,
			))

			secondContainer, err = docker.Container.Run.
				WithCommand("ruby run.rb").
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				WithPublishAll().
				Execute(secondImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(secondContainer).Should(BeAvailable())
			Eventually(secondContainer).Should(Serve(MatchRegexp(`Hello from Ruby 3\.2\.\d+`)).OnPort(8080))

			Expect(secondImage.Buildpacks[0].Layers["mri"].SHA).NotTo(Equal(firstImage.Buildpacks[0].Layers["mri"].SHA))
		})
	})
}
