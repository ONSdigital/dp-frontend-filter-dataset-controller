package mapper

import (
	"fmt"
	"strconv"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/data"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/filterOverview"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/hierarchy"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/previewPage"
)

var dimensionTitleTranslator = map[string]string{
	"geography":          "Geographic Areas",
	"year":               "Year",
	"age-range":          "Age",
	"sex":                "Sex",
	"month":              "Month",
	"goods-and-services": "Goods and Services",
}

// CreateFilterOverview maps data items from API responses to form a filter overview
// front end page model
func CreateFilterOverview(dimensions []data.Dimension, filter data.Filter, dataset data.Dataset, filterID string) filterOverview.Page {
	var p filterOverview.Page

	p.FilterID = filterID

	for _, d := range dimensions {
		var fod filterOverview.Dimension

		for _, ac := range d.Values {
			fod.AddedCategories = append(fod.AddedCategories, ac)
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
