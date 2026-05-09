package controllers

import (
	"net/http"

	"com.jetapcglobal.b2b.com/services"
)

type DestinationController interface {
	GetAllPages(w http.ResponseWriter, r *http.Request)
}

type destinationController struct {
	svc services.DestinationService
}

func NewDestinationController(svc services.DestinationService) DestinationController {
	return &destinationController{svc: svc}
}

func (c *destinationController) GetAllPages(w http.ResponseWriter, r *http.Request) {
	pages, err := c.svc.GetAllPages()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch pages")
		return
	}
	writeJSON(w, http.StatusOK, Response{Data: pages, Message: "success"})
}
