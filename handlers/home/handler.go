package home

import (
	"html/template"
	"learn-gin/pkg/router"
	"learn-gin/views"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

type IndexRequest struct{}

func (h *Handler) Index(ctx *router.Context, req IndexRequest) (template.HTML, error) {
	return views.Index(views.IndexData{})
}

type GetDatasetRequest struct {
	DatasetID string `json:"datasetId"`
}

func (h *Handler) GetDataset(ctx *router.Context, req GetDatasetRequest) (template.HTML, error) {
	return views.Index(views.IndexData{})
}
