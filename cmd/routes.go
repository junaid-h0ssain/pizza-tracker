package main

import (
	"github.com/gin-gonic/gin"
)

func setupRoutes(router *gin.Engine, h *Handler) {
	// Customer routes
	router.GET("/", h.ServeNewOrderForm)
	router.POST("/new-order", h.HandleNewOrderPost)
	router.GET("/customer/:id", h.serveCustomer)

	// Static files
	router.Static("/static", "./tmp/static")
}