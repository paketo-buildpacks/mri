package components_test

import (
	"archive/tar"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/mri/dependency/retrieval/components"
	"github.com/sclevine/spec"
)

func testLicenseRetrieval(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		server           *httptest.Server
		licenseRetriever components.LicenseRetriever
	)

	it.Before(func() {
		var err error
		licenseRetriever = components.NewLicenseRetriever()

		// Set up tar files
		buffer := bytes.NewBuffer(nil)
		tw := tar.NewWriter(buffer)

		Expect(tw.WriteHeader(&tar.Header{Name: "some-dir", Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
		_, err = tw.Write(nil)
		Expect(err).NotTo(HaveOccurred())

		licenseFile := filepath.Join("some-dir", "LICENSE")
		licenseContent, err := os.ReadFile(filepath.Join("testdata", "LICENSE"))
		Expect(err).NotTo(HaveOccurred())

		Expect(tw.WriteHeader(&tar.Header{Name: licenseFile, Mode: 0755, Size: int64(len(licenseContent))})).To(Succeed())
		_, err = tw.Write(licenseContent)
		Expect(err).NotTo(HaveOccurred())

		Expect(tw.Close()).To(Succeed())

		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method == http.MethodHead {
				http.Error(w, "NotFound", http.StatusNotFound)

				return
			}

			switch req.URL.Path {
			case "/":
				w.WriteHeader(http.StatusOK)

			case "/default-dependency-source-url.tgz":
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, buffer.String())

			case "/bad-url":
				w.WriteHeader(http.StatusBadRequest)

			case "/non-tar-file-artifact":
				w.WriteHeader(http.StatusOK)
				// return a flac header, which is an unrecognized mime-type
				fmt.Fprint(w, bytes.NewBuffer([]byte("\x66\x4C\x61\x43\x00\x00\x00\x22")))

			case "/no-license.tgz":
				buffer = bytes.NewBuffer(nil)
				tw = tar.NewWriter(buffer)

				Expect(tw.WriteHeader(&tar.Header{Name: "./", Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
				_, err = tw.Write(nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(tw.Close()).To(Succeed())

				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, buffer.String())

			default:
				t.Fatal(fmt.Sprintf("unknown path: %s", req.URL.Path))
			}
		}))
	})

	it.After(func() {
		server.Close()
	})

	context("given a dependency URL to get the license for", func() {
		it("gets the artifact and retrieves the license from it", func() {
			licenses, err := licenseRetriever.LookupLicenses("dependency", fmt.Sprintf("%s/default-dependency-source-url.tgz", server.URL))
			Expect(err).NotTo(HaveOccurred())
			Expect(licenses).To(Equal([]interface{}{"MIT", "MIT-0"}))
		})
	})

	context("the artifact does not contain a license", func() {
		it("returns an empty slice of licenses and no error", func() {
			licenses, err := licenseRetriever.LookupLicenses("dependency", fmt.Sprintf("%s/no-license.tgz", server.URL))
			Expect(err).ToNot(HaveOccurred())
			Expect(licenses).To(Equal([]interface{}{}))
		})
	})

	context("failure cases", func() {
		context("the request to the source URL fails", func() {
			it("returns an error and exits non-zero", func() {
				_, err := licenseRetriever.LookupLicenses("dependency", "non-existent/url")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring(`failed to query url: Get "non-existent/url"`)))
			})
		})

		context("the status code of the response is not OK", func() {
			it("returns an error and exits non-zero", func() {
				_, err := licenseRetriever.LookupLicenses("dependency", fmt.Sprintf("%s/bad-url", server.URL))
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(`failed to query url %s/bad-url with: status code 400`, server.URL))))
			})
		})

		context("the artifact cannot be decompressed", func() {
			it("returns an error and exits non-zero", func() {
				_, err := licenseRetriever.LookupLicenses("dependency", fmt.Sprintf("%s/non-tar-file-artifact", server.URL))
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("failed to decompress source file")))
			})
		})
	})
}
