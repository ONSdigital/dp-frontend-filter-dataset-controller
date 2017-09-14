package mapper

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/dates"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/filterOverview"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/hierarchy"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/listSelector"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/previewPage"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/rangeSelector"
	"github.com/ONSdigital/go-ns/clients/codelist"
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
	"CPI":                "Goods and Services", // This is temporary until the codelist api is updated
}

// CreateFilterOverview maps data items from API responses to form a filter overview
// front end page model
func CreateFilterOverview(dimensions []filter.ModelDimension, filter filter.Model, dst dataset.Model, filterID, datasetID, releaseDate string) filterOverview.Page {
	var p filterOverview.Page

	p.FilterID = filterID
	p.Metadata.Title = "Filter Options"

	disableButton := true

	for _, d := range dimensions {
		var fod filterOverview.Dimension

		if len(d.Values) > 0 {
			disableButton = false
		}

		if d.Name == "time" {
			var selectedDates []string
			for _, val := range d.Values {
				selectedDates = append(selectedDates, val)
			}

			selectedDats, _ := dates.ConvertToReadable(selectedDates)
			selectedDats = dates.Sort(selectedDats)

			for _, ac := range selectedDats {
				fod.AddedCategories = append(fod.AddedCategories, dates.ConvertToMonthYear(ac))
			}
		} else {
			for _, ac := range d.Values {
				fod.AddedCategories = append(fod.AddedCategories, ac)
			}
		}

		if d.Hierarchy.ID != "" {
			fod.Link.URL = fmt.Sprintf("/filters/%s/dimensions/%s/%s", filterID, d.Name, d.Hierarchy.ID)
		} else {
			fod.Link.URL = fmt.Sprintf("/filters/%s/dimensions/%s", filterID, d.Name)
		}
		if len(fod.AddedCategories) > 0 {
			fod.Link.Label = "Filter"
		} else {
			fod.Link.Label = "Please select"
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

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dst.Title,
		URI:   filter.DatasetFilterID,
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter this dataset",
	})

	p.Metadata.Footer = model.Footer{
		Enabled:     true,
		Contact:     dst.Contact.Name,
		ReleaseDate: releaseDate,
		NextRelease: dst.NextRelease,
		DatasetID:   datasetID,
	}

	return p
}

// CreateListSelectorPage maps items from API responses to form the model for a
// dimension list selector page
func CreateListSelectorPage(name string, selectedValues []filter.DimensionOption, allValues codelist.DimensionValues, filter filter.Model, dst dataset.Model, datasetID, releaseDate string) listSelector.Page {
	var p listSelector.Page

	p.SearchDisabled = true
	p.FilterID = filter.FilterID
	p.Data.Title = dimensionTitleTranslator[name]

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dst.Title,
		URI:   filter.DatasetFilterID,
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter this dataset",
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

	lookup := getIDNameLookup(allValues.Items)

	var selectedListValues, selectedListIDs []string
	for _, opt := range selectedValues {
		selectedListValues = append(selectedListValues, lookup[opt.Option])
		selectedListIDs = append(selectedListIDs, opt.Option)
	}

	var allListValues, allListIDs []string
	valueIDmap := make(map[string]string)
	for _, val := range allValues.Items {
		allListValues = append(allListValues, val.Label)
		allListIDs = append(allListIDs, val.ID)
		valueIDmap[val.Label] = val.ID
	}

	if name == "time" {
		dats, err := dates.ConvertToReadable(allListValues)
		if err != nil {
			log.Error(err, nil)
		}

		dats = dates.Sort(dats)

		selectedDats, err := dates.ConvertToReadable(selectedListValues)
		if err != nil {
			log.Error(err, nil)
		}

		selectedDats = dates.Sort(selectedDats)

		for _, val := range selectedDats {
			p.Data.FiltersAdded = append(p.Data.FiltersAdded, listSelector.Filter{
				RemoveURL: fmt.Sprintf("/filters/%s/dimensions/%s/remove/%s?selectorType=list", filter.FilterID, name, valueIDmap[fmt.Sprintf("%d.%02d", val.Year(), val.Month())]),
				Label:     dates.ConvertToMonthYear(val),
				ID:        valueIDmap[fmt.Sprintf("%d.%02d", val.Year(), val.Month())],
			})
		}

		for _, val := range dats {
			var isSelected bool
			for _, selDat := range selectedDats {
				if selDat.Equal(val) {
					isSelected = true
				}
			}

			p.Data.RangeData.Values = append(p.Data.RangeData.Values, listSelector.Value{
				Label:      dates.ConvertToMonthYear(val),
				ID:         valueIDmap[fmt.Sprintf("%d.%02d", val.Year(), val.Month())],
				IsSelected: isSelected,
			})
		}

	}

	if len(allListValues) == len(selectedListValues) {
		p.Data.AddAllChecked = true
	}

	p.Data.FiltersAmount = len(selectedListValues)

	p.Metadata.Footer = model.Footer{
		Enabled:     true,
		Contact:     dst.Contact.Name,
		ReleaseDate: releaseDate,
		NextRelease: dst.NextRelease,
		DatasetID:   datasetID,
	}

	return p
}

