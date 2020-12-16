package handlers

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/filter"
)

// these form vars are not regular input fields, but transmit meta form info
var specialFormVars = map[string]bool{
	"save-and-return": true,
	":uri":            true,
	"q":               true,
}

// getOptionsAndRedirect iterates the provided form values and creates a list of options
// and updates a redirectURI if the form contains a redirect.
func getOptionsAndRedirect(form url.Values, redirectURI *string) (options []string) {
	options = []string{}
	for k := range form {
		if _, foundSpecial := specialFormVars[k]; foundSpecial {
			continue
		}

		if strings.Contains(k, "redirect:") {
			redirectReg := regexp.MustCompile(`^redirect:(.+)$`)
			redirectSubs := redirectReg.FindStringSubmatch(k)
			*redirectURI = redirectSubs[1]
			continue
		}

		options = append(options, k)
	}
	return options
}

// getIDNameLookupFromDatasetAPI creates a map of option keys and labels from the provided filter options,
// getting the labels from DatasetAPI with paginated calls until all values have been found.
// Note that this method may be expensive, if you can get the labels from some other available source, it would be preferred.
func (f *Filter) getIDNameLookupFromDatasetAPI(ctx context.Context, userAccessToken, collectionID, datasetID, edition, version, name string,
	filterOptions filter.DimensionOptions) (idLabelMap map[string]string, err error) {

	// initialise map of options to find, allocating empty strings for all values.
	idLabelMap = make(map[string]string, len(filterOptions.Items))
	for _, option := range filterOptions.Items {
		idLabelMap[option.Option] = ""
	}

	// call datasetAPI GetOptions with pagination until we find all values
	offset := 0
	totalCount := 1
	foundCount := 0
	for offset < totalCount {
		// get options batch from dataset API
		options, err := f.DatasetClient.GetOptions(ctx, userAccessToken, "", collectionID, datasetID, edition, version, name, offset, f.BatchSize)
		if err != nil {
			return idLabelMap, err
		}

		// (first iteration only) - set totalCount
		if offset == 0 {
			totalCount = options.TotalCount
		}

		// iterate items in batch and populate labels in idLabelMap
		for _, opt := range options.Items {
			if _, found := idLabelMap[opt.Option]; found {
				idLabelMap[opt.Option] = opt.Label
				foundCount++
			}
		}

		// return if we have found all the required options
		if foundCount == len(idLabelMap) {
			return idLabelMap, nil
		}

		// set offset for the next iteration
		offset += f.BatchSize
	}

	// return error because some value(s) could not be found
	return idLabelMap, errors.New("could not find all required filter options in dataset API")
}
