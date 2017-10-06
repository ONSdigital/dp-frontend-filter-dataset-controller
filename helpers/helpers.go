package helpers

import (
	"fmt"
	"regexp"

	"github.com/ONSdigital/go-ns/log"
)

// ExtractDatasetInfoFromPath gets the datasetID, edition and version from a given path
func ExtractDatasetInfoFromPath(path string) (datasetID, edition, version string, err error) {
	log.Debug("attempting to extract dataset details from path", log.Data{"path": path})
	pathReg := regexp.MustCompile(`\/datasets\/(.+)\/editions\/(.+)\/versions\/(.+)`)
	subs := pathReg.FindStringSubmatch(path)
	if len(subs) < 4 {
		err = fmt.Errorf("unabled to extract datasetID, edition and version from path: %s", path)
		return
	}
	return subs[1], subs[2], subs[3], nil
}