// CreateRangeSelectorPage maps items from API responses to form a dimension range
// selector page model
func CreateRangeSelectorPage(name string, selectedValues []filter.DimensionOption, allValues codelist.DimensionValues, filter filter.Model, dst dataset.Model, datasetID, releaseDate string) rangeSelector.Page {
	var p rangeSelector.Page

	p.SearchDisabled = true

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dst.Title,
		URI:   filter.DatasetFilterID,
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter this dataset",
		URI:   fmt.Sprintf("/filters/%s/dimensions", filter.FilterID),
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dimensionTitleTranslator[name],
	})

	p.Data.Title = dimensionTitleTranslator[name]
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
			if val.ID == opt.Option {
				selectedRangeValues = append(selectedRangeValues, val.Label)
				selectedRangeIDs = append(selectedRangeIDs, val.ID)
			}
		}
	}

	p.Data.FiltersAmount = len(selectedRangeValues)

	for _, val := range allValues.Items {
		p.Data.RangeData.Values = append(p.Data.RangeData.Values, val.Label)
	}

	if name == "time" {
		selectedDats, err := dates.ConvertToReadable(selectedRangeValues)
		if err != nil {
			log.Error(err, nil)
		}

		timeIDLookup := make(map[time.Time]string)
		for i, dat := range selectedDats {
			timeIDLookup[dat] = selectedRangeIDs[i]
		}

		selectedDats = dates.Sort(selectedDats)

		for _, val := range selectedDats {
			p.Data.FiltersAdded = append(p.Data.FiltersAdded, rangeSelector.Filter{
				RemoveURL: fmt.Sprintf("/filters/%s/dimensions/%s/remove/%s", filter.FilterID, name, timeIDLookup[val]),
				Label:     dates.ConvertToMonthYear(val),
				ID:        timeIDLookup[val],
			})
		}

		allDats, err := dates.ConvertToReadable(p.Data.RangeData.Values)
		if err != nil {
			log.Error(err, nil)
		}

		allDats = dates.Sort(allDats)

		firstYear := allDats[0].Year()
		lastYear := allDats[len(allDats)-1].Year()
		diffYears := lastYear - firstYear

		for i := 0; i < diffYears+1; i++ {
			p.Data.DateRangeData.YearValues = append(p.Data.DateRangeData.YearValues, fmt.Sprintf("%d", firstYear+i))
		}

		for i := 0; i < 12; i++ {
			p.Data.DateRangeData.MonthValues = append(p.Data.DateRangeData.MonthValues, time.Month(i+1).String())
		}
	}

	p.Data.RangeData.URL = fmt.Sprintf("/filters/%s/dimensions/%s/range", filter.FilterID, name)

	p.Data.RemoveAll.URL = fmt.Sprintf("/filters/%s/dimensions/%s/remove-all", filter.FilterID, name)

	p.Data.RangeData.StartLabel = "Start"
	p.Data.RangeData.EndLabel = "End"
	p.Data.RangeData.URL = fmt.Sprintf("/filters/%s/dimensions/%s/range", filter.FilterID, name)

	p.Metadata.Footer = model.Footer{
		Enabled:     true,
		Contact:     dst.Contact.Name,
		ReleaseDate: releaseDate,
		NextRelease: dst.NextRelease,
		DatasetID:   datasetID,
	}

	return p

}

