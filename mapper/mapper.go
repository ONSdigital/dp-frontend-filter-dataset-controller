package mapper

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/dates"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/filterOverview"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/hierarchy"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/listSelector"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/previewPage"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/rangeSelector"
	timeModel "github.com/ONSdigital/dp-frontend-models/model/dataset-filter/time"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/clients/filter"
	hierarchyClient "github.com/ONSdigital/go-ns/clients/hierarchy"
	"github.com/ONSdigital/go-ns/log"
)

var dimensionTitleTranslator = map[string]string{
	"geography":          "Geographic Areas",
	"year":               "Year",
	"age-range":          "Age",
	"sex":                "Sex",
	"month":              "Month",
	"time":               "Time",
	"goods-and-services": "Goods and Services",
	"aggregate":          "Goods and Services",
}

// CreateFilterOverview maps data items from API responses to form a filter overview
// front end page model
func CreateFilterOverview(dimensions []filter.ModelDimension, filter filter.Model, dst dataset.Model, filterID, datasetID, releaseDate string) filterOverview.Page {
	var p filterOverview.Page

	log.Debug("mapping api response models into filter overview page model", log.Data{"filterID": filterID, "datasetID": datasetID})

	p.FilterID = filterID
	p.Metadata.Title = "Filter Options"
	p.TaxonomyDomain = os.Getenv("TAXONOMY_DOMAIN")

	disableButton := true

	for _, d := range dimensions {
		var fod filterOverview.Dimension

		if len(d.Values) > 0 {
			disableButton = false
		}

		if d.Name == "time" {
			times, err := dates.ConvertToReadable(d.Values)
			if err != nil {
				log.Error(err, nil)
			}

			times = dates.Sort(times)
			for _, time := range times {
				fod.AddedCategories = append(fod.AddedCategories, time.Format("January 2006"))
			}
		} else {
			for _, ac := range d.Values {
				fod.AddedCategories = append(fod.AddedCategories, ac)
			}
		}

		fod.Link.URL = fmt.Sprintf("/filters/%s/dimensions/%s", filterID, d.Name)

		if len(fod.AddedCategories) > 0 {
			fod.Link.Label = "Filter"
		} else {
			fod.Link.Label = "Add at least one"
		}

		fod.Filter = dimensionTitleTranslator[d.Name]

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
		log.Error(err, nil)
	}

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dst.Title,
		URI:   versionURL.Path,
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter options",
	})

	var name string
	if len(dst.Contacts) > 0 {
		name = dst.Contacts[0].Name
	}

	p.Metadata.Footer = model.Footer{
		Enabled:     true,
		Contact:     name,
		ReleaseDate: releaseDate,
		NextRelease: dst.NextRelease,
		DatasetID:   datasetID,
	}

	return p
}

// CreateListSelectorPage maps items from API responses to form the model for a
// dimension list selector page
func CreateListSelectorPage(name string, selectedValues []filter.DimensionOption, allValues dataset.Options, filter filter.Model, dst dataset.Model, datasetID, releaseDate string) listSelector.Page {
	var p listSelector.Page

	log.Debug("mapping api response models to list selector page model", log.Data{"filterID": filter.FilterID, "datasetID": datasetID, "dimension": name})

	p.SearchDisabled = true
	p.FilterID = filter.FilterID
	p.Data.Title = dimensionTitleTranslator[name]
	p.Metadata.Title = dimensionTitleTranslator[name]
	p.TaxonomyDomain = os.Getenv("TAXONOMY_DOMAIN")

	versionURL, err := url.Parse(filter.Links.Version.HRef)
	if err != nil {
		log.Error(err, nil)
	}

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dst.Title,
		URI:   versionURL.Path,
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter options",
		URI:   fmt.Sprintf("/filters/%s/dimensions", filter.FilterID),
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dimensionTitleTranslator[name],
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

	p.Data.RemoveAll.URL = fmt.Sprintf("/filters/%s/dimensions/%s/remove-all?selectorType=list", filter.FilterID, name)

	lookup := getIDNameLookup(allValues)

	var selectedListValues, selectedListIDs []string
	for _, opt := range selectedValues {
		selectedListValues = append(selectedListValues, lookup[opt.Option])
		selectedListIDs = append(selectedListIDs, opt.Option)
	}

	var allListValues, allListIDs []string
	valueIDmap := make(map[string]string)
	for _, val := range allValues.Items {
		allListValues = append(allListValues, val.Label)
		allListIDs = append(allListIDs, val.Option)
		valueIDmap[val.Label] = val.Option
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
			RemoveURL: fmt.Sprintf("/filters/%s/dimensions/%s/remove/%s?selectorType=list", filter.FilterID, name, valueIDmap[val]),
			Label:     val,
			ID:        valueIDmap[val],
		})
	}

	if len(allListValues) == len(selectedListValues) {
		p.Data.AddAllChecked = true
	}

	p.Data.FiltersAmount = len(selectedListValues)

	var contactName string
	if len(dst.Contacts) > 0 {
		contactName = dst.Contacts[0].Name
	}

	p.Metadata.Footer = model.Footer{
		Enabled:     true,
		Contact:     contactName,
		ReleaseDate: releaseDate,
		NextRelease: dst.NextRelease,
		DatasetID:   datasetID,
	}

	return p
}

