package helpers

import (
	"context"
	"fmt"
	"net/url"
	"regexp"

	"github.com/ONSdigital/log.go/log"
)

// ExtractDatasetInfoFromPath gets the datasetID, edition and version from a given path
func ExtractDatasetInfoFromPath(ctx context.Context, path string) (datasetID, edition, version string, err error) {
	log.Event(ctx, "attempting to extract dataset details from path", log.INFO, log.Data{"path": path})
	pathReg := regexp.MustCompile(`\/datasets\/(.+)\/editions\/(.+)\/versions\/(.+)`)
	subs := pathReg.FindStringSubmatch(path)
	if len(subs) < 4 {
		err = fmt.Errorf("unabled to extract datasetID, edition and version from path: %s", path)
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

// StringInSlice will check if a string is in a slice and return a corresponding boolean value
func StringInSlice(str string, slice []string) bool {
	for _, sliceStr := range slice {
		if sliceStr == str {
			return true
		}
	}
	return false
}
