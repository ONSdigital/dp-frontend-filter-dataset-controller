package helpers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/log.go/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:generate moq -out helperstest/helper.go -pkg helperstest . Handler
type Handler http.Handler

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

// ReturnSecondSegmentFromPath returns the second segment of a path and assumes the path is formed /firstSegment/secondSegment
func ReturnSecondSegmentFromPath(path string) (secondSegment string, err error) {
	subs := strings.Split(path, "/")
	if len(subs) < 3 {
		err = fmt.Errorf("unable to extract secondSegment from path: %s", path)
		return
	}
	return subs[2], nil
}

func CreateReverseProxy(proxyName string, proxyURL *url.URL) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(proxyURL)
	director := proxy.Director
	proxy.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       180 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	proxy.Director = func(req *http.Request) {
		log.Info(req.Context(), "proxying request", log.HTTP(req, 0, 0, nil, nil), log.Data{
			"destination": proxyURL,
			"proxy_name":  proxyName,
		})
		otel.GetTextMapPropagator().Inject(req.Context(), propagation.HeaderCarrier(req.Header))
		director(req)
	}
	return proxy
}