// CreateRangeSelectorPage maps items from API responses to form a dimension range
// selector page model
func CreateRangeSelectorPage(name string, selectedValues []filter.DimensionOption, allValues dataset.Options, filter filter.Model, dst dataset.Model, datasetID, releaseDate string) rangeSelector.Page {
	var p rangeSelector.Page

	log.Debug("mapping api response models to range selector page model", log.Data{"filterID": filter.FilterID, "datasetID": datasetID, "dimension": name})

	p.SearchDisabled = true
	p.TaxonomyDomain = os.Getenv("TAXONOMY_DOMAIN")

	versionURL, err := url.Parse(filter.Links.Version.HRef)
	if err != nil {
		log.Error(err, nil)
	}

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dst.Title,
		URI:   versionURL.Path,
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter options",
		URI:   fmt.Sprintf("/filters/%s/dimensions", filter.FilterID),
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dimensionTitleTranslator[name],
	})

	p.Data.Title = dimensionTitleTranslator[name]
	p.Metadata.Title = dimensionTitleTranslator[name]
	p.FilterID = filter.FilterID

	p.Data.AddFromList = rangeSelector.Link{
		Label: fmt.Sprintf("Add %s range", name),
		URL:   fmt.Sprintf("/filters/%s/dimensions/%s?selectorType=list", filter.FilterID, name),
	}
	p.Data.AddRange = rangeSelector.Link{
		Label: fmt.Sprintf("Add %ss", name),
		URL:   fmt.Sprintf("/filters/%s/dimensions/%s/add", filter.FilterID, name),
	}
	p.Data.AddAllInRange = rangeSelector.Link{
		Label: fmt.Sprintf("All %ss", name),
	}
	if len(selectedValues) == len(allValues.Items) {
		p.Data.AddAllChecked = true
	}
	p.Data.SaveAndReturn = rangeSelector.Link{
		URL: fmt.Sprintf("/filters/%s/dimensions", filter.FilterID),
	}
	p.Data.Cancel = rangeSelector.Link{
		URL: fmt.Sprintf("/filters/%s/dimensions", filter.FilterID),
	}
	var selectedRangeValues, selectedRangeIDs []string
	for _, opt := range selectedValues {
		for _, val := range allValues.Items {
			if val.Option == opt.Option {
				selectedRangeValues = append(selectedRangeValues, val.Label)
				selectedRangeIDs = append(selectedRangeIDs, val.Option)
			}
		}
	}

	p.Data.FiltersAmount = len(selectedRangeValues)

	for _, val := range allValues.Items {
		p.Data.RangeData.Values = append(p.Data.RangeData.Values, val.Label)
	}

	p.Data.RangeData.URL = fmt.Sprintf("/filters/%s/dimensions/%s/range", filter.FilterID, name)

	p.Data.RemoveAll.URL = fmt.Sprintf("/filters/%s/dimensions/%s/remove-all", filter.FilterID, name)

	p.Data.RangeData.StartLabel = "Start"
	p.Data.RangeData.EndLabel = "End"
	p.Data.RangeData.URL = fmt.Sprintf("/filters/%s/dimensions/%s/range", filter.FilterID, name)

	var contactName string
	if len(dst.Contacts) > 0 {
		contactName = dst.Contacts[0].Name
	}

	p.Metadata.Footer = model.Footer{
		Enabled:     true,
		Contact:     contactName,
		ReleaseDate: releaseDate,
		NextRelease: dst.NextRelease,
		DatasetID:   datasetID,
	}

	return p

}

