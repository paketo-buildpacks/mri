package components

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"

	"github.com/go-enry/go-license-detector/v4/licensedb"
	"github.com/go-enry/go-license-detector/v4/licensedb/filer"
	"github.com/paketo-buildpacks/packit/v2/vacation"
)

type LicenseRetriever struct{}

func NewLicenseRetriever() LicenseRetriever {
	return LicenseRetriever{}
}

func (LicenseRetriever) LookupLicenses(dependencyName, sourceURL string) ([]interface{}, error) {
	// getting the dependency artifact from sourceURL
	url := sourceURL
	resp, err := http.Get(url) // nolint
	if err != nil {
		return []interface{}{}, fmt.Errorf("failed to query url: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return []interface{}{}, fmt.Errorf("failed to query url %s with: status code %d", url, resp.StatusCode)
	}

	// decompressing the dependency artifact
	tempDir, err := os.MkdirTemp("", "destination")
	if err != nil {
		return []interface{}{}, err
	}
	defer os.RemoveAll(tempDir)

	err = defaultDecompress(resp.Body, tempDir, 1)
	if err != nil {
		return []interface{}{}, err
	}

	// scanning artifact for license file
	filer, err := filer.FromDirectory(tempDir)
	if err != nil {
		return []interface{}{}, fmt.Errorf("failed to setup a licensedb filer: %w", err)
	}

	licenses, err := licensedb.Detect(filer)
	// if no licenses are found, just return an empty slice.
	if err != nil {
		if err.Error() != "no license file was found" {
			return []interface{}{}, fmt.Errorf("failed to detect licenses: %w", err)
		}
		return []interface{}{}, nil
	}

	// Only return the license IDs, in alphabetical order
	licenseIDs := []string{}
	for key := range licenses {
		licenseIDs = append(licenseIDs, key)
	}
	sort.Strings(licenseIDs)

	licenseInterface := []interface{}{}
	for _, license := range licenseIDs {
		licenseInterface = append(licenseInterface, license)
	}

	return licenseInterface, nil
}

func defaultDecompress(artifact io.Reader, destination string, stripComponents int) error {
	archive := vacation.NewArchive(artifact)

	err := archive.StripComponents(stripComponents).Decompress(destination)
	if err != nil {
		return fmt.Errorf("failed to decompress source file: %w", err)
	}

	return nil
}
