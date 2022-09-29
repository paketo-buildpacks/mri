package components_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/mri/dependency/retrieval/components"
	"github.com/sclevine/spec"
)

func testGetDeprecationDate(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("GetDate", func() {
		var (
			deprecationDateRetriever components.DeprecationDateRetriever
			server                   *httptest.Server
		)

		it.Before(func() {
			deprecationDateRetriever = components.NewDeprecationDateRetriever()
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method == http.MethodHead {
					http.Error(w, "NotFound", http.StatusNotFound)
					return
				}

				switch req.URL.Path {
				case "/":
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, `
- name: 4.0
  date: 2021-05-01
  eol_date: 2022-05-01

- name: 3.2
  date: 2021-11-01
  eol_date: 2022-11-01

- name: 3.1
  date: 2021-12-25
  eol_date: 2021-11-01

- name: 3.0
  date: 2021-12-25`)
				case "/bad-content":
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, "bad yaml")
				default:
					t.Fatalf("unknown path: %s", req.URL.Path)
				}
			}))
		})

		it.After(func() {
			server.Close()
		})

		it("retrieves deprecation date for version", func() {
			date, err := deprecationDateRetriever.GetDate(server.URL, "3.2")
			Expect(err).To(Not(HaveOccurred()))
			Expect(date).To(Equal("2022-11-01"))
		})

		context("version has no deprecation date", func() {
			it("returns empty string", func() {
				date, err := deprecationDateRetriever.GetDate(server.URL, "3.0")
				Expect(err).To(Not(HaveOccurred()))
				Expect(date).To(Equal(""))
			})
		})

		context("version does not exist in feed", func() {
			it("returns empty string", func() {
				date, err := deprecationDateRetriever.GetDate(server.URL, "1.2.3")
				Expect(err).To(Not(HaveOccurred()))
				Expect(date).To(Equal(""))
			})
		})

		context("failure cases", func() {
			context("feed endpoint cannot be retrieved", func() {
				it("returns an error", func() {
					_, err := deprecationDateRetriever.GetDate("", "3.0")
					Expect(err).To(MatchError(ContainSubstring("unsupported protocol scheme")))
				})
			})

			context("endpoint body cannot be read", func() {
				it("returns an error", func() {
					_, err := deprecationDateRetriever.GetDate(fmt.Sprintf("%s/bad-content", server.URL), "3.0")
					Expect(err).To(MatchError(ContainSubstring("cannot unmarshal")))
				})
			})
		})
	})
}
