package mapper

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/data"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/dates"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/filterOverview"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/hierarchy"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/listSelector"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/previewPage"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/rangeSelector"
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
}

// CreateFilterOverview maps data items from API responses to form a filter overview
// front end page model
func CreateFilterOverview(dimensions []data.Dimension, filter data.Filter, dataset data.Dataset, filterID string) filterOverview.Page {
	var p filterOverview.Page

	p.FilterID = filterID

	for _, d := range dimensions {
		var fod filterOverview.Dimension

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
		fod.Link.Label = "Filter"

		fod.Filter = dimensionTitleTranslator[d.Name]

		p.Data.Dimensions = append(p.Data.Dimensions, fod)
	}

	p.Data.PreviewAndDownload.URL = fmt.Sprintf("/filters/%s", filterID)
	p.Data.Cancel.URL = "/"
	p.SearchDisabled = true

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dataset.Title,
		URI:   fmt.Sprintf("/datasets/%s/editions/%s/versions/%s", filter.Dataset, filter.Edition, filter.Version),
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter this dataset",
	})

	p.Metadata.Footer = model.Footer{
		Enabled:     true,
		Contact:     dataset.Contact.Name,
		ReleaseDate: dataset.ReleaseDate,
		NextRelease: dataset.NextRelease,
		DatasetID:   dataset.ID,
	}

	return p
}

// CreateListSelectorPage maps items from API responses to form the model for a
// dimension list selector page
func CreateListSelectorPage(name string, selectedValues, allValues data.DimensionValues, filter data.Filter, dataset data.Dataset) listSelector.Page {
	var p listSelector.Page

	p.SearchDisabled = true
	p.FilterID = filter.FilterID
	p.Data.Title = dimensionTitleTranslator[name]

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dataset.Title,
		URI:   fmt.Sprintf("/datasets/%s/editions/%s/versions/%s", filter.Dataset, filter.Edition, filter.Version),
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

	p.Data.RemoveAll.URL = fmt.Sprintf("/filters/%s/dimensions/%s/remove-all", filter.FilterID, name)

	if name == "time" {

		var origDates []string
		for _, val := range allValues.Items {
			origDates = append(origDates, val.Label)
		}

		dats, _ := dates.ConvertToReadable(origDates)
		dats = dates.Sort(dats)

		var selectedDates []string
		for _, val := range selectedValues.Items {
			selectedDates = append(selectedDates, val.Name)
		}

		selectedDats, _ := dates.ConvertToReadable(selectedDates)

		for i, val := range selectedDats {
			p.Data.FiltersAdded = append(p.Data.FiltersAdded, listSelector.Filter{
				RemoveURL: fmt.Sprintf("/filters/%s/dimensions/%s/remove/%s", filter.FilterID, name, selectedValues.Items[i].ID),
				Label:     dates.ConvertToMonthYear(val),
			})
		}

		for _, val := range dats {
			var isSelected bool
			for _, selVal := range selectedDats {
				if selVal.Equal(val) {
					isSelected = true
				}
			}

			p.Data.RangeData.Values = append(p.Data.RangeData.Values, listSelector.Value{
				Label:      dates.ConvertToMonthYear(val),
				IsSelected: isSelected,
			})
		}

	}

	p.Data.FiltersAmount = len(selectedValues.Items)

	p.Metadata.Footer = model.Footer{
		Enabled:     true,
		Contact:     dataset.Contact.Name,
		ReleaseDate: dataset.ReleaseDate,
		NextRelease: dataset.NextRelease,
		DatasetID:   dataset.ID,
	}

	return p
}

// CreateRangeSelectorPage maps items from API responses to form a dimension range
// selector page model
func CreateRangeSelectorPage(name string, selectedValues, allValues data.DimensionValues, filter data.Filter, dataset data.Dataset) rangeSelector.Page {
	var p rangeSelector.Page

	p.SearchDisabled = true

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dataset.Title,
		URI:   fmt.Sprintf("/datasets/%s/editions/%s/versions/%s", filter.Dataset, filter.Edition, filter.Version),
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
	p.Data.SaveAndReturn = rangeSelector.Link{
		URL: fmt.Sprintf("/filters/%s/dimensions", filter.FilterID),
	}
	p.Data.Cancel = rangeSelector.Link{
		URL: fmt.Sprintf("/filters/%s/dimensions", filter.FilterID),
	}
	var selectedDates []string
	for _, val := range selectedValues.Items {
		selectedDates = append(selectedDates, val.Name)
	}

	selectedDats, _ := dates.ConvertToReadable(selectedDates)

	for i, val := range selectedDats {
		p.Data.FiltersAdded = append(p.Data.FiltersAdded, rangeSelector.Filter{
			RemoveURL: fmt.Sprintf("/filters/%s/dimensions/%s/remove/%s", filter.FilterID, name, selectedValues.Items[i].ID),
			Label:     dates.ConvertToMonthYear(val),
		})
	}

	p.Data.FiltersAmount = len(selectedValues.Items)

	for _, val := range allValues.Items {
		p.Data.RangeData.Values = append(p.Data.RangeData.Values, val.Label)
	}

	p.Data.RangeData.URL = fmt.Sprintf("/filters/%s/dimensions/%s/range", filter.FilterID, name)

	p.Data.RemoveAll.URL = fmt.Sprintf("/filters/%s/dimensions/%s/remove-all", filter.FilterID, name)

	p.Data.RangeData.StartLabel = "Start"
	p.Data.RangeData.EndLabel = "End"
	p.Data.RangeData.URL = fmt.Sprintf("/filters/%s/dimensions/%s/range", filter.FilterID, name)

	p.Metadata.Footer = model.Footer{
		Enabled:     true,
		Contact:     dataset.Contact.Name,
		ReleaseDate: dataset.ReleaseDate,
		NextRelease: dataset.NextRelease,
		DatasetID:   dataset.ID,
	}

	return p

}

// Random boolean generator
func randBool() bool {
	b := rand.Float32() < 0.5
	log.Debug("random bool", log.Data{"bool": b})
	return b
}

// CreatePreviewPage maps data items from API responses to create a preview page
func CreatePreviewPage(dimensions []data.Dimension, filter data.Filter, dataset data.Dataset, filterID string) previewPage.Page {
	var p previewPage.Page

	p.SearchDisabled = true

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: dataset.Title,
		URI:   fmt.Sprintf("/datasets/%s/editions/%s/versions/%s", filter.Dataset, filter.Edition, filter.Version),
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
		Contact:     dataset.Contact.Name,
		ReleaseDate: dataset.ReleaseDate,
		NextRelease: dataset.NextRelease,
		DatasetID:   dataset.ID,
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

	p.IsContentLoaded = randBool()

	return p
}

// CreateHierarchyPage ...
func CreateHierarchyPage(h data.Hierarchy, d data.Dataset, f data.Filter, met data.Metadata, curPath, dimensionTitle string) hierarchy.Page {
	var p hierarchy.Page

	p.SearchDisabled = true

	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: d.Title,
		URI:   fmt.Sprintf("/datasets/%s/editions/%s/versions/%s", f.Dataset, f.Edition, f.Version),
	})
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: "Filter this dataset",
		URI:   fmt.Sprintf("/filters/%s/dimensions", f.FilterID),
	})
	for _, par := range h.Parents {
		p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
			Title: par.Label,
			URI:   fmt.Sprintf("/filters/%s%s", f.FilterID, par.URI),
		})
	}
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: h.Label,
	})

	p.FilterID = f.FilterID
	p.Data.Title = h.Label

	if len(h.Parents) > 0 {
		p.Data.Parent = h.Parents[len(h.Parents)-1].Label
		p.Data.GoBack = hierarchy.Link{
			URL: fmt.Sprintf("/filters/%s%s", f.FilterID, h.Parents[len(h.Parents)-1].URI),
		}
	}

	p.Data.AddAllFilters.Amount = strconv.Itoa(len(h.Children))
	p.Data.AddAllFilters.URL = curPath + "/add-all"
	p.Data.RemoveAll.URL = curPath + "/remove-all"

	for _, dim := range f.Dimensions {
		if dim.Name == dimensionTitle {
			for _, val := range dim.Values {
				p.Data.FiltersAdded = append(p.Data.FiltersAdded, hierarchy.Filter{
					Label:     val,
					RemoveURL: fmt.Sprintf("%s/remove/%s", curPath, val),
				})
			}
		}
	}

	for _, child := range h.Children {
		var selected bool
		for _, dim := range f.Dimensions {
			if dim.Name == dimensionTitle {
				for _, val := range dim.Values {
					if val == child.Label {
						selected = true
					}
				}
			}
		}
		p.Data.FilterList = append(p.Data.FilterList, hierarchy.List{
			Label:    child.Label,
			SubNum:   strconv.Itoa(child.NumberofChildren),
			SubURL:   fmt.Sprintf("/filters/%s%s", f.FilterID, child.URI),
			Selected: selected,
		})

	}

	p.Data.Metadata = hierarchy.Metadata(met)
	p.Data.SaveAndReturn.URL = fmt.Sprintf("/filters/%s/dimensions", f.FilterID)

	p.Metadata.Footer = model.Footer{
		Enabled:     true,
		Contact:     d.Contact.Name,
		ReleaseDate: d.ReleaseDate,
		NextRelease: d.NextRelease,
		DatasetID:   d.ID,
	}

	return p
}
