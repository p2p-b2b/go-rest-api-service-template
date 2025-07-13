package model

import (
	"fmt"

	"github.com/google/uuid"
)

type InvalidProductNameError struct {
	MinLength int
	MaxLength int
}

func (e *InvalidProductNameError) Error() string {
	return fmt.Sprintf("invalid name: must be between %d and %d characters", e.MinLength, e.MaxLength)
}

type InvalidProductDescriptionError struct {
	MinLength int
	MaxLength int
}

func (e *InvalidProductDescriptionError) Error() string {
	return fmt.Sprintf("invalid description: must be between %d and %d characters", e.MinLength, e.MaxLength)
}

type InvalidProductPriceError struct{}

func (e *InvalidProductPriceError) Error() string {
	return "invalid price: must be greater than 0"
}

type InvalidProductCurrencyError struct{}

func (e *InvalidProductCurrencyError) Error() string {
	return "invalid currency: must be 3 characters"
}

type InvalidProductIDError struct {
	ID      uuid.UUID
	Message string
}

func (e *InvalidProductIDError) Error() string {
	if e.ID != uuid.Nil {
		return fmt.Sprintf("invalid product ID: %s", e.ID.String())
	}
	if e.Message != "" {
		return fmt.Sprintf("invalid product ID: %s", e.Message)
	}

	return "invalid product ID"
}

type ProductNotFoundError struct {
	ProductID string
}

func (e *ProductNotFoundError) Error() string {
	return fmt.Sprintf("product not found: %s", e.ProductID)
}

type ProductNameExistsError struct {
	Name string
}

func (e *ProductNameExistsError) Error() string {
	return fmt.Sprintf("product name already exists: %s", e.Name)
}

type ProductIDExistsError struct {
	ID string
}

func (e *ProductIDExistsError) Error() string {
	return fmt.Sprintf("product ID already exists: %s", e.ID)
}

type InvalidProductUpdateError struct {
	Message string
}

func (e *InvalidProductUpdateError) Error() string {
	return fmt.Sprintf("invalid product update: %s", e.Message)
}

type ProductNameAlreadyExistsError struct {
	Name string
}

func (e *ProductNameAlreadyExistsError) Error() string {
	if e.Name != "" {
		return fmt.Sprintf("product name already exists: %s", e.Name)
	}
	return "product name already exists"
}

type ProductIDAlreadyExistsError struct {
	ID string
}

func (e *ProductIDAlreadyExistsError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("product ID already exists: %s", e.ID)
	}
	return "product ID already exists"
}

type InvalidPaymentProcessorProductIDError struct {
	Message string
}

func (e *InvalidPaymentProcessorProductIDError) Error() string {
	return fmt.Sprintf("invalid payment processor product ID: %s", e.Message)
}
