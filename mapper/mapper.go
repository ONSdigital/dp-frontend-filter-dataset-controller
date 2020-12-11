package mapper

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/filter"
	hierarchyClient "github.com/ONSdigital/dp-api-clients-go/hierarchy"
	"github.com/ONSdigital/dp-api-clients-go/search"
	"github.com/ONSdigital/dp-cookies/cookies"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/dates"
	fdHelpers "github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"

	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/age"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/filterOverview"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/hierarchy"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/listSelector"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/previewPage"
	timeModel "github.com/ONSdigital/dp-frontend-models/model/dataset-filter/time"
	"github.com/ONSdigital/log.go/log"
)

var hierarchyBrowseLookup = map[string]string{
	"geography": "area",
}

var topLevelGeographies = map[string]bool{
	"K02000001": true,
	"K03000001": true,
	"K04000001": true,
}

// CreateFilterOverview maps data items from API responses to form a filter overview
// front end page model
func CreateFilterOverview(req *http.Request, dimensions []filter.ModelDimension, datasetDims dataset.VersionDimensionItems, filter filter.Model, dst dataset.DatasetDetails, filterID, datasetID, releaseDate, apiRouterVersion, lang string) filterOverview.Page {
	var p filterOverview.Page
	p.BetaBannerEnabled = true

	mapCookiePreferences(req, &p.CookiesPreferencesSet, &p.CookiesPolicy)

	ctx := req.Context()
	log.Event(ctx, "mapping api response models into filter overview page model", log.INFO, log.Data{"filterID": filterID, "datasetID": datasetID})

	p.FilterID = filterID
	p.DatasetTitle = dst.Title
	p.Metadata.Title = "Filter Options"
	p.DatasetId = datasetID
	p.Language = lang

	disableButton := true

	for _, d := range dimensions {
		var fod filterOverview.Dimension

		if len(d.Values) > 0 {
			disableButton = false
		}

		if d.Name == "time" {
			for _, dim := range datasetDims {
				if dim.Name == d.Name {
					fod.Filter = strings.Title(dim.Name)
					if len(dim.Label) > 0 {
						fod.Filter = dim.Label
					}
				}
			}
			times, err := dates.ConvertToReadable(d.Values)
			if err != nil {
				log.Event(ctx, "unable to convert dates to human readable values", log.WARN, log.Error(err))
				for _, ac := range d.Values {
					fod.AddedCategories = append(fod.AddedCategories, ac)
				}
			}

			times = dates.Sort(times)
			for _, time := range times {
				fod.AddedCategories = append(fod.AddedCategories, time.Format("January 2006"))
			}
		} else {
			for _, ac := range d.Values {
				fod.AddedCategories = append(fod.AddedCategories, ac)
			}

			for _, dim := range datasetDims {
				if dim.Name == d.Name {
					fod.Filter = strings.Title(dim.Name)
					if len(dim.Label) > 0 {
						fod.Filter = dim.Label
					}
				}
			}

			if d.Name == "age" {
				var ages []int
				for _, a := range fod.AddedCategories {
					age, err := strconv.Atoi(a)
					if err != nil {
						continue
					}

					ages = append(ages, age)
				}

				sort.Ints(ages)
				for i, age := range ages {
					fod.AddedCategories[i] = strconv.Itoa(age)
				}
			}
		}

		fod.Link.URL = fmt.Sprintf("/filters/%s/dimensions/%s", filterID, d.Name)

		if len(fod.AddedCategories) > 0 {
			fod.Link.Label = "Edit"
		} else {
			fod.Link.Label = "Add"
		}

		p.Data.Dimensions = append(p.Data.Dimensions, fod)
	}

	if p.Data.PreviewAndDownloadDisabled = disableButton; !p.Data.PreviewAndDownloadDisabled {
		p.Data.PreviewAndDownload.URL = fmt.Sprintf("/filters/%s", filterID)
	}

	p.Data.Cancel.URL = "/"
	p.Data.ClearAll.URL = fmt.Sprintf("/filters/%s/dimensions/clear-all", filterID)
	p.SearchDisabled = true

	versionURL, err := url.Parse(filter.Links.Version.HRef)
	if err != nil {
		log.Event(ctx, "unable to parse version url", log.WARN, log.Error(err))
	}
	versionPath := strings.TrimPrefix(versionURL.Path, apiRouterVersion)

	p.IsInFilterBreadcrumb = true

	_, edition, _, _ := helpers.ExtractDatasetInfoFromPath(versionPath)

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dst.Title,
		URI:   fmt.Sprintf("/datasets/%s/editions", dst.ID),
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: edition,
		URI:   versionPath,
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter options",
	})

	return p
}

