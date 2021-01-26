package handlers

import (
	"context"
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
// concurrently getting the labels for the provided IDs from DatasetAPI.
// Note that this method may be expensive if lots of filterOptions are provided, if you can get the labels from some other available source, it would be preferred.
func (f *Filter) getIDNameLookupFromDatasetAPI(ctx context.Context, userAccessToken, collectionID, datasetID, edition, version, name string,
	filterOptions filter.DimensionOptions) (idLabelMap map[string]string, err error) {

	// if no items are provided, return straight away (nothing to map)
	if filterOptions.Items == nil || len(filterOptions.Items) == 0 {
		return map[string]string{}, nil
	}

	// generate the complete list of IDs that will need to be requested
	optionIDs := make([]string, len(filterOptions.Items))
	for i, opt := range filterOptions.Items {
		optionIDs[i] = url.QueryEscape(opt.Option)
	}

	// initialise map of options to find and a batch processor that will map option IDs to Labels
	idLabelMap = make(map[string]string, len(filterOptions.Items))
	processBatch := generateOptionMapperBatchProcessor(&idLabelMap)

	// call dataset API GetOptionsBatchProcess with the batch processor
	err = f.DatasetClient.GetOptionsBatchProcess(ctx, userAccessToken, "", collectionID, datasetID, edition, version, name, &optionIDs, processBatch, f.maxDatasetOptions, f.BatchMaxWorkers)
	return idLabelMap, err
}

// generateOptionMapperBatchProcessor maps the batch option IDs to labels into the provided idLabelMap
var generateOptionMapperBatchProcessor = func(idLabelMap *map[string]string) dataset.OptionsBatchProcessor {
	return func(batch dataset.Options) (forceAbort bool, err error) {
		for _, opt := range batch.Items {
			(*idLabelMap)[opt.Option] = opt.Label
		}
		return false, nil
	}
}
