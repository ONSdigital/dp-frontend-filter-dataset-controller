package mapper

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	hierarchyClient "github.com/ONSdigital/dp-api-clients-go/v2/hierarchy"
	"github.com/ONSdigital/dp-api-clients-go/v2/search"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-cookies/cookies"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/dates"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/model"
	core "github.com/ONSdigital/dp-renderer/v2/model"
	"github.com/ONSdigital/log.go/v2/log"
)

const (
	age             = "age"
	geography       = "geography"
	latest          = "latest"
	list            = "list"
	single          = "single"
	sixteensVersion = "a18521a"
	strRange        = "range"
	strTime         = "time"
	strType         = "type"
)

var hierarchyBrowseLookup = map[string]string{
	"geography": "area",
}

var topLevelGeographies = map[string]bool{
	"K02000001": true,
	"K03000001": true,
	"K04000001": true,
}

var (
	cfg, _ = config.Get()
)

// CreateFilterOverview maps data items from API responses to form a filter overview
// front end page model
func CreateFilterOverview(req *http.Request, bp core.Page, dimensions []filter.ModelDimension, datasetDims dataset.VersionDimensionItems, fm filter.Model, dst dataset.DatasetDetails, filterID, datasetID, apiRouterVersion, lang, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) model.Overview {
	p := model.Overview{
		Page: bp,
	}
	p.BetaBannerEnabled = true
	p.FeatureFlags.SixteensVersion = sixteensVersion
	p.RemoveGalleryBackground = true

	mapCookiePreferences(req, &p.CookiesPreferencesSet, &p.CookiesPolicy)

	ctx := req.Context()
	log.Info(ctx, "mapping api response models into filter overview page model", log.Data{"filterID": filterID, "datasetID": datasetID})

	p.FilterID = filterID
	p.DatasetTitle = dst.Title
	p.Metadata.Title = "Filter Options"
	p.DatasetId = datasetID
	p.Language = lang
	p.URI = req.URL.Path
	p.ServiceMessage = serviceMessage
	p.EmergencyBanner = mapEmergencyBanner(emergencyBannerContent)
	p.FeatureFlags.FeedbackAPIURL = cfg.FeedbackAPIURL

	for i := range dimensions {
		var fod model.Dimension

		if dimensions[i].Name == strTime {
			for j := range datasetDims {
				if datasetDims[j].Name == dimensions[i].Name {
					fod.Filter = helpers.TitleCaseStr(datasetDims[j].Name)
					if datasetDims[j].Label != "" {
						fod.Filter = datasetDims[j].Label
					}
				}
			}

			times, err := dates.ConvertToReadable(dimensions[i].Values)
			if err != nil {
				log.Warn(ctx, "unable to convert dates to human readable values", log.FormatErrors([]error{err}))
				fod.AddedCategories = append(fod.AddedCategories, dimensions[i].Values...)
			}

			for _, time := range times {
				fod.AddedCategories = append(fod.AddedCategories, time.Format("January 2006"))
			}
		} else {
			fod.AddedCategories = append(fod.AddedCategories, dimensions[i].Values...)

			for j := range datasetDims {
				if datasetDims[j].Name == dimensions[i].Name {
					fod.Filter = helpers.TitleCaseStr(datasetDims[j].Name)
					if datasetDims[j].Label != "" {
						fod.Filter = datasetDims[j].Label
					}
				}
			}
		}

		fod.Link.URL = fmt.Sprintf("/filters/%s/dimensions/%s", filterID, dimensions[i].Name)

		if len(fod.AddedCategories) > 0 {
			fod.Link.Label = "Edit"
		} else {
			fod.Link.Label = "Add"
			fod.HasNoCategory = true
			p.Data.UnsetDimensions = append(p.Data.UnsetDimensions, fod.Filter)
		}

		p.Data.Dimensions = append(p.Data.Dimensions, fod)
	}

	p.Data.Cancel.URL = "/"
	p.Data.ClearAll.URL = fmt.Sprintf("/filters/%s/dimensions/clear-all", filterID)
	p.SearchDisabled = true

	versionURL, err := url.Parse(fm.Links.Version.HRef)
	if err != nil {
		log.Warn(ctx, "unable to parse version url", log.FormatErrors([]error{err}))
	}
	versionPath := strings.TrimPrefix(versionURL.Path, apiRouterVersion)

	p.IsInFilterBreadcrumb = true

	_, edition, _, err := helpers.ExtractDatasetInfoFromPath(ctx, versionPath)
	if err != nil {
		log.Warn(ctx, "unable to extract edition from url", log.FormatErrors([]error{err}))
	}

	p.Breadcrumb = append(
		p.Breadcrumb,
		core.TaxonomyNode{
			Title: dst.Title,
			URI:   fmt.Sprintf("/datasets/%s/editions", dst.ID),
		}, core.TaxonomyNode{
			Title: edition,
			URI:   versionPath,
		}, core.TaxonomyNode{
			Title: "Filter options",
		})

	return p
}

