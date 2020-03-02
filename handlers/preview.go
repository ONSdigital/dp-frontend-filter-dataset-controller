package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"strings"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-api-clients-go/headers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/previewPage"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

// Submit handles the submitting of a filter job through the filter API
func (f Filter) Submit(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]
	ctx := req.Context()

	collectionID := getCollectionIDFromContext(ctx)
	userAccessToken, err := headers.GetUserAuthToken(req)
	if err != nil {
		if headers.IsNotErrNotFound(err) {
			log.Event(ctx, "error getting access token header", log.Error(err))
		}
	}

	fil, err := f.FilterClient.GetJobState(req.Context(), userAccessToken, "", "", collectionID, filterID)
	if err != nil {
		log.Event(ctx, "failed to get job state", log.Error(err), log.Data{"filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	mdl, err := f.FilterClient.UpdateBlueprint(req.Context(), userAccessToken, "", "", collectionID, fil, true)
	if err != nil {
		log.Event(ctx, "failed to submit filter blueprint", log.Error(err), log.Data{"filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	filterOutputID := mdl.Links.FilterOutputs.ID

	http.Redirect(w, req, fmt.Sprintf("/filter-outputs/%s", filterOutputID), 302)
}

// OutputPage controls the rendering of the preview and download page
func (f *Filter) OutputPage(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterOutputID := vars["filterOutputID"]
	ctx := req.Context()

	collectionID := getCollectionIDFromContext(ctx)
	userAccessToken, err := headers.GetUserAuthToken(req)
	if err != nil {
		if headers.IsNotErrNotFound(err) {
			log.Event(ctx, "error getting access token header", log.Error(err))
		}
	}

	fj, err := f.FilterClient.GetOutput(req.Context(), userAccessToken, "", "", collectionID, filterOutputID)
	if err != nil {
		log.Event(ctx, "failed to get filter output", log.Error(err), log.Data{"filter_output_id": filterOutputID})
		setStatusCode(req, w, err)
		return
	}

	filterID := fj.Links.FilterBlueprint.ID

	dimensions := make([]filter.ModelDimension, 0)
	if f.EnableDatasetPreview {
		prev, err := f.FilterClient.GetPreview(req.Context(), userAccessToken, "", "", collectionID, filterOutputID)
		if err != nil {
			log.Event(ctx, "failed to get preview", log.Error(err), log.Data{"filter_output_id": filterOutputID})
			setStatusCode(req, w, err)
			return
		}

		if len(prev.Headers[0]) < 4 || strings.ToUpper(prev.Headers[0][0:3]) != "V4_" {
			err = errors.New("Unexpected format - expected `V4_N` in header")
			log.Event(ctx, "failed to format header", log.Error(err), log.Data{"filter_output_id": filterOutputID, "header": prev.Headers})
			setStatusCode(req, w, err)
			return
		}

		markingsColumnCount, err := strconv.Atoi(prev.Headers[0][3:])
		if err != nil {
			log.Event(ctx, "failed to get column count from header cell", log.Error(err), log.Data{"filter_output_id": filterOutputID, "header": prev.Headers[0]})
			setStatusCode(req, w, err)
			return
		}

		indexOfFirstLabelColumn := markingsColumnCount + 2 // +1 for observation, +1 for first codelist column
		dimensions = []filter.ModelDimension{filter.ModelDimension{Name: "Values"}}
		// add markings column headers
		for i := 1; i <= markingsColumnCount; i++ {
			dimensions = append(dimensions, filter.ModelDimension{Name: prev.Headers[i]})
		}
		// add non-codelist column headers
		for i := indexOfFirstLabelColumn; i < len(prev.Headers); i += 2 {
			dimensions = append(dimensions, filter.ModelDimension{Name: prev.Headers[i]})
		}

		for rowN, row := range prev.Rows {
			if rowN >= 10 {
				break
			}
			if len(row) > 0 {
				// add observation[0]+markings[1:markingsColumnCount+1] columns of row
				for i := 0; i <= markingsColumnCount; i++ {
					dimensions[i].Values = append(dimensions[i].Values, row[i])
				}
				// add non-codelist[indexOfFirstLabelColumn:/2] columns of row
				dimIndex := markingsColumnCount + 1
				for i := indexOfFirstLabelColumn; i < len(row); i += 2 {
					dimensions[dimIndex].Values = append(dimensions[dimIndex].Values, row[i])
					dimIndex++
				}
			}
		}
	}

	versionURL, err := url.Parse(fj.Links.Version.HRef)
	if err != nil {
		log.Event(ctx, "failed to parse version href", log.Error(err), log.Data{"filter_output_id": filterOutputID})
		setStatusCode(req, w, err)
		return
	}
	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(ctx, versionURL.Path)
	if err != nil {
		log.Event(ctx, "failed to extract dataset info from path", log.Error(err), log.Data{"filter_output_id": filterOutputID, "path": versionURL})
		setStatusCode(req, w, err)
		return
	}

	dataset, err := f.DatasetClient.Get(req.Context(), userAccessToken, "", collectionID, datasetID)
	if err != nil {
		log.Event(ctx, "failed to get dataset", log.Error(err), log.Data{"dataset_id": datasetID})
		setStatusCode(req, w, err)
		return
	}
	ver, err := f.DatasetClient.GetVersion(req.Context(), userAccessToken, "", "", collectionID, datasetID, edition, version)
	if err != nil {
		log.Event(ctx, "failed to get version", log.Error(err), log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	latestURL, err := url.Parse(dataset.Links.LatestVersion.URL)
	if err != nil {
		log.Event(ctx, "failed to parse latest version href", log.Error(err), log.Data{"filter_output_id": filterOutputID})
		setStatusCode(req, w, err)
		return
	}

	p := mapper.CreatePreviewPage(req.Context(), req, dimensions, fj, dataset, filterOutputID, datasetID, ver.ReleaseDate, f.EnableDatasetPreview, f.EnableLoop11)

	if latestURL.Path == versionURL.Path {
		p.Data.IsLatestVersion = true
	}

	metadata, err := f.DatasetClient.GetVersionMetadata(req.Context(), userAccessToken, "", collectionID, datasetID, edition, version)
	if err != nil {
		log.Event(ctx, "failed to get version metadata", log.Error(err), log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	dims, err := f.DatasetClient.GetDimensions(req.Context(), userAccessToken, "", collectionID, datasetID, edition, version)
	if err != nil {
		log.Event(ctx, "failed to get dimensions", log.Error(err), log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	size, err := f.getMetadataTextSize(req.Context(), userAccessToken, datasetID, edition, version, metadata, dims)
	if err != nil {
		log.Event(ctx, "failed to get metadata text size", log.Error(err), log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	for _, dim := range dims.Items {
		opts, err := f.DatasetClient.GetOptions(req.Context(), userAccessToken, "", collectionID, datasetID, edition, version, dim.Name)
		if err != nil {
			log.Event(ctx, "failed to get options from dataset client", log.Error(err), log.Data{"dimension": dim.Name, "dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		if len(opts.Items) == 1 {
			p.Data.SingleValueDimensions = append(p.Data.SingleValueDimensions, previewPage.Dimension{
				Name:   strings.Title(dim.Name),
				Values: []string{opts.Items[0].Label},
			})
		}
	}

	p.Data.LatestVersion.DatasetLandingPageURL = latestURL.Path
	p.Data.LatestVersion.FilterJourneyWithLatestJourney = fmt.Sprintf("/filters/%s/use-latest-version", filterID)

	if len(p.Data.Dimensions) > 0 {
		p.IsPreviewLoaded = true
	}

	for i, d := range p.Data.Downloads {
		if d.Extension == "xls" && (len(d.Size) > 0 || d.Skipped) {
			p.IsDownloadLoaded = true
		}

		if len(f.downloadServiceURL) > 0 {
			downloadURL, err := url.Parse(d.URI)
			if err != nil {
				setStatusCode(req, w, err)
				return
			}

			d.URI = f.downloadServiceURL + downloadURL.Path
			p.Data.Downloads[i] = d
		}
	}

	// Text file is created on the fly in this app, so do not prepend the
	// download service url as is the case with other downloads
	p.Data.Downloads = append(p.Data.Downloads, previewPage.Download{
		Extension: "txt",
		Size:      strconv.Itoa(size),
		URI:       fmt.Sprintf("/datasets/%s/editions/%s/versions/%s/metadata.txt", datasetID, edition, version),
	})

	body, err := json.Marshal(p)
	if err != nil {
		log.Event(ctx, "failed to marshal json", log.Error(err), log.Data{"filter_output_id": filterOutputID})
		setStatusCode(req, w, err)
		return
	}

	b, err := f.Renderer.Do("dataset-filter/preview-page", body)
	if err != nil {
		log.Event(ctx, "failed to render", log.Error(err), log.Data{"filter_output_id": filterOutputID})
		setStatusCode(req, w, err)
		return
	}

	if _, err := w.Write(b); err != nil {
		log.Event(ctx, "failed to write response", log.Error(err), log.Data{"filter_output_id": filterOutputID})
		setStatusCode(req, w, err)
		return
	}
}

// GetFilterJob returns the filter output json to the client to form preview
// for AJAX request
func (f *Filter) GetFilterJob(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterOutputID := vars["filterOutputID"]
	ctx := req.Context()

	collectionID := getCollectionIDFromContext(ctx)
	userAccessToken, err := headers.GetUserAuthToken(req)
	if err != nil {
		if headers.IsNotErrNotFound(err) {
			log.Event(ctx, "error getting access token header", log.Error(err))
		}
	}
	prev, err := f.FilterClient.GetOutput(req.Context(), userAccessToken, "", "", collectionID, filterOutputID)
	if err != nil {
		log.Event(ctx, "failed to get filter output", log.Error(err), log.Data{"filter_output_id": filterOutputID})
		setStatusCode(req, w, err)
		return
	}

	for k, download := range prev.Downloads {
		if len(download.URL) == 0 {
			continue
		}
		downloadURL, err := url.Parse(download.URL)
		if err != nil {
			log.Event(ctx, "failed to parse download url", log.Error(err), log.Data{"filter_output_id": filterOutputID})
			setStatusCode(req, w, err)
			return
		}

		download.URL = f.downloadServiceURL + downloadURL.Path
		prev.Downloads[k] = download
	}

	b, err := json.Marshal(prev)
	if err != nil {
		log.Event(ctx, "failed to marshal json", log.Error(err), log.Data{"filter_output_id": filterOutputID})
		setStatusCode(req, w, err)
		return
	}

	w.Write(b)
}

func (f *Filter) getMetadataTextSize(ctx context.Context, userAccessToken, datasetID, edition, version string, metadata dataset.Metadata, dimensions dataset.Dimensions) (int, error) {
	var b bytes.Buffer
	collectionID := getCollectionIDFromContext(ctx)

	b.WriteString(metadata.ToString())
	b.WriteString("Dimensions:\n")
	for _, dimension := range dimensions.Items {
		options, err := f.DatasetClient.GetOptions(ctx, userAccessToken, "", collectionID, datasetID, edition, version, dimension.Name)
		if err != nil {
			return 0, err
		}

		b.WriteString(options.String())
	}
	return len(b.Bytes()), nil
}
