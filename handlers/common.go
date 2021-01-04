package handlers

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/filter"
)

// MaxNumOptionsOnPage is the maximum number of options that will be presented on a screen.
// If more options need to be presented, then the hierarchy will be used, if possible.
const MaxNumOptionsOnPage = 20

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
// getting the labels for the provided IDs from DatasetAPI.
// Note that this method may be expensive if lots of filterOptions are provided, if you can get the labels from some other available source, it would be preferred.
func (f *Filter) getIDNameLookupFromDatasetAPI(ctx context.Context, userAccessToken, collectionID, datasetID, edition, version, name string,
	filterOptions filter.DimensionOptions) (idLabelMap map[string]string, err error) {

	// if no items are provided, return straight away (nothing to map)
	if filterOptions.Items == nil {
		return map[string]string{}, nil
	}

	// initialise map of options to find, allocating empty strings for all values.
	idLabelMap = make(map[string]string, len(filterOptions.Items))
	for _, option := range filterOptions.Items {
		idLabelMap[option.Option] = ""
	}

	// call datasetAPI GetOptions by IDs in batches until we find all required values
	offset := 0
	foundCount := 0
	for offset < len(filterOptions.Items) {
		// get batch of option IDs to obtain
		batchOpts := []string{}
		bachEnd := min(len(filterOptions.Items), offset+f.maxDatasetOptions)
		for _, opt := range filterOptions.Items[offset:bachEnd] {
			batchOpts = append(batchOpts, opt.Option)
		}

		// get options batch from dataset API
		options, err := f.DatasetClient.GetOptions(ctx, userAccessToken, "", collectionID, datasetID, edition, version, name, dataset.QueryParams{IDs: batchOpts})
		if err != nil {
			return idLabelMap, err
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
		offset += f.maxDatasetOptions
	}

	// return error because some value(s) could not be found
	return idLabelMap, errors.New("could not find all required filter options in dataset API")
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