// CreateListSelectorPage maps items from API responses to form the model for a
// dimension list selector page
func CreateListSelectorPage(req *http.Request, bp core.Page, name string, selectedValues []filter.DimensionOption, allValues dataset.Options, fm filter.Model, dst dataset.DatasetDetails, dims dataset.VersionDimensions, datasetID, apiRouterVersion, lang, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) model.Selector {
	p := model.Selector{
		Page: bp,
	}
	p.BetaBannerEnabled = true
	p.FeatureFlags.SixteensVersion = sixteensVersion
	p.RemoveGalleryBackground = true

	mapCookiePreferences(req, &p.CookiesPreferencesSet, &p.CookiesPolicy)

	ctx := req.Context()
	log.Info(ctx, "mapping api response models to list selector page model", log.Data{"filterID": fm.FilterID, "datasetID": datasetID, "dimension": name})

	pageTitle := helpers.TitleCaseStr(name)

	for i := range dims.Items {
		if dims.Items[i].Name == name {
			p.Metadata.Description = dims.Items[i].Description
			if dims.Items[i].Label != "" {
				pageTitle = dims.Items[i].Label
			}
		}
	}

	p.SearchDisabled = true
	p.FilterID = fm.FilterID
	p.DatasetTitle = dst.Title
	p.Data.Title = pageTitle
	p.Metadata.Title = pageTitle
	p.DatasetId = datasetID
	p.Language = lang
	p.URI = req.URL.Path
	p.ServiceMessage = serviceMessage
	p.EmergencyBanner = mapEmergencyBanner(emergencyBannerContent)
	p.FeatureFlags.FeedbackAPIURL = cfg.FeedbackAPIURL

	versionURL, err := url.Parse(fm.Links.Version.HRef)
	if err != nil {
		log.Warn(ctx, "unable to parse version url", log.FormatErrors([]error{err}))
	}
	versionPath := strings.TrimPrefix(versionURL.Path, apiRouterVersion)

	p.IsInFilterBreadcrumb = true

	_, edition, _, err := helpers.ExtractDatasetInfoFromPath(ctx, versionPath)
	if err != nil {
		log.Warn(ctx, "unable to extract edition from url", log.FormatErrors([]error{err}))
	}

	p.Breadcrumb = append(
		p.Breadcrumb,
		core.TaxonomyNode{
			Title: dst.Title,
			URI:   fmt.Sprintf("/datasets/%s/editions", dst.ID),
		}, core.TaxonomyNode{
			Title: edition,
			URI:   versionPath,
		}, core.TaxonomyNode{
			Title: "Filter options",
			URI:   fmt.Sprintf("/filters/%s/dimensions", fm.FilterID),
		}, core.TaxonomyNode{
			Title: pageTitle,
		})

	p.Data.AddFromRange = model.Link{
		Label: fmt.Sprintf("add %s range", name),
		URL:   fmt.Sprintf("/filters/%s/dimensions/%s", fm.FilterID, name),
	}

	p.Data.SaveAndReturn = model.Link{
		URL: fmt.Sprintf("/filters/%s/dimensions", fm.FilterID),
	}
	p.Data.Cancel = model.Link{
		URL: fmt.Sprintf("/filters/%s/dimensions", fm.FilterID),
	}

	p.Data.AddAllInRange = model.Link{
		Label: fmt.Sprintf("All %ss", name),
	}

	p.Data.RangeData.URL = fmt.Sprintf("/filters/%s/dimensions/%s/list", fm.FilterID, name)

	p.Data.RemoveAll.URL = fmt.Sprintf("/filters/%s/dimensions/%s/remove-all", fm.FilterID, name)

	lookup := getIDNameLookup(allValues)

	selectedListValues := []string{}
	for _, opt := range selectedValues {
		selectedListValues = append(selectedListValues, lookup[opt.Option])
	}

	allListValues := []string{}
	valueIDmap := make(map[string]string)
	for i := range allValues.Items {
		allListValues = append(allListValues, allValues.Items[i].Label)
		valueIDmap[allValues.Items[i].Label] = allValues.Items[i].Option
	}

	for _, val := range allListValues {
		var isSelected bool
		for _, sval := range selectedListValues {
			if sval == val {
				isSelected = true
			}
		}
		p.Data.RangeData.Values = append(p.Data.RangeData.Values, model.Value{
			Label:      val,
			ID:         valueIDmap[val],
			IsSelected: isSelected,
		})
	}

	for _, val := range selectedListValues {
		p.Data.FiltersAdded = append(p.Data.FiltersAdded, model.Filter{
			RemoveURL: fmt.Sprintf("/filters/%s/dimensions/%s/remove/%s", fm.FilterID, name, valueIDmap[val]),
			Label:     val,
			ID:        valueIDmap[val],
		})
	}

	if len(allListValues) == len(selectedListValues) {
		p.Data.AddAllChecked = true
	}

	p.Data.FiltersAmount = len(selectedListValues)

	return p
}

