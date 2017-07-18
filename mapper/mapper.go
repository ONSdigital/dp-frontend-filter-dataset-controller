package mapper

import (
	"fmt"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/data"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/filterOverview"
)

var dimensionTitleTranslator = map[string]string{
	"geography": "Geographic Areas",
	"year":      "Year",
	"age-range": "Age",
	"sex":       "Sex",
}

// CreateFilterOverview ...
func CreateFilterOverview(dimensions []data.Dimension, filter data.Filter, dataset data.Dataset, filterID string) filterOverview.Page {
	var p filterOverview.Page

	p.FilterID = filterID

	for _, d := range dimensions {
		var fod filterOverview.Dimension

		for _, ac := range d.Values {
			fod.AddedCategories = append(fod.AddedCategories, ac)
		}

		fod.Link.URL = fmt.Sprintf("/filters/%s/dimensions/%s", filterID, d.Name)
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
		NextRelease: dataset.ReleaseDate,
		DatasetID:   dataset.ID,
	}

	return p
}