// CreateListSelectorPage maps items from API responses to form the model for a
// dimension list selector page
func CreateListSelectorPage(req *http.Request, name string, selectedValues []filter.DimensionOption, allValues dataset.Options, filter filter.Model, dst dataset.DatasetDetails, dims dataset.VersionDimensions, datasetID, releaseDate, apiRouterVersion, lang string) listSelector.Page {
	var p listSelector.Page
	p.BetaBannerEnabled = true

	mapCookiePreferences(req, &p.CookiesPreferencesSet, &p.CookiesPolicy)

	ctx := req.Context()
	log.Event(ctx, "mapping api response models to list selector page model", log.INFO, log.Data{"filterID": filter.FilterID, "datasetID": datasetID, "dimension": name})

	pageTitle := strings.Title(name)

	for _, dim := range dims.Items {
		if dim.Name == name {
			p.Metadata.Description = dim.Description
			if len(dim.Label) > 0 {
				pageTitle = dim.Label
			}
		}
	}

	p.SearchDisabled = true
	p.FilterID = filter.FilterID
	p.DatasetTitle = dst.Title
	p.Data.Title = pageTitle
	p.Metadata.Title = pageTitle
	p.DatasetId = datasetID
	p.Language = lang

	versionURL, err := url.Parse(filter.Links.Version.HRef)
	if err != nil {
		log.Event(ctx, "unable to parse version url", log.WARN, log.Error(err))
	}
	versionPath := strings.TrimPrefix(versionURL.Path, apiRouterVersion)

	p.IsInFilterBreadcrumb = true

	_, edition, _, _ := helpers.ExtractDatasetInfoFromPath(versionPath)

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dst.Title,
		URI:   fmt.Sprintf("/datasets/%s/editions", dst.ID),
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: edition,
		URI:   versionPath,
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter options",
		URI:   fmt.Sprintf("/filters/%s/dimensions", filter.FilterID),
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: pageTitle,
	})

	p.Data.AddFromRange = listSelector.Link{
		Label: fmt.Sprintf("add %s range", name),
		URL:   fmt.Sprintf("/filters/%s/dimensions/%s", filter.FilterID, name),
	}

	p.Data.SaveAndReturn = listSelector.Link{
		URL: fmt.Sprintf("/filters/%s/dimensions", filter.FilterID),
	}
	p.Data.Cancel = listSelector.Link{
		URL: fmt.Sprintf("/filters/%s/dimensions", filter.FilterID),
	}

	p.Data.AddAllInRange = listSelector.Link{
		Label: fmt.Sprintf("All %ss", name),
	}

	p.Data.RangeData.URL = fmt.Sprintf("/filters/%s/dimensions/%s/list", filter.FilterID, name)

	p.Data.RemoveAll.URL = fmt.Sprintf("/filters/%s/dimensions/%s/remove-all", filter.FilterID, name)

	lookup := getIDNameLookup(allValues)

	var selectedListValues, selectedListIDs []string
	for _, opt := range selectedValues {
		selectedListValues = append(selectedListValues, lookup[opt.Option])
		selectedListIDs = append(selectedListIDs, opt.Option)
	}

	var allListValues []string
	valueIDmap := make(map[string]string)
	for _, val := range allValues.Items {
		allListValues = append(allListValues, val.Label)
		valueIDmap[val.Label] = val.Option
	}

	if name == "time" || name == "age" {
		isValid := true
		var intVals []int
		for _, val := range allListValues {
			intVal, err := strconv.Atoi(val)
			if err != nil {
				isValid = false
				break
			}
			intVals = append(intVals, intVal)
		}

		if isValid {
			sort.Ints(intVals)
			if name == "time" {
				intVals = reverseInts(intVals)
			}
			for i, val := range intVals {
				allListValues[i] = strconv.Itoa(val)
			}
		}
	} else {
		sort.Strings(allListValues)
	}

	for _, val := range allListValues {
		var isSelected bool
		for _, sval := range selectedListValues {
			if sval == val {
				isSelected = true
			}
		}
		p.Data.RangeData.Values = append(p.Data.RangeData.Values, listSelector.Value{
			Label:      val,
			ID:         valueIDmap[val],
			IsSelected: isSelected,
		})
	}

	for _, val := range selectedListValues {
		p.Data.FiltersAdded = append(p.Data.FiltersAdded, listSelector.Filter{
			RemoveURL: fmt.Sprintf("/filters/%s/dimensions/%s/remove/%s", filter.FilterID, name, valueIDmap[val]),
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
func CreatePreviewPage(req *http.Request, dimensions []filter.ModelDimension, filter filter.Model, dst dataset.DatasetDetails, filterOutputID, datasetID, releaseDate, apiRouterVersion string, enableDatasetPreivew bool, lang string) previewPage.Page {
	var p previewPage.Page
	p.Metadata.Title = "Preview and Download"
	p.BetaBannerEnabled = true
	p.EnableDatasetPreview = enableDatasetPreivew
	p.Language = lang

	mapCookiePreferences(req, &p.CookiesPreferencesSet, &p.CookiesPolicy)

	ctx := req.Context()
	log.Event(ctx, "mapping api responses to preview page model", log.INFO, log.Data{"filterOutputID": filterOutputID, "datasetID": datasetID})

	p.SearchDisabled = false
	p.ReleaseDate = releaseDate
	p.Data.UnitOfMeasurement = dst.UnitOfMeasure

	versionURL, err := url.Parse(filter.Links.Version.HRef)
	if err != nil {
		log.Event(ctx, "unable to parse version url", log.WARN, log.Error(err))
	}
	versionPath := strings.TrimPrefix(versionURL.Path, apiRouterVersion)

	p.Data.CurrentVersionURL = versionPath

	p.IsInFilterBreadcrumb = true

	_, edition, _, _ := helpers.ExtractDatasetInfoFromPath(versionPath)

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dst.Title,
		URI:   fmt.Sprintf("/datasets/%s/editions", dst.ID),
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: edition,
		URI:   versionPath,
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter options",
		URI:   fmt.Sprintf("/filters/%s/dimensions", filter.Links.FilterBlueprint.ID),
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Preview",
	})

	p.Data.FilterID = filter.Links.FilterBlueprint.ID
	p.Data.FilterOutputID = filterOutputID

	p.DatasetTitle = dst.Title
	p.Data.DatasetID = datasetID
	p.DatasetId = datasetID
	_, p.Data.Edition, _, _ = helpers.ExtractDatasetInfoFromPath(versionPath)

	for ext, d := range filter.Downloads {
		p.Data.Downloads = append(p.Data.Downloads, previewPage.Download{
			Extension: ext,
			Size:      d.Size,
			URI:       d.URL,
			Skipped:   d.Skipped,
		})
	}

	for _, dim := range dimensions {
		p.Data.Dimensions = append(p.Data.Dimensions, previewPage.Dimension{
			Name:   dim.Name,
			Values: dim.Values,
		})
	}
	if enableDatasetPreivew && p.Data.Dimensions == nil {
		p.NoDimensionData = true
	}

	return p
}

func getNameIDLookup(vals dataset.Options) map[string]string {
	lookup := make(map[string]string)
	for _, val := range vals.Items {
		lookup[val.Label] = val.Option
	}
	return lookup
}

func getIDNameLookup(vals dataset.Options) map[string]string {
	lookup := make(map[string]string)
	for _, val := range vals.Items {
		lookup[val.Option] = val.Label
	}
	return lookup
}

func getIDNameLookupFromHierarchy(vals hierarchyClient.Model) map[string]string {
	lookup := make(map[string]string)
	for _, val := range vals.Children {
		lookup[val.Links.Self.ID] = val.Label
	}
	return lookup
}

// CreateAgePage creates an age selector page based on api responses
func CreateAgePage(req *http.Request, f filter.Model, d dataset.DatasetDetails, v dataset.Version, allVals dataset.Options, selVals []filter.DimensionOption, dims dataset.VersionDimensions, datasetID, apiRouterVersion, lang string) (age.Page, error) {
	var p age.Page
	p.BetaBannerEnabled = true

	mapCookiePreferences(req, &p.CookiesPreferencesSet, &p.CookiesPolicy)

	log.Event(req.Context(), "mapping api responses to age page model", log.INFO, log.Data{"filterID": f.FilterID, "datasetID": datasetID})

	for _, dim := range dims.Items {
		if dim.Name == "age" {
			p.Metadata.Description = dim.Description
		}
	}

	p.FilterID = f.FilterID
	p.SearchDisabled = true
	p.DatasetId = datasetID
	p.Language = lang

	versionURL, err := url.Parse(f.Links.Version.HRef)
	if err != nil {
		return p, err
	}
	versionPath := strings.TrimPrefix(versionURL.Path, apiRouterVersion)

	p.IsInFilterBreadcrumb = true

	_, edition, _, _ := helpers.ExtractDatasetInfoFromPath(versionPath)

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: d.Title,
		URI:   fmt.Sprintf("/datasets/%s/editions", d.ID),
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: edition,
		URI:   versionPath,
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter options",
		URI:   fmt.Sprintf("/filters/%s/dimensions", f.FilterID),
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Age",
	})

	p.Metadata.Title = "Age"
	p.DatasetTitle = d.Title

	p.Data.FormAction.URL = fmt.Sprintf("/filters/%s/dimensions/age/update", f.FilterID)

	var ages []int
	labelIDs := getNameIDLookup(allVals)
	for _, age := range allVals.Items {
		ageInt, err := strconv.Atoi(age.Label)
		if err != nil {
			if strings.Contains(age.Label, "+") {
				p.Data.Oldest = age.Label
			} else {
				p.Data.HasAllAges = true
				p.Data.AllAgesOption = age.Option
			}
			continue
		}
		ages = append(ages, ageInt)
	}

	sort.Ints(ages)

	for _, ageVal := range ages {
		var isSelected bool
		ageString := strconv.Itoa(ageVal)
		for _, selVal := range selVals {
			if selVal.Option == labelIDs[ageString] {
				isSelected = true
			}
		}

		p.Data.Ages = append(p.Data.Ages, age.Value{
			Option:     labelIDs[ageString],
			Label:      ageString,
			IsSelected: isSelected,
		})
	}

	p.Data.Youngest = strconv.Itoa(ages[0])

	if len(p.Data.Oldest) > 0 {
		var isSelected bool
		for _, selVal := range selVals {
			if selVal.Option == labelIDs[p.Data.Oldest] {
				isSelected = true
			}
		}

		p.Data.Ages = append(p.Data.Ages, age.Value{
			Option:     labelIDs[p.Data.Oldest],
			Label:      p.Data.Oldest,
			IsSelected: isSelected,
		})
	} else {
		p.Data.Oldest = strconv.Itoa(ages[len(ages)-1])
	}

	p.Data.CheckedRadio = "range"

	for i, val := range p.Data.Ages {
		if val.IsSelected {
			for j := i; j < len(p.Data.Ages); j++ {
				if p.Data.Ages[j].IsSelected {
					continue
				} else {
					for k := j; k < len(p.Data.Ages); k++ {
						if p.Data.Ages[k].IsSelected {
							p.Data.CheckedRadio = "list"
							break
						}
					}
				}
			}
		}
	}

	if p.Data.CheckedRadio == "range" {
		for _, val := range p.Data.Ages {
			if val.IsSelected {
				if len(p.Data.FirstSelected) == 0 {
					p.Data.FirstSelected = val.Label
				}
				p.Data.LastSelected = val.Label
			}
		}
	}

	return p, nil
}

// CreateTimePage will create a time selector page based on api response models
func CreateTimePage(req *http.Request, f filter.Model, d dataset.DatasetDetails, v dataset.Version, allVals dataset.Options, selVals []filter.DimensionOption, dims dataset.VersionDimensions, datasetID, apiRouterVersion, lang string) (timeModel.Page, error) {
	var p timeModel.Page
	p.BetaBannerEnabled = true

	mapCookiePreferences(req, &p.CookiesPreferencesSet, &p.CookiesPolicy)

	ctx := req.Context()

	if _, err := time.Parse("Jan-06", allVals.Items[0].Option); err == nil {
		p.Data.Type = "month"
	}

	p.DatasetTitle = d.Title
	p.FilterID = f.FilterID
	p.SearchDisabled = true
	p.DatasetId = datasetID
	p.Language = lang

	for _, dim := range dims.Items {
		if dim.Name == "time" {
			p.Metadata.Description = dim.Description
		}
	}

	versionURL, err := url.Parse(f.Links.Version.HRef)
	if err != nil {
		return p, err
	}
	versionPath := strings.TrimPrefix(versionURL.Path, apiRouterVersion)

	p.IsInFilterBreadcrumb = true

	_, edition, _, _ := helpers.ExtractDatasetInfoFromPath(versionPath)

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: d.Title,
		URI:   fmt.Sprintf("/datasets/%s/editions", d.ID),
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: edition,
		URI:   versionPath,
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter options",
		URI:   fmt.Sprintf("/filters/%s/dimensions", f.FilterID),
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Time",
	})

	p.Metadata.Title = "Time"

	lookup := getNameIDLookup(allVals)

	var allTimes []string
	for _, val := range allVals.Items {
		allTimes = append(allTimes, val.Label)
	}

	times, err := dates.ConvertToReadable(allTimes)
	if err != nil {
		return p, err
	}

	times = dates.Sort(times)

	p.Data.FirstTime = timeModel.Value{
		Option: lookup[times[0].Format("Jan-06")],
		Month:  times[0].Month().String(),
		Year:   fmt.Sprintf("%d", times[0].Year()),
	}

	p.Data.LatestTime = timeModel.Value{
		Option: lookup[times[len(times)-1].Format("Jan-06")],
		Month:  times[len(times)-1].Month().String(),
		Year:   fmt.Sprintf("%d", times[len(times)-1].Year()),
	}

	firstYear := times[0].Year()
	lastYear := times[len(times)-1].Year()
	diffYears := lastYear - firstYear

	p.Data.Years = append(p.Data.Years, "Select")
	for i := 0; i < diffYears+1; i++ {
		p.Data.Years = append(p.Data.Years, fmt.Sprintf("%d", firstYear+i))
	}

	p.Data.Months = append(p.Data.Months, "Select")
	for i := 0; i < 12; i++ {
		p.Data.Months = append(p.Data.Months, time.Month(i+1).String())
	}

	// Reverse times so latest is first
	for i, j := 0, len(times)-1; i < j; i, j = i+1, j-1 {
		times[i], times[j] = times[j], times[i]
	}

	for _, val := range times {
		var isSelected bool
		for _, selVal := range selVals {
			if val.Format("Jan-06") == selVal.Option {
				isSelected = true
			}
		}

		p.Data.Values = append(p.Data.Values, timeModel.Value{
			Option:     lookup[val.Format("Jan-06")],
			Month:      val.Month().String(),
			Year:       fmt.Sprintf("%d", val.Year()),
			IsSelected: isSelected,
		})
	}

	p.Data.FormAction = timeModel.Link{
		URL: fmt.Sprintf("/filters/%s/dimensions/time/update", f.FilterID),
	}

	if len(selVals) == 1 && p.Data.Values[0].IsSelected {
		p.Data.CheckedRadio = "latest"
	} else if len(selVals) == 1 {
		p.Data.CheckedRadio = "single"
		date, err := time.Parse("Jan-06", selVals[0].Option)
		if err != nil {
			log.Event(ctx, "unable to parse date", log.WARN, log.Error(err))
		}
		p.Data.SelectedStartMonth = date.Month().String()
		p.Data.SelectedStartYear = fmt.Sprintf("%d", date.Year())
	} else if len(selVals) == 0 {
		p.Data.CheckedRadio = ""
	} else if len(selVals) == len(allVals.Items) {
		p.Data.CheckedRadio = "list"
	} else {
		p.Data.CheckedRadio = "range"

		for i, val := range p.Data.Values {
			if val.IsSelected {
				for j := i; j < len(p.Data.Values); j++ {
					if p.Data.Values[j].IsSelected {
						continue
					} else {
						for k := j; k < len(p.Data.Values); k++ {
							if p.Data.Values[k].IsSelected {
								p.Data.CheckedRadio = "list"
								break
							}
						}
					}
				}
			}
		}
	}

	if p.Data.CheckedRadio == "range" {
		var selOptions []string
		for _, val := range selVals {
			selOptions = append(selOptions, val.Option)
		}

		selDates, err := dates.ConvertToReadable(selOptions)
		if err != nil {
			log.Event(ctx, "unable to convert dates to human readable values", log.WARN, log.Error(err))
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
			log.Event(ctx, "unable to convert date to month value", log.ERROR, log.Error(err))
			continue
		}
		monthStr := month.Format("January")
		_, found := fdHelpers.StringInSlice(monthStr, selectedMonths)
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
			log.Event(ctx, "unable to convert year string to int for comparison", log.ERROR, log.Error(err))
			continue
		}
		maxYearInt, err := strconv.Atoi(maxYear)
		if err != nil {
			log.Event(ctx, "unable to convert max year string to int for comparison", log.ERROR, log.Error(err))
			continue
		}
		minYearInt, err := strconv.Atoi(minYear)
		if err != nil {
			log.Event(ctx, "unable to convert min year string to int for comparison", log.ERROR, log.Error(err))
			continue
		}
		if yearInt > maxYearInt {
			maxYear = yearStr
		} else if yearInt < minYearInt {
			minYear = yearStr
		}
	}
	var listOfAllMonths []timeModel.Month
	numberOfMonthsInAYear := 12
	for i := 0; i < numberOfMonthsInAYear; i++ {
		monthName := time.Month(i + 1).String()
		_, isSelected := fdHelpers.StringInSlice(monthName, selectedMonths)
		singleMonth := timeModel.Month{
			Name:       monthName,
			IsSelected: isSelected,
		}
		listOfAllMonths = append(listOfAllMonths, singleMonth)
	}
	GroupedSelection := timeModel.GroupedSelection{
		Months:    listOfAllMonths,
		YearStart: minYear,
		YearEnd:   maxYear,
	}
	p.Data.GroupedSelection = GroupedSelection

	return p, nil
}

// CreateHierarchySearchPage forms a search page based on various api response models
func CreateHierarchySearchPage(req *http.Request, items []search.Item, dst dataset.DatasetDetails, f filter.Model, selectedIDLabelMap map[string]string, dims []dataset.VersionDimension, name, curPath, datasetID, releaseDate, referrer, query, apiRouterVersion, lang string) hierarchy.Page {
	var p hierarchy.Page
	p.BetaBannerEnabled = true

	mapCookiePreferences(req, &p.CookiesPreferencesSet, &p.CookiesPolicy)

	ctx := req.Context()
	log.Event(ctx, "mapping api response models to hierarchy search page", log.INFO, log.Data{"filterID": f.FilterID, "datasetID": datasetID, "name": name})

	pageTitle := strings.Title(name)
	for _, dim := range dims {
		if dim.Name == name && len(dim.Label) > 0 {
			pageTitle = dim.Label
		}
	}
	p.DatasetTitle = dst.Title
	p.Data.DimensionName = pageTitle
	p.DatasetId = datasetID
	p.Data.IsSearchResults = true
	p.Data.Query = query
	p.Language = lang

	title := pageTitle

	p.IsInFilterBreadcrumb = true
	var ok bool
	if p.Type, ok = hierarchyBrowseLookup[name]; !ok {
		p.Type = "type"
	}

	p.SearchDisabled = true

	p.Data.SearchURL = fmt.Sprintf("/filters/%s/dimensions/%s/search", f.FilterID, name)

	versionURL, err := url.Parse(f.Links.Version.HRef)
	if err != nil {
		log.Event(ctx, "unable to parse version url", log.WARN, log.Error(err))
	}
	versionPath := strings.TrimPrefix(versionURL.Path, apiRouterVersion)

	p.Data.LandingPageURL = versionPath + "#id-dimensions"
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dst.Title,
		URI:   versionPath,
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter options",
		URI:   fmt.Sprintf("/filters/%s/dimensions", f.FilterID),
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: title,
		URI:   fmt.Sprintf("/filters/%s/dimensions/%s", f.FilterID, name),
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Search results",
	})

	p.FilterID = f.FilterID
	p.Data.Title = title
	p.Metadata.Title = title

	p.Data.GoBack.URL = referrer

	p.Data.AddAllFilters.URL = curPath + "/add-all"
	p.Data.RemoveAll.URL = curPath + "/remove-all"

	for option, label := range selectedIDLabelMap {
		p.Data.FiltersAdded = append(p.Data.FiltersAdded, hierarchy.Filter{
			Label:     label,
			RemoveURL: fmt.Sprintf("%s/remove/%s", curPath, option),
			ID:        option,
		})
	}

	if len(items) == 0 {
		p.Data.IsSearchError = true
	} else {

		for _, item := range items {
			_, selected := selectedIDLabelMap[item.Code]
			p.Data.FilterList = append(p.Data.FilterList, hierarchy.List{
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

// CreateHierarchyPage maps data items from API responses to form a hirearchy page
func CreateHierarchyPage(req *http.Request, h hierarchyClient.Model, dst dataset.DatasetDetails, f filter.Model, selVals filter.DimensionOptions, dims dataset.VersionDimensions, name, curPath, datasetID, releaseDate, apiRouterVersion, lang string) hierarchy.Page {
	var p hierarchy.Page
	p.BetaBannerEnabled = true
	p.Language = lang

	mapCookiePreferences(req, &p.CookiesPreferencesSet, &p.CookiesPolicy)

	ctx := req.Context()
	log.Event(ctx, "mapping api response models to hierarchy page", log.INFO, log.Data{"filterID": f.FilterID, "datasetID": datasetID, "label": h.Label})

	pageTitle := strings.Title(name)
	for _, dim := range dims.Items {
		if dim.Name == name {
			p.Metadata.Description = dim.Description
			if len(dim.Label) > 0 {
				pageTitle = dim.Label
			}
		}
	}

	p.DatasetTitle = dst.Title
	p.Data.DimensionName = pageTitle
	p.DatasetId = datasetID

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
		log.Event(ctx, "unable to parse version url", log.WARN, log.Error(err))
	}
	versionPath := strings.TrimPrefix(versionURL.Path, apiRouterVersion)

	p.IsInFilterBreadcrumb = true

	_, edition, _, _ := helpers.ExtractDatasetInfoFromPath(versionPath)

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dst.Title,
		URI:   fmt.Sprintf("/datasets/%s/editions", dst.ID),
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: edition,
		URI:   versionPath,
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter options",
		URI:   fmt.Sprintf("/filters/%s/dimensions", f.FilterID),
	})

	if len(h.Breadcrumbs) > 0 {
		if name == "geography" {
			p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
				Title: "Geographic Areas",
				URI:   fmt.Sprintf("/filters/%s/dimensions/%s", f.FilterID, "geography"),
			})

			if !topLevelGeographies[h.Links.Code.ID] {
				for i := len(h.Breadcrumbs) - 1; i >= 0; i-- {
					breadcrumb := h.Breadcrumbs[i]

					if !topLevelGeographies[breadcrumb.Links.Code.ID] {
						var url string
						if breadcrumb.Links.Code.ID != "" {
							url = fmt.Sprintf("/filters/%s/dimensions/%s/%s", f.FilterID, name, breadcrumb.Links.Code.ID)
						} else {
							url = fmt.Sprintf("/filters/%s/dimensions/%s", f.FilterID, name)
						}

						p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
							Title: breadcrumb.Label,
							URI:   url,
						})
					}
				}
			}
		} else {

			for i := len(h.Breadcrumbs) - 1; i >= 0; i-- {
				breadcrumb := h.Breadcrumbs[i]

				var url string
				if breadcrumb.Links.Code.ID != "" {
					url = fmt.Sprintf("/filters/%s/dimensions/%s/%s", f.FilterID, name, breadcrumb.Links.Code.ID)
				} else {
					url = fmt.Sprintf("/filters/%s/dimensions/%s", f.FilterID, name)
				}

				p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
					Title: breadcrumb.Label,
					URI:   url,
				})
			}

		}
	}

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: title,
	})

	p.FilterID = f.FilterID
	p.Data.Title = title
	p.Metadata.Title = fmt.Sprintf("Filter Options - %s", title)

	if len(h.Breadcrumbs) > 0 {
		if len(h.Breadcrumbs) == 1 || topLevelGeographies[h.Breadcrumbs[0].Links.Code.ID] && name == "geography" {
			p.Data.Parent = pageTitle
			p.Data.GoBack = hierarchy.Link{
				URL: fmt.Sprintf("/filters/%s/dimensions/%s", f.FilterID, name),
			}
		} else {
			p.Data.Parent = h.Breadcrumbs[0].Label
			p.Data.GoBack = hierarchy.Link{
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

	idLabelMap := getIDNameLookupFromHierarchy(h)

	for _, val := range selVals.Items {
		p.Data.FiltersAdded = append(p.Data.FiltersAdded, hierarchy.Filter{
			Label:     idLabelMap[val.Option],
			RemoveURL: fmt.Sprintf("%s/remove/%s", curPath, val.Option),
			ID:        val.Option,
		})
	}

	if h.HasData && len(h.Breadcrumbs) == 0 {
		var selected bool
		for _, val := range selVals.Items {
			if val.Option == h.Links.Code.ID {
				selected = true
			}
		}

		p.Data.FilterList = append(p.Data.FilterList, hierarchy.List{
			Label:    h.Label,
			ID:       h.Links.Code.ID,
			SubNum:   "0",
			SubURL:   "",
			Selected: selected,
			HasData:  true,
		})
	}

	for _, child := range h.Children {
		var selected bool
		for _, val := range selVals.Items {
			if val.Option == child.Links.Code.ID {
				selected = true
			}
		}
		p.Data.FilterList = append(p.Data.FilterList, hierarchy.List{
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

func reverseInts(input []int) []int {
	if len(input) == 0 {
		return input
	}
	return append(reverseInts(input[1:]), input[0])
}

// mapCookiePreferences reads cookie policy and preferences cookies and then maps the values to the page model
func mapCookiePreferences(req *http.Request, preferencesIsSet *bool, policy *model.CookiesPolicy) {
	preferencesCookie := cookies.GetCookiePreferences(req)
	*preferencesIsSet = preferencesCookie.IsPreferenceSet
	*policy = model.CookiesPolicy{
		Essential: preferencesCookie.Policy.Essential,
		Usage:     preferencesCookie.Policy.Usage,
	}
}