// CreatePreviewPage maps data items from API responses to create a preview page
func CreatePreviewPage(req *http.Request, bp core.Page, dimensions []filter.ModelDimension, fm filter.Model, dst dataset.DatasetDetails, filterOutputID, datasetID, releaseDate, apiRouterVersion string, enableDatasetPreview bool, lang, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) model.Preview {
	p := model.Preview{
		Page: bp,
	}
	p.FeatureFlags.SixteensVersion = sixteensVersion
	p.Metadata.Title = "Preview and Download"
	p.BetaBannerEnabled = true
	p.EnableDatasetPreview = enableDatasetPreview
	p.Language = lang
	p.ServiceMessage = serviceMessage
	p.EmergencyBanner = mapEmergencyBanner(emergencyBannerContent)
	p.RemoveGalleryBackground = true
	p.FeatureFlags.FeedbackAPIURL = cfg.FeedbackAPIURL

	mapCookiePreferences(req, &p.CookiesPreferencesSet, &p.CookiesPolicy)

	ctx := req.Context()
	log.Info(ctx, "mapping api responses to preview page model", log.Data{"filterOutputID": filterOutputID, "datasetID": datasetID})

	p.SearchDisabled = false
	p.ReleaseDate = releaseDate
	p.Data.UnitOfMeasurement = dst.UnitOfMeasure
	p.URI = req.URL.Path

	versionURL, err := url.Parse(fm.Links.Version.HRef)
	if err != nil {
		log.Warn(ctx, "unable to parse version url", log.FormatErrors([]error{err}))
	}
	versionPath := strings.TrimPrefix(versionURL.Path, apiRouterVersion)

	p.Data.CurrentVersionURL = versionPath

	p.IsInFilterBreadcrumb = true

	_, edition, _, err := helpers.ExtractDatasetInfoFromPath(ctx, versionPath)
	if err != nil {
		log.Warn(ctx, "unable to extract edition from url", log.FormatErrors([]error{err}))
	}

	p.Breadcrumb = append(
		p.Breadcrumb,
		core.TaxonomyNode{
			Title: dst.Title,
			URI:   fmt.Sprintf("/datasets/%s/editions", dst.ID),
		}, core.TaxonomyNode{
			Title: edition,
			URI:   versionPath,
		}, core.TaxonomyNode{
			Title: "Filter options",
			URI:   fmt.Sprintf("/filters/%s/dimensions", fm.Links.FilterBlueprint.ID),
		}, core.TaxonomyNode{
			Title: "Preview",
		})

	p.Data.FilterID = fm.Links.FilterBlueprint.ID
	p.Data.FilterOutputID = filterOutputID

	p.DatasetTitle = dst.Title
	p.Data.DatasetID = datasetID
	p.DatasetId = datasetID
	_, editionFromPath, _, err := helpers.ExtractDatasetInfoFromPath(ctx, versionPath)
	if err != nil {
		log.Warn(ctx, "unable to extract edition from url", log.FormatErrors([]error{err}))
	}
	p.Data.Edition = editionFromPath

	for ext, d := range fm.Downloads {
		p.Data.Downloads = append(p.Data.Downloads, model.Download{
			Extension: ext,
			Size:      d.Size,
			URI:       d.URL,
			Skipped:   d.Skipped,
		})
	}

	for i := range dimensions {
		p.Data.Dimensions = append(p.Data.Dimensions, model.PreviewDimension{
			Name:   dimensions[i].Name,
			Values: dimensions[i].Values,
		})
	}
	if enableDatasetPreview && p.Data.Dimensions == nil {
		p.NoDimensionData = true
	}

	return p
}

func getNameIDLookup(vals dataset.Options) map[string]string {
	lookup := make(map[string]string)
	for i := range vals.Items {
		lookup[vals.Items[i].Label] = vals.Items[i].Option
	}
	return lookup
}

func getIDNameLookup(vals dataset.Options) map[string]string {
	lookup := make(map[string]string)
	for i := range vals.Items {
		lookup[vals.Items[i].Option] = vals.Items[i].Label
	}
	return lookup
}

// CreateAgePage creates an age selector page based on api responses
// TODO: refactor to reduce complexity
//
//nolint:gocyclo // cyclomatic complexity 27
func CreateAgePage(req *http.Request, bp core.Page, f filter.Model, d dataset.DatasetDetails, allVals dataset.Options, selVals filter.DimensionOptions, dims dataset.VersionDimensions, datasetID, apiRouterVersion, lang, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) (model.Age, error) {
	p := model.Age{
		Page: bp,
	}
	if req == nil {
		return p, errors.New("invalid request provided to CreateAgePage")
	}
	p.BetaBannerEnabled = true
	p.FeatureFlags.SixteensVersion = sixteensVersion
	p.RemoveGalleryBackground = true

	ctx := req.Context()

	mapCookiePreferences(req, &p.CookiesPreferencesSet, &p.CookiesPolicy)

	log.Info(ctx, "mapping api responses to age page model", log.Data{"filterID": f.FilterID, "datasetID": datasetID})

	for i := range dims.Items {
		if dims.Items[i].Name == age {
			p.Metadata.Description = dims.Items[i].Description
		}
	}

	p.FilterID = f.FilterID
	p.SearchDisabled = true
	p.DatasetId = datasetID
	p.Language = lang
	p.URI = req.URL.Path
	p.ServiceMessage = serviceMessage
	p.EmergencyBanner = mapEmergencyBanner(emergencyBannerContent)
	p.FeatureFlags.FeedbackAPIURL = cfg.FeedbackAPIURL

	versionURL, err := url.Parse(f.Links.Version.HRef)
	if err != nil {
		return model.Age{}, err
	}
	versionPath := strings.TrimPrefix(versionURL.Path, apiRouterVersion)

	p.IsInFilterBreadcrumb = true

	_, edition, _, err := helpers.ExtractDatasetInfoFromPath(ctx, versionPath)
	if err != nil {
		log.Warn(ctx, "unable to extract edition from url", log.FormatErrors([]error{err}))
	}

	p.Breadcrumb = append(
		p.Breadcrumb,
		core.TaxonomyNode{
			Title: d.Title,
			URI:   fmt.Sprintf("/datasets/%s/editions", d.ID),
		}, core.TaxonomyNode{
			Title: edition,
			URI:   versionPath,
		}, core.TaxonomyNode{
			Title: "Filter options",
			URI:   fmt.Sprintf("/filters/%s/dimensions", f.FilterID),
		}, core.TaxonomyNode{
			Title: "Age",
		})

	p.Metadata.Title = "Age"
	p.DatasetTitle = d.Title

	p.Data.FormAction.URL = fmt.Sprintf("/filters/%s/dimensions/age/update", f.FilterID)

	// get mapping of labels (keys) to options (values) and initialise aux vars
	labelIDs := getNameIDLookup(allVals)
	youngest := math.MaxInt32
	oldest := math.MinInt32

	// iterate all values, and add them to the Page in the same order,
	// setting the 'isSelected' for each one of them (according to selVals)
	// and setting oldest and youngest values
	for i := range allVals.Items {
		// if the age Label contains '+', we assume that it is the oldest age value
		if strings.Contains(allVals.Items[i].Label, "+") {
			p.Data.Oldest = allVals.Items[i].Label
		} else {
			// get the Int values, if there is an error, we assume that 'allOptions' was selected
			ageInt, err := strconv.Atoi(allVals.Items[i].Label)
			if err != nil {
				p.Data.HasAllAges = true
				p.Data.AllAgesOption = allVals.Items[i].Option
				continue
			}
			// refresh youngest and oldest values if needed
			if ageInt < youngest {
				youngest = ageInt
			}
			if ageInt > oldest {
				oldest = ageInt
			}
		}

		// find if the option is selected
		var isSelected bool
		for _, selVal := range selVals.Items {
			if selVal.Option == labelIDs[allVals.Items[i].Label] {
				isSelected = true
			}
		}

		// append the age value to the page
		p.Data.Ages = append(p.Data.Ages, model.AgeValue{
			Option:     labelIDs[allVals.Items[i].Label],
			Label:      allVals.Items[i].Label,
			IsSelected: isSelected,
		})
	}

	if p.Data.Youngest == "" && youngest < math.MaxInt32 {
		p.Data.Youngest = strconv.Itoa(youngest)
	}
	if p.Data.Oldest == "" && oldest > math.MinInt32 {
		p.Data.Oldest = strconv.Itoa(oldest)
	}

	p.Data.CheckedRadio = strRange

	for i, val := range p.Data.Ages {
		if val.IsSelected {
			for j := i; j < len(p.Data.Ages); j++ {
				if p.Data.Ages[j].IsSelected {
					continue
				}
				for k := j; k < len(p.Data.Ages); k++ {
					if p.Data.Ages[k].IsSelected {
						p.Data.CheckedRadio = list
						break
					}
				}
			}
		}
	}

	if p.Data.CheckedRadio == strRange {
		for _, val := range p.Data.Ages {
			if val.IsSelected {
				if p.Data.FirstSelected == "" {
					p.Data.FirstSelected = val.Label
				}
				p.Data.LastSelected = val.Label
			}
		}
	}

	return p, nil
}

// CreateTimePage will create a time selector page based on api response models
// TODO: refactor to reduce complexity
//
//nolint:gocyclo // cyclomatic complexity 36
func CreateTimePage(req *http.Request, bp core.Page, f filter.Model, d dataset.DatasetDetails, allVals dataset.Options, selVals []filter.DimensionOption, dims dataset.VersionDimensions, datasetID, apiRouterVersion, lang, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) (model.Time, error) {
	p := model.Time{
		Page: bp,
	}
	p.BetaBannerEnabled = true
	p.FeatureFlags.SixteensVersion = sixteensVersion
	p.RemoveGalleryBackground = true

	mapCookiePreferences(req, &p.CookiesPreferencesSet, &p.CookiesPolicy)

	ctx := req.Context()

	if len(allVals.Items) == 0 {
		return p, nil
	}

	if _, err := time.Parse("Jan-06", allVals.Items[0].Option); err == nil {
		p.Data.Type = "month"
	}

	p.DatasetTitle = d.Title
	p.FilterID = f.FilterID
	p.SearchDisabled = true
	p.DatasetId = datasetID
	p.Language = lang
	p.URI = req.URL.Path
	p.ServiceMessage = serviceMessage
	p.EmergencyBanner = mapEmergencyBanner(emergencyBannerContent)
	p.FeatureFlags.FeedbackAPIURL = cfg.FeedbackAPIURL

	for i := range dims.Items {
		if dims.Items[i].Name == strTime {
			p.Metadata.Description = dims.Items[i].Description
		}
	}

	versionURL, err := url.Parse(f.Links.Version.HRef)
	if err != nil {
		return p, err
	}
	versionPath := strings.TrimPrefix(versionURL.Path, apiRouterVersion)

	p.IsInFilterBreadcrumb = true

	_, edition, _, err := helpers.ExtractDatasetInfoFromPath(ctx, versionPath)
	if err != nil {
		log.Warn(ctx, "unable to extract edition from url", log.FormatErrors([]error{err}))
	}

	p.Breadcrumb = append(p.Breadcrumb, core.TaxonomyNode{
		Title: d.Title,
		URI:   fmt.Sprintf("/datasets/%s/editions", d.ID),
	}, core.TaxonomyNode{
		Title: edition,
		URI:   versionPath,
	}, core.TaxonomyNode{
		Title: "Filter options",
		URI:   fmt.Sprintf("/filters/%s/dimensions", f.FilterID),
	}, core.TaxonomyNode{
		Title: "Time",
	})

	p.Metadata.Title = "Time"

	lookup := getNameIDLookup(allVals)

	allTimes := []string{}
	for i := range allVals.Items {
		allTimes = append(allTimes, allVals.Items[i].Label)
	}

	times, err := dates.ConvertToReadable(allTimes)
	if err != nil {
		return p, err
	}

	// sort just to find first and latest, but not to be used as the order in the UI
	sortedTimes := make([]time.Time, len(times))
	copy(sortedTimes, times)
	dates.Sort(sortedTimes)

	p.Data.FirstTime = model.TimeValue{
		Option: lookup[sortedTimes[0].Format("Jan-06")],
		Month:  sortedTimes[0].Month().String(),
		Year:   fmt.Sprintf("%d", sortedTimes[0].Year()),
	}

	p.Data.LatestTime = model.TimeValue{
		Option: lookup[sortedTimes[len(sortedTimes)-1].Format("Jan-06")],
		Month:  sortedTimes[len(sortedTimes)-1].Month().String(),
		Year:   fmt.Sprintf("%d", sortedTimes[len(sortedTimes)-1].Year()),
	}

	firstYear := sortedTimes[0].Year()
	lastYear := sortedTimes[len(sortedTimes)-1].Year()
	diffYears := lastYear - firstYear

	p.Data.Years = append(p.Data.Years, "Select")
	for i := 0; i < diffYears+1; i++ {
		p.Data.Years = append(p.Data.Years, fmt.Sprintf("%d", firstYear+i))
	}

	p.Data.Months = append(p.Data.Months, "Select")
	for i := 0; i < 12; i++ {
		p.Data.Months = append(p.Data.Months, time.Month(i+1).String())
	}

	latestSelected := false
	for _, val := range times {
		var isSelected bool
		for _, selVal := range selVals {
			if val.Format("Jan-06") == selVal.Option {
				isSelected = true
				if val == sortedTimes[len(sortedTimes)-1] {
					latestSelected = true
				}
			}
		}

		p.Data.Values = append(p.Data.Values, model.TimeValue{
			Option:     lookup[val.Format("Jan-06")],
			Month:      val.Month().String(),
			Year:       fmt.Sprintf("%d", val.Year()),
			IsSelected: isSelected,
		})
	}

	p.Data.FormAction = model.Link{
		URL: fmt.Sprintf("/filters/%s/dimensions/time/update", f.FilterID),
	}

	if len(selVals) == 1 && latestSelected {
		p.Data.CheckedRadio = latest
	} else if len(selVals) == 1 {
		p.Data.CheckedRadio = single
		date, err := time.Parse("Jan-06", selVals[0].Option)
		if err != nil {
			log.Warn(ctx, "unable to parse date", log.FormatErrors([]error{err}))
		}
		p.Data.SelectedStartMonth = date.Month().String()
		p.Data.SelectedStartYear = fmt.Sprintf("%d", date.Year())
	} else if len(selVals) == 0 {
		p.Data.CheckedRadio = ""
	} else if len(selVals) == len(allVals.Items) {
		p.Data.CheckedRadio = list
	} else {
		if isTimeRange(sortedTimes, selVals) {
			p.Data.CheckedRadio = strRange
		} else {
			p.Data.CheckedRadio = list
		}
	}

	if p.Data.CheckedRadio == strRange {
		var selOptions []string
		for _, val := range selVals {
			selOptions = append(selOptions, val.Option)
		}

		selDates, err := dates.ConvertToReadable(selOptions)
		if err != nil {
			log.Warn(ctx, "unable to convert dates to human readable values", log.FormatErrors([]error{err}))
		}

		selDates = dates.Sort(selDates)

		p.Data.SelectedStartMonth = selDates[0].Month().String()
		p.Data.SelectedStartYear = fmt.Sprintf("%d", selDates[0].Year())
		p.Data.SelectedEndMonth = selDates[len(selDates)-1].Month().String()
		p.Data.SelectedEndYear = fmt.Sprintf("%d", selDates[len(selDates)-1].Year())
	}
	var minYear, maxYear string
	var selectedMonths []string
	for _, selVal := range selVals {
		month, err := time.Parse("Jan-06", selVal.Option)
		if err != nil {
			log.Error(ctx, "unable to convert date to month value", err)
			continue
		}
		monthStr := month.Format("January")
		_, found := helpers.StringInSlice(monthStr, selectedMonths)
		if !found {
			selectedMonths = append(selectedMonths, monthStr)
		}
		yearStr := month.Format("2006")
		if minYear == "" {
			minYear = yearStr
		}
		if maxYear == "" {
			maxYear = yearStr
		}
		yearInt, err := strconv.Atoi(yearStr)
		if err != nil {
			log.Error(ctx, "unable to convert year string to int for comparison", err)
			continue
		}
		maxYearInt, err := strconv.Atoi(maxYear)
		if err != nil {
			log.Error(ctx, "unable to convert max year string to int for comparison", err)
			continue
		}
		minYearInt, err := strconv.Atoi(minYear)
		if err != nil {
			log.Error(ctx, "unable to convert min year string to int for comparison", err)
			continue
		}
		if yearInt > maxYearInt {
			maxYear = yearStr
		} else if yearInt < minYearInt {
			minYear = yearStr
		}
	}
	var listOfAllMonths []model.Month
	numberOfMonthsInAYear := 12
	for i := 0; i < numberOfMonthsInAYear; i++ {
		monthName := time.Month(i + 1).String()
		_, isSelected := helpers.StringInSlice(monthName, selectedMonths)
		singleMonth := model.Month{
			Name:       monthName,
			IsSelected: isSelected,
		}
		listOfAllMonths = append(listOfAllMonths, singleMonth)
	}
	GroupedSelection := model.GroupedSelection{
		Months:    listOfAllMonths,
		YearStart: minYear,
		YearEnd:   maxYear,
	}
	p.Data.GroupedSelection = GroupedSelection

	return p, nil
}

