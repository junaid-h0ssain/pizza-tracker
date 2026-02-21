package main

import (
	"pizza-tracker/models"
)

// Handler holds all dependencies for HTTP handlers
type Handler struct {
	orders *models.OrderModel
}

func NewHandler(dbModel *models.DBModel) *Handler {
	return &Handler{
		orders: &dbModel.Order,
	}
}