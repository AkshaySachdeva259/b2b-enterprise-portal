package controllers

import (
	"net/http"

	"com.jetapcglobal.b2b.com/services"
)

type DestinationController interface {
	GetDestinations(w http.ResponseWriter, r *http.Request)
}

type destinationController struct {
	svc services.DestinationService
}

func NewDestinationController(svc services.DestinationService) DestinationController {
	return &destinationController{svc: svc}
}

func (c *destinationController) GetDestinations(w http.ResponseWriter, r *http.Request) {
	destinations, err := c.svc.ListDestinations()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch destinations")
		return
	}
	writeJSON(w, http.StatusOK, Response{Data: destinations, Message: "success"})
}