// isTimeRange determines if the selected values define a single continuous range of sorted items
// - sortedTimes is a list of times in the order against which we want to determine the range selection
// - selVals is a list of selected Options, with the Option value having "Jan-06" format
func isTimeRange(sortedTimes []time.Time, selVals []filter.DimensionOption) bool {
	// a range has to have at least two items
	if len(selVals) < 2 {
		return false
	}

	// state variables to determine that a single range is found
	inRange := false
	fullRangeFound := false

	// iterate sortedTimes, we assume that the times are already sorted in the required order to determine the range
	for _, val := range sortedTimes {
		// state variable to determine if val is selected
		isSelected := false

		valueToFind := val.Format("Jan-06")

		// determine if the time value is selected
		for _, selVal := range selVals {
			// if this condition is satisfied, the value is selected
			if valueToFind == selVal.Option {
				isSelected = true

				// if there was already a complete range, this selected item would start a new discontinuous range.
				if fullRangeFound {
					return false
				}

				// we are in either a new range or continuing an existing range
				inRange = true
				continue
			}
		}

		// value is not selected. If we were in a range, this is now complete
		if !isSelected {
			if inRange {
				fullRangeFound = true
			}
			inRange = false
		}
	}

	// we reached the end of the loop without fining a discontinuity.
	return fullRangeFound
}

// CreateHierarchySearchPage forms a search page based on various api response models
func CreateHierarchySearchPage(req *http.Request, bp core.Page, items []search.Item, dst dataset.DatasetDetails, f filter.Model, selectedValueLabels map[string]string, dims []dataset.VersionDimension, name, curPath, datasetID, referrer, query, apiRouterVersion, lang, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) model.Hierarchy {
	p := model.Hierarchy{
		Page: bp,
	}
	p.BetaBannerEnabled = true
	p.FeatureFlags.SixteensVersion = sixteensVersion
	p.RemoveGalleryBackground = true

	mapCookiePreferences(req, &p.CookiesPreferencesSet, &p.CookiesPolicy)

	ctx := req.Context()
	log.Info(ctx, "mapping api response models to hierarchy search page", log.Data{"filterID": f.FilterID, "datasetID": datasetID, "name": name})

	pageTitle := helpers.TitleCaseStr(name)
	for i := range dims {
		if dims[i].Name == name && dims[i].Label != "" {
			pageTitle = dims[i].Label
		}
	}
	p.DatasetTitle = dst.Title
	p.Data.DimensionName = pageTitle
	p.DatasetId = datasetID
	p.Data.IsSearchResults = true
	p.Data.Query = query
	p.Language = lang
	p.URI = fmt.Sprintf("%s?q=%s", req.URL.Path, url.QueryEscape(req.URL.Query().Get("q")))
	p.ServiceMessage = serviceMessage
	p.EmergencyBanner = mapEmergencyBanner(emergencyBannerContent)
	p.FeatureFlags.FeedbackAPIURL = cfg.FeedbackAPIURL

	title := pageTitle

	p.IsInFilterBreadcrumb = true
	var ok bool
	if p.Type, ok = hierarchyBrowseLookup[name]; !ok {
		p.Type = strType
	}

	p.SearchDisabled = true

	p.Data.SearchURL = fmt.Sprintf("/filters/%s/dimensions/%s/search", f.FilterID, name)

	versionURL, err := url.Parse(f.Links.Version.HRef)
	if err != nil {
		log.Warn(ctx, "unable to parse version url", log.FormatErrors([]error{err}))
	}
	versionPath := strings.TrimPrefix(versionURL.Path, apiRouterVersion)

	p.Data.LandingPageURL = versionPath + "#id-dimensions"
	p.Breadcrumb = append(p.Breadcrumb, core.TaxonomyNode{
		Title: dst.Title,
		URI:   versionPath,
	}, core.TaxonomyNode{
		Title: "Filter options",
		URI:   fmt.Sprintf("/filters/%s/dimensions", f.FilterID),
	}, core.TaxonomyNode{
		Title: title,
		URI:   fmt.Sprintf("/filters/%s/dimensions/%s", f.FilterID, name),
	}, core.TaxonomyNode{
		Title: "Search results",
	})

	p.FilterID = f.FilterID
	p.Data.Title = title
	p.Metadata.Title = title

	p.Data.GoBack.URL = referrer

	p.Data.AddAllFilters.URL = curPath + "/add-all"
	p.Data.RemoveAll.URL = curPath + "/remove-all"

	for option, label := range selectedValueLabels {
		p.Data.FiltersAdded = append(p.Data.FiltersAdded, model.Filter{
			Label:     label,
			RemoveURL: fmt.Sprintf("%s/remove/%s", curPath, option),
			ID:        option,
		})
	}

	if len(items) == 0 {
		p.Data.IsSearchError = true
	} else {
		for _, item := range items {
			_, selected := selectedValueLabels[item.Code]
			p.Data.FilterList = append(p.Data.FilterList, model.List{
				Label:    item.Label,
				ID:       item.Code,
				SubNum:   strconv.Itoa(item.NumberOfChildren),
				SubURL:   fmt.Sprintf("redirect:/filters/%s/dimensions/%s/%s", f.FilterID, name, item.Code),
				Selected: selected,
				HasData:  item.HasData,
			})
		}
	}

	p.Data.SaveAndReturn.URL = fmt.Sprintf("/filters/%s/dimensions/%s/search/update", f.FilterID, name)
	p.Data.Cancel.URL = fmt.Sprintf("/filters/%s/dimensions", f.FilterID)

	return p
}

// CreateHierarchyPage maps data items from API responses to form a hierarchy page
// TODO: refactor to reduce complexity
//
//nolint:gocyclo // cyclomatic complexity 26
func CreateHierarchyPage(req *http.Request, bp core.Page, h hierarchyClient.Model, dst dataset.DatasetDetails, f filter.Model, selectedValueLabels map[string]string, dims dataset.VersionDimensions, name, curPath, datasetID, apiRouterVersion, lang, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) model.Hierarchy {
	p := model.Hierarchy{
		Page: bp,
	}
	p.BetaBannerEnabled = true
	p.FeatureFlags.SixteensVersion = sixteensVersion
	p.Language = lang
	p.RemoveGalleryBackground = true

	mapCookiePreferences(req, &p.CookiesPreferencesSet, &p.CookiesPolicy)

	ctx := req.Context()
	log.Info(ctx, "mapping api response models to hierarchy page", log.Data{"filterID": f.FilterID, "datasetID": datasetID, "label": h.Label})

	pageTitle := helpers.TitleCaseStr(name)
	for i := range dims.Items {
		if dims.Items[i].Name == name {
			p.Metadata.Description = dims.Items[i].Description
			if dims.Items[i].Label != "" {
				pageTitle = dims.Items[i].Label
			}
		}
	}

	p.DatasetTitle = dst.Title
	p.Data.DimensionName = pageTitle
	p.DatasetId = datasetID
	p.URI = req.URL.Path
	p.ServiceMessage = serviceMessage
	p.EmergencyBanner = mapEmergencyBanner(emergencyBannerContent)
	p.FeatureFlags.FeedbackAPIURL = cfg.FeedbackAPIURL

	var title string
	if len(h.Breadcrumbs) == 0 {
		title = pageTitle
	} else {
		title = h.Label
	}

	var ok bool
	if p.Type, ok = hierarchyBrowseLookup[name]; !ok {
		p.Type = "type"
	}

	p.SearchDisabled = true

	p.Data.SearchURL = fmt.Sprintf("/filters/%s/dimensions/%s/search", f.FilterID, name)

	versionURL, err := url.Parse(f.Links.Version.HRef)
	if err != nil {
		log.Warn(ctx, "unable to parse version url", log.FormatErrors([]error{err}))
	}
	versionPath := strings.TrimPrefix(versionURL.Path, apiRouterVersion)

	p.IsInFilterBreadcrumb = true

	_, edition, _, err := helpers.ExtractDatasetInfoFromPath(ctx, versionPath)
	if err != nil {
		log.Warn(ctx, "unable to extract edition from url", log.FormatErrors([]error{err}))
	}

	p.Breadcrumb = append(
		p.Breadcrumb,
		core.TaxonomyNode{
			Title: dst.Title,
			URI:   fmt.Sprintf("/datasets/%s/editions", dst.ID),
		}, core.TaxonomyNode{
			Title: edition,
			URI:   versionPath,
		}, core.TaxonomyNode{
			Title: "Filter options",
			URI:   fmt.Sprintf("/filters/%s/dimensions", f.FilterID),
		})

	if len(h.Breadcrumbs) > 0 {
		if name == geography {
			p.Breadcrumb = append(p.Breadcrumb, core.TaxonomyNode{
				Title: "Geographic Areas",
				URI:   fmt.Sprintf("/filters/%s/dimensions/%s", f.FilterID, geography),
			})

			if !topLevelGeographies[h.Links.Code.ID] {
				for i := len(h.Breadcrumbs) - 1; i >= 0; i-- {
					breadcrumb := h.Breadcrumbs[i]

					if !topLevelGeographies[breadcrumb.Links.Code.ID] {
						var uri string
						if breadcrumb.Links.Code.ID != "" {
							uri = fmt.Sprintf("/filters/%s/dimensions/%s/%s", f.FilterID, name, breadcrumb.Links.Code.ID)
						} else {
							uri = fmt.Sprintf("/filters/%s/dimensions/%s", f.FilterID, name)
						}

						p.Breadcrumb = append(p.Breadcrumb, core.TaxonomyNode{
							Title: breadcrumb.Label,
							URI:   uri,
						})
					}
				}
			}
		} else {
			for i := len(h.Breadcrumbs) - 1; i >= 0; i-- {
				breadcrumb := h.Breadcrumbs[i]

				var uri string
				if breadcrumb.Links.Code.ID != "" {
					uri = fmt.Sprintf("/filters/%s/dimensions/%s/%s", f.FilterID, name, breadcrumb.Links.Code.ID)
				} else {
					uri = fmt.Sprintf("/filters/%s/dimensions/%s", f.FilterID, name)
				}

				p.Breadcrumb = append(p.Breadcrumb, core.TaxonomyNode{
					Title: breadcrumb.Label,
					URI:   uri,
				})
			}
		}
	}

	p.Breadcrumb = append(p.Breadcrumb, core.TaxonomyNode{
		Title: title,
	})

	p.FilterID = f.FilterID
	p.Data.Title = title
	p.Metadata.Title = fmt.Sprintf("Filter Options - %s", title)

	if len(h.Breadcrumbs) > 0 {
		if len(h.Breadcrumbs) == 1 || topLevelGeographies[h.Breadcrumbs[0].Links.Code.ID] && name == geography {
			p.Data.Parent = pageTitle
			p.Data.GoBack = model.Link{
				URL: fmt.Sprintf("/filters/%s/dimensions/%s", f.FilterID, name),
			}
		} else {
			p.Data.Parent = h.Breadcrumbs[0].Label
			p.Data.GoBack = model.Link{
				URL: fmt.Sprintf("/filters/%s/dimensions/%s/%s", f.FilterID, name, h.Breadcrumbs[0].Links.Code.ID),
			}
		}
	}

	p.Data.AddAllFilters.Amount = strconv.Itoa(len(h.Children))
	p.Data.AddAllFilters.URL = curPath + "/add-all"
	for _, child := range h.Children {
		if child.HasData {
			p.Data.HasData = true
			break
		}
	}
	p.Data.RemoveAll.URL = curPath + "/remove-all"

	for option, label := range selectedValueLabels {
		p.Data.FiltersAdded = append(p.Data.FiltersAdded, model.Filter{
			Label:     label,
			RemoveURL: fmt.Sprintf("%s/remove/%s", curPath, option),
			ID:        option,
		})
	}

	if h.HasData && len(h.Breadcrumbs) == 0 {
		_, selected := selectedValueLabels[h.Links.Code.ID]
		p.Data.FilterList = append(p.Data.FilterList, model.List{
			Label:    h.Label,
			ID:       h.Links.Code.ID,
			SubNum:   "0",
			SubURL:   "",
			Selected: selected,
			HasData:  true,
		})
	}

	for _, child := range h.Children {
		_, selected := selectedValueLabels[child.Links.Code.ID]
		p.Data.FilterList = append(p.Data.FilterList, model.List{
			Label:    child.Label,
			ID:       child.Links.Code.ID,
			SubNum:   strconv.Itoa(child.NumberofChildren),
			SubURL:   fmt.Sprintf("redirect:/filters/%s/dimensions/%s/%s", f.FilterID, name, child.Links.Code.ID),
			Selected: selected,
			HasData:  child.HasData,
		})
	}

	p.Data.SaveAndReturn.URL = curPath + "/update"
	p.Data.Cancel.URL = fmt.Sprintf("/filters/%s/dimensions", f.FilterID)

	return p
}

