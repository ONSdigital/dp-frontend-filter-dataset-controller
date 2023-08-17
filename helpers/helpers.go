package helpers

import (
	"context"
	"fmt"
	"net/url"
	"regexp"

	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/log.go/v2/log"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ExtractDatasetInfoFromPath gets the datasetID, edition and version from a given path
func ExtractDatasetInfoFromPath(ctx context.Context, path string) (datasetID, edition, version string, err error) {
	log.Info(ctx, "attempting to extract dataset details from path", log.Data{"path": path})
	pathReg := regexp.MustCompile(`/datasets/(.+)/editions/(.+)/versions/(.+)`)
	subs := pathReg.FindStringSubmatch(path)
	if len(subs) < 4 {
		err = fmt.Errorf("unable to extract datasetID, edition and version from path: %s", path)
		return
	}
	return subs[1], subs[2], subs[3], nil
}

// GetAPIRouterVersion returns the path of the provided url, which corresponds to the api router version
func GetAPIRouterVersion(rawurl string) (string, error) {
	apiRouterURL, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	return apiRouterURL.Path, nil
}

// StringInSlice will check if a string is in a slice and return a corresponding boolean value along with the
// first index it was found at. If not present then it will return false and negative -1
func StringInSlice(str string, slice []string) (int, bool) {
	for i, sliceStr := range slice {
		if sliceStr == str {
			return i, true
		}
	}
	return -1, false
}

// Check each dimension has an option selected
func CheckAllDimensionsHaveAnOption(dims []filter.ModelDimension) (check bool, err error) {
	if len(dims) == 0 {
		err = fmt.Errorf("no dimensions provided: %v", dims)
		return
	}
	check = true
	for i := range dims {
		if len(dims[i].Options) == 0 {
			check = false
		}
	}
	return
}

// TitleCaseStr is a helper function that returns a given string in title case
func TitleCaseStr(input string) string {
	c := cases.Title(language.English, cases.NoLower)
	return c.String(input)
}
