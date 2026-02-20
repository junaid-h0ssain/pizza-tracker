package main

import (
	"slices"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"pizza-tracker/models"
)

func RegisterValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Register custom validation functions here
		err := v.RegisterValidation("valid_pizza_type", sliceValidate(models.PizzaTypes))
		if err != nil {
			return 
		}
		err = v.RegisterValidation("valid_pizza_size", sliceValidate(models.PizzaSizes))
		if err != nil {
			return
		}
	}
}	

func sliceValidate(allowedValues []string) validator.Func {
	return func(fl validator.FieldLevel) bool {
		return slices.Contains(allowedValues, fl.Field().String())
	}
}