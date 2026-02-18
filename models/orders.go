package models

import (
	"github.com/jinzhu/gorm"
	"github.com/teris-io/shortid"
)

var (
	OrderStatuses = []string{
		"Received",
		"Preparing",
		"Baking",
		"Quality Check",
		"Out for Delivery",
		"Delivered",
	}
	PizzaTypes    = []string{
		"Margherita",
		"Pepperoni", "Veggie",
		"BBQ Chicken",
		"Hawaiian",
	}
	PizzaSizes    = []string{
		"Small",
		"Medium",
		"Large",
		"Extra Large",
	}
)

type OrderModel struct {
	DB *gorm.DB
}

type Order struct {
	gorm.Model
	ID           string      `gorm:"primary_key;size:13" json:"id"`
	CustomerName string      `gorm:"size:100;not null" json:"customer_name"`
	PizzaType    string      `gorm:"size:50" json:"pizza_type"`
	PizzaSize    string      `gorm:"size:20" json:"pizza_size"`
	Status       string      `gorm:"size:30" json:"status"`
	Phone        string      `gorm:"size:20" json:"phone"`
	Address      string      `gorm:"size:200" json:"address"`
	Items        []OrderItem `gorm:"foreignkey:OrderID" json:"items"`
}

type OrderItem struct {
	gorm.Model
	ID           string    `gorm:"primary_key;size:13" json:"id"`
	OrderID      string    `gorm:"index;not null" json:"order_id"`
	Size         string `gorm:"not null;size:20" json:"size"`
	ItemName     string `gorm:"not null;size:100" json:"item_name"`
	Instructions string `gorm:"size:200" json:"instructions"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = shortid.MustGenerate()
	}
	return nil
}

func (oi *OrderItem) BeforeCreate(tx *gorm.DB) error {
	if oi.ID == "" {
		oi.ID = shortid.MustGenerate()
	}
	return nil
}

func (o *OrderModel) CreateOrder(order *Order) error {
	return o.DB.Create(order).Error
}

func (o *OrderModel) GetOrder(id string) (*Order, error) {
	var order Order
	err := o.DB.Preload("Items").First(&order, "id = ?", id).Error
	return &order, err
}