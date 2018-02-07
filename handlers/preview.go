package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/previewPage"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/clients/filter"
	"github.com/gorilla/mux"
)

// Submit handles the submitting of a filter job through the filter API
func (f Filter) Submit(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]

	fil, err := f.FilterClient.GetJobState(filterID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	mdl, err := f.FilterClient.UpdateBlueprint(fil, true)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	filterOutputID := mdl.Links.FilterOutputs.ID

	http.Redirect(w, req, fmt.Sprintf("/filter-outputs/%s", filterOutputID), 302)
}

// PreviewPage controls the rendering of the preview and download page
func (f *Filter) PreviewPage(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterOutputID := vars["filterOutputID"]

	datasetCfg := setAuthTokenIfRequired(req)

	fj, err := f.FilterClient.GetOutput(filterOutputID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	prev, err := f.FilterClient.GetPreview(filterOutputID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	var dimensions []filter.ModelDimension
	for _, header := range prev.Headers {
		dimensions = append(dimensions, filter.ModelDimension{Name: header})
	}

	for rowN, row := range prev.Rows {
		if rowN >= 10 {
			break
		}
		for i, val := range row {
			if i < len(dimensions) {
				dimensions[i].Values = append(dimensions[i].Values, val)
			}
		}
	}

	versionURL, err := url.Parse(fj.Links.Version.HRef)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}
	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(versionURL.Path)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	dataset, err := f.DatasetClient.Get(datasetID, datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}
	ver, err := f.DatasetClient.GetVersion(datasetID, edition, version, datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	latestURL, err := url.Parse(dataset.Links.LatestVersion.URL)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	p := mapper.CreatePreviewPage(dimensions, fj, dataset, filterOutputID, datasetID, ver.ReleaseDate)

	if latestURL.Path == versionURL.Path {
		p.Data.IsLatestVersion = true
	}

	metadata, err := f.DatasetClient.GetVersionMetadata(datasetID, edition, version, datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	dims, err := f.DatasetClient.GetDimensions(datasetID, edition, version, datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	size, err := f.getMetadataTextSize(datasetID, edition, version, metadata, dims, datasetCfg)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	p.Data.Downloads = append(p.Data.Downloads, previewPage.Download{
		Extension: "txt",
		Size:      strconv.Itoa(size),
		URI:       fmt.Sprintf("/datasets/%s/editions/%s/versions/%s/metadata.txt", datasetID, edition, version),
	})

	for _, dim := range dims.Items {
		opts, err := f.DatasetClient.GetOptions(datasetID, edition, version, dim.ID, datasetCfg...)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}

		if len(opts.Items) == 1 {
			p.Data.SingleValueDimensions = append(p.Data.SingleValueDimensions, previewPage.Dimension{
				Name:   strings.Title(dim.ID),
				Values: []string{opts.Items[0].Label},
			})
		}
	}

	p.Data.LatestVersion.DatasetLandingPageURL = latestURL.Path
	p.Data.LatestVersion.FilterJourneyWithLatestJourney = fmt.Sprintf("/filters/%s/use-latest-version", filterOutputID)

	if len(p.Data.Dimensions) > 0 {
		p.IsPreviewLoaded = true
	}

	for _, d := range p.Data.Downloads {
		if d.Extension == "xls" && len(d.Size) > 0 {
			p.IsDownloadLoaded = true
		}
	}

	body, err := json.Marshal(p)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	b, err := f.Renderer.Do("dataset-filter/preview-page", body)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	if _, err := w.Write(b); err != nil {
		setStatusCode(req, w, err)
		return
	}
}

func (f *Filter) GetFilterJob(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterOutputID := vars["filterOutputID"]

	prev, err := f.FilterClient.GetOutput(filterOutputID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	b, err := json.Marshal(prev)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	w.Write(b)
}

func (f *Filter) getMetadataTextSize(datasetID, edition, version string, metadata dataset.Metadata, dimensions dataset.Dimensions, cfg []dataset.Config) (int, error) {
	var b bytes.Buffer

	b.WriteString(metadata.String())
	b.WriteString("Dimensions:\n")
	for _, dimension := range dimensions.Items {
		options, err := f.DatasetClient.GetOptions(datasetID, edition, version, dimension.ID, cfg...)
		if err != nil {
			return 0, err
		}

		b.WriteString(options.String())
	}
	return len(b.Bytes()), nil
}
