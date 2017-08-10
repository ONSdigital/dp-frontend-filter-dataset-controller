package handlers

import (
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/renderer"
	"github.com/ONSdigital/dp-frontend-models/model"
)

// Filter represents the handlers for Filtering
type Filter struct {
	r   renderer.Renderer
	fc  FilterClient
	dc  DatasetClient
	clc CodelistClient
	val Validator
}

// NewFilter creates a new instance of Filter
func NewFilter(r renderer.Renderer, fc FilterClient, dc DatasetClient, clc CodelistClient, val Validator) *Filter {
	return &Filter{
		r:   r,
		fc:  fc,
		dc:  dc,
		clc: clc,
		val: val,
	}
}
func getStubbedMetadataFooter() model.Footer {
	return model.Footer{
		Enabled:     true,
		Contact:     "Matt Rout",
		ReleaseDate: "11 November 2016",
		NextRelease: "11 November 2017",
		DatasetID:   "MR",
	}
}