// CreatePreviewPage maps data items from API responses to create a preview page
func CreatePreviewPage(dimensions []filter.ModelDimension, filter filter.Model, dst dataset.Model, filterID, datasetID, releaseDate string) previewPage.Page {
	var p previewPage.Page

	p.SearchDisabled = true

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dst.Title,
		URI:   filter.DatasetFilterID,
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter this dataset",
		URI:   fmt.Sprintf("/filters/%s/dimensions", filterID),
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Preview",
	})

	p.Data.FilterID = filterID

	p.Metadata.Footer = model.Footer{
		Enabled:     true,
		Contact:     dst.Contact.Name,
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

func getNameIDLookup(vals []codelist.Item) map[string]string {
	lookup := make(map[string]string)
	for _, val := range vals {
		lookup[val.Name] = val.ID
	}
	return lookup
}

func getIDNameLookup(vals []codelist.Item) map[string]string {
	lookup := make(map[string]string)
	for _, val := range vals {
		lookup[val.ID] = val.Label
	}
	return lookup
}

// CreateHierarchyPage maps data items from API responses to form a hirearchy page
func CreateHierarchyPage(h hierarchyClient.Model, parents []hierarchyClient.Parent, dst dataset.Model, f filter.Model, curPath, dimensionTitle, datasetID, releaseDate string) hierarchy.Page {
	var p hierarchy.Page

	var title string
	if len(parents) == 0 {
		title = dimensionTitleTranslator[dimensionTitle]
	} else {
		title = h.Label
	}

	p.SearchDisabled = true

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dst.Title,
		URI:   f.DatasetFilterID,
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter this dataset",
		URI:   fmt.Sprintf("/filters/%s/dimensions", f.FilterID),
	})
	for i, par := range parents {
		var breadrumbTitle string
		if i != 0 {
			breadrumbTitle = par.Label
		} else {
			breadrumbTitle = dimensionTitleTranslator[dimensionTitle]
		}
		p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
			Title: breadrumbTitle,
			URI:   fmt.Sprintf("/filters/%s%s", f.FilterID, par.URL),
		})
	}
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: title,
	})

	p.FilterID = f.FilterID
	p.Data.Title = title
	if len(parents) > 0 {
		if len(parents) == 1 {
			p.Data.Parent = dimensionTitleTranslator[dimensionTitle]
		} else {
			p.Data.Parent = parents[len(parents)-1].Label
		}
		p.Data.GoBack = hierarchy.Link{
			URL: fmt.Sprintf("/filters/%s%s", f.FilterID, parents[len(parents)-1].URL),
		}
	}

	p.Data.AddAllFilters.Amount = strconv.Itoa(len(h.Children))
	p.Data.AddAllFilters.URL = curPath + "/add-all"
	p.Data.RemoveAll.URL = curPath + "/remove-all"

	for _, dim := range f.Dimensions {
		if dim.Name == dimensionTitle {
			for i, val := range dim.Values {
				p.Data.FiltersAdded = append(p.Data.FiltersAdded, hierarchy.Filter{
					Label:     val,
					RemoveURL: fmt.Sprintf("%s/remove/%s", curPath, dim.IDs[i]),
					ID:        dim.IDs[i],
				})
			}
		}
	}

	for _, child := range h.Children {
		var selected bool
		for _, dim := range f.Dimensions {
			if dim.Name == dimensionTitle {
				for _, id := range dim.IDs {
					if id == child.ID {
						selected = true
					}
				}
			}
		}
		p.Data.FilterList = append(p.Data.FilterList, hierarchy.List{
			Label:    child.Label,
			ID:       child.ID,
			SubNum:   strconv.Itoa(child.NumberofChildren),
			SubURL:   fmt.Sprintf("redirect:/filters/%s%s", f.FilterID, child.URL),
			Selected: selected,
		})

	}

	//p.Data.Metadata = hierarchy.Metadata(met)
	p.Data.SaveAndReturn.URL = curPath + "/update"
	p.Data.Cancel.URL = fmt.Sprintf("/filters/%s/dimensions", f.FilterID)

	p.Metadata.Footer = model.Footer{
		Enabled:     true,
		Contact:     dst.Contact.Name,
		ReleaseDate: releaseDate,
		NextRelease: dst.NextRelease,
		DatasetID:   datasetID,
	}

	return p
}