// CreatePreviewPage maps data items from API responses to create a preview page
func CreatePreviewPage(dimensions []filter.ModelDimension, filter filter.Model, dst dataset.Model, filterID, datasetID, releaseDate string) previewPage.Page {
	var p previewPage.Page
	p.Metadata.Title = "Preview and Download"

	log.Debug("mapping api responses to preview page model", log.Data{"filterID": filterID, "datasetID": datasetID})

	p.SearchDisabled = false
	p.TaxonomyDomain = os.Getenv("TAXONOMY_DOMAIN")

	versionURL, err := url.Parse(filter.Links.Version.HRef)
	if err != nil {
		log.Error(err, nil)
	}

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dst.Title,
		URI:   versionURL.Path,
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter options",
		URI:   fmt.Sprintf("/filters/%s/dimensions", filterID),
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Preview",
	})

	p.Data.FilterID = filterID

	var contactName string
	if len(dst.Contacts) > 0 {
		contactName = dst.Contacts[0].Name
	}

	p.Metadata.Footer = model.Footer{
		Enabled:     true,
		Contact:     contactName,
		ReleaseDate: releaseDate,
		NextRelease: dst.NextRelease,
		DatasetID:   datasetID,
	}

	for ext, d := range filter.Downloads {
		p.Data.Downloads = append(p.Data.Downloads, previewPage.Download{
			Extension: ext,
			Size:      d.Size,
			URI:       d.URL,
		})
	}

	for _, dim := range dimensions {
		p.Data.Dimensions = append(p.Data.Dimensions, previewPage.Dimension{
			Name:   dim.Name,
			Values: dim.Values,
		})
	}

	p.IsContentLoaded = true

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

// CreateTimePage will create a time selector page based on api response models
func CreateTimePage(f filter.Model, d dataset.Model, v dataset.Version, allVals dataset.Options, selVals []filter.DimensionOption, datasetID string) (timeModel.Page, error) {
	var p timeModel.Page

	log.Debug("mapping api responses to time page model", log.Data{"filterID": f.FilterID, "datasetID": datasetID})

	p.FilterID = f.FilterID
	p.SearchDisabled = true
	p.TaxonomyDomain = os.Getenv("TAXONOMY_DOMAIN")

	versionURL, err := url.Parse(f.Links.Version.HRef)
	if err != nil {
		return p, err
	}

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: d.Title,
		URI:   versionURL.Path,
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter this dataset",
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

	var contactName string
	if len(d.Contacts) > 0 {
		contactName = d.Contacts[0].Name
	}

	p.Metadata.Footer = model.Footer{
		Enabled:     true,
		Contact:     contactName,
		ReleaseDate: v.ReleaseDate,
		NextRelease: d.NextRelease,
		DatasetID:   datasetID,
	}

	if len(selVals) == 1 && p.Data.Values[0].IsSelected {
		p.Data.CheckedRadio = "latest"
	} else if len(selVals) == 1 {
		p.Data.CheckedRadio = "single"
		date, err := time.Parse("Jan-06", selVals[0].Option)
		if err != nil {
			log.Error(err, nil)
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
			log.Error(err, nil)
		}

		selDates = dates.Sort(selDates)

		p.Data.SelectedStartMonth = selDates[0].Month().String()
		p.Data.SelectedStartYear = fmt.Sprintf("%d", selDates[0].Year())
		p.Data.SelectedEndMonth = selDates[len(selDates)-1].Month().String()
		p.Data.SelectedEndYear = fmt.Sprintf("%d", selDates[len(selDates)-1].Year())
	}

	return p, nil
}

// CreateHierarchyPage maps data items from API responses to form a hirearchy page
func CreateHierarchyPage(h hierarchyClient.Model, dst dataset.Model, f filter.Model, selVals []filter.DimensionOption, allVals dataset.Options, name, curPath, datasetID, releaseDate string) hierarchy.Page {
	var p hierarchy.Page

	log.Debug("mapping api response models to hierarchy page", log.Data{"filterID": f.FilterID, "datasetID": datasetID, "label": h.Label})

	p.Data.DimensionName = dimensionTitleTranslator[name]

	var title string
	if len(h.Breadcrumbs) == 0 {
		title = dimensionTitleTranslator[name]
	} else {
		title = h.Label
	}

	p.SearchDisabled = true
	p.TaxonomyDomain = os.Getenv("TAXONOMY_DOMAIN")

	versionURL, err := url.Parse(f.Links.Version.HRef)
	if err != nil {
		log.Error(err, nil)
	}

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dst.Title,
		URI:   versionURL.Path,
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter options",
		URI:   fmt.Sprintf("/filters/%s/dimensions", f.FilterID),
	})
	if len(h.Breadcrumbs) > 0 {
		for i := len(h.Breadcrumbs) - 1; i >= 0; i-- {
			breadcrumb := h.Breadcrumbs[i]

			var url string
			if breadcrumb.Links.Self.ID != "" {
				url = fmt.Sprintf("/filters/%s/dimensions/%s/%s", f.FilterID, name, breadcrumb.Links.Self.ID)
			} else {
				url = fmt.Sprintf("/filters/%s/dimensions/%s", f.FilterID, name)
			}

			p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
				Title: breadcrumb.Label,
				URI:   url,
			})
		}
	}

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: title,
	})

	p.FilterID = f.FilterID
	p.Data.Title = title
	p.Metadata.Title = title

	if len(h.Breadcrumbs) > 0 {
		if len(h.Breadcrumbs) == 1 {
			p.Data.Parent = dimensionTitleTranslator[name]
			p.Data.GoBack = hierarchy.Link{
				URL: fmt.Sprintf("/filters/%s/dimensions/%s", f.FilterID, name),
			}
		} else {
			p.Data.Parent = h.Breadcrumbs[0].Label
			p.Data.GoBack = hierarchy.Link{
				URL: fmt.Sprintf("/filters/%s/dimensions/%s/%s", f.FilterID, name, h.Breadcrumbs[0].Links.Self.ID),
			}
		}
	}

	p.Data.AddAllFilters.Amount = strconv.Itoa(len(h.Children))
	p.Data.AddAllFilters.URL = curPath + "/add-all"
	p.Data.RemoveAll.URL = curPath + "/remove-all"

	idLabelMap := getIDNameLookup(allVals)

	for _, val := range selVals {
		p.Data.FiltersAdded = append(p.Data.FiltersAdded, hierarchy.Filter{
			Label:     idLabelMap[val.Option],
			RemoveURL: fmt.Sprintf("%s/remove/%s", curPath, val.Option),
			ID:        val.Option,
		})
	}

	for _, child := range h.Children {
		var selected bool
		for _, val := range selVals {
			if val.Option == child.Links.Self.ID {
				selected = true
			}
		}

		p.Data.FilterList = append(p.Data.FilterList, hierarchy.List{
			Label:    child.Label,
			ID:       child.Links.Self.ID,
			SubNum:   strconv.Itoa(child.NumberofChildren),
			SubURL:   fmt.Sprintf("redirect:/filters/%s/dimensions/%s/%s", f.FilterID, name, child.Links.Self.ID),
			Selected: selected,
		})

	}

	//p.Data.Metadata = hierarchy.Metadata(met)
	p.Data.SaveAndReturn.URL = curPath + "/update"
	p.Data.Cancel.URL = fmt.Sprintf("/filters/%s/dimensions", f.FilterID)

	var contactName string
	if len(dst.Contacts) > 0 {
		contactName = dst.Contacts[0].Name
	}

	p.Metadata.Footer = model.Footer{
		Enabled:     true,
		Contact:     contactName,
		ReleaseDate: releaseDate,
		NextRelease: dst.NextRelease,
		DatasetID:   datasetID,
	}

	return p
}