// mapCookiePreferences reads cookie policy and preferences cookies and then maps the values to the page model
func mapCookiePreferences(req *http.Request, preferencesIsSet *bool, policy *core.CookiesPolicy) {
	preferencesCookie := cookies.GetONSCookiePreferences(req)
	*preferencesIsSet = preferencesCookie.IsPreferenceSet
	*policy = core.CookiesPolicy{
		Communications: preferencesCookie.Policy.Campaigns,
		Essential:      preferencesCookie.Policy.Essential,
		Settings:       preferencesCookie.Policy.Settings,
		Usage:          preferencesCookie.Policy.Usage,
	}
}

func mapEmergencyBanner(bannerData zebedee.EmergencyBanner) core.EmergencyBanner {
	var mappedEmergencyBanner core.EmergencyBanner
	emptyBannerObj := zebedee.EmergencyBanner{}
	if bannerData != emptyBannerObj {
		mappedEmergencyBanner.Title = bannerData.Title
		mappedEmergencyBanner.Type = strings.Replace(bannerData.Type, "_", "-", -1)
		mappedEmergencyBanner.Description = bannerData.Description
		mappedEmergencyBanner.URI = bannerData.URI
		mappedEmergencyBanner.LinkText = bannerData.LinkText
	}
	return mappedEmergencyBanner
}
