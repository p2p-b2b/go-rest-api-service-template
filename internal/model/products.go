package model

import (
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/p2p-b2b/qfv"
)

const (
	ProductNameMinLength                = 2
	ProductNameMaxLength                = 100
	ProductDescriptionMinLength         = 2
	ProductDescriptionMaxLength         = 1000
	ProductCurrencyLength               = 3
	ProductMaxNumberOfPaymentProcessors = 5

	ProductsProductCreatedSuccessfully = "Product created successfully"
	ProductsProductUpdatedSuccessfully = "Product updated successfully"
	ProductsProductDeletedSuccessfully = "Product deleted successfully"
)

var (
	// ProductsFilterFields is a list of valid fields for filtering models.
	ProductsFilterFields = []string{"id", "name", "price", "currency", "created_at", "updated_at"}

	// ProductsSortFields is a list of valid fields for sorting models.
	ProductsSortFields = []string{"id", "name", "price", "currency", "created_at", "updated_at"}

	// ProductsPartialFields is a list of valid fields for partial responses.
	ProductsPartialFields = []string{"id", "name", "description", "price", "currency", "created_at", "updated_at"}
)

// Product represents a product entity used to model the data stored in the database.
//
// @Description Product represents a product.
type Product struct {
	ID          uuid.UUID `json:"id,omitempty,omitzero" example:"01980434-b7ff-7abe-a45d-7311bc7011f5" format:"uuid"`
	Projects    *Project  `json:"project,omitempty"`
	Name        string    `json:"name,omitempty" example:"Product Name" format:"string"`
	Description string    `json:"description,omitempty" example:"This is a product" format:"string"`
	CreatedAt   time.Time `json:"created_at,omitzero" example:"2021-01-01T00:00:00Z" format:"date-time"`
	UpdatedAt   time.Time `json:"updated_at,omitzero" example:"2021-01-01T00:00:00Z" format:"date-time"`
	SerialID    int64     `json:"-"`
}

type InsertProductInput struct {
	ID          uuid.UUID
	ProjectID   uuid.UUID
	Name        string
	Description string
}

func (ref *InsertProductInput) Validate() error {
	var validationErrors ValidationErrors

	// Validate ID
	if err := ValidateUUID(ref.ID, 7, "id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Validate ProjectID
	if err := ValidateUUID(ref.ProjectID, 7, "project_id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Validate Name
	normalizedName, err := ValidateString(ref.Name, StringValidationOptions{
		MinLength:      ProductNameMinLength,
		MaxLength:      ProductNameMaxLength,
		TrimWhitespace: true,
		AllowEmpty:     false,
		NoControlChars: true,
		NoHTMLTags:     true,
		NoScriptTags:   true,

		NoNullBytes:      true,
		NormalizeUnicode: true,
		FieldName:        "name",
	})
	if err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	} else {
		ref.Name = normalizedName
	}

	// Validate Description
	normalizedDescription, err := ValidateString(ref.Description, StringValidationOptions{
		MinLength:      ProductDescriptionMinLength,
		MaxLength:      ProductDescriptionMaxLength,
		TrimWhitespace: true,
		AllowEmpty:     false,
		NoControlChars: true,
		NoHTMLTags:     true,
		NoScriptTags:   true,

		NoNullBytes:      true,
		NormalizeUnicode: true,
		FieldName:        "description",
	})
	if err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	} else {
		ref.Description = normalizedDescription
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

type CreateProductInput = InsertProductInput

type UpdateProductInput struct {
	ID          uuid.UUID
	ProjectID   uuid.UUID
	Name        *string
	Description *string
}

func (ref *UpdateProductInput) Validate() error {
	var validationErrors ValidationErrors

	// Validate ID
	if err := ValidateUUID(ref.ID, 7, "id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Validate ProjectID
	if err := ValidateUUID(ref.ProjectID, 7, "project_id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	if ref.Name != nil {
		normalizedName, err := ValidateString(*ref.Name, StringValidationOptions{
			MinLength:      ProductNameMinLength,
			MaxLength:      ProductNameMaxLength,
			TrimWhitespace: true,
			AllowEmpty:     false,
			NoControlChars: true,
			NoHTMLTags:     true,
			NoScriptTags:   true,

			NoNullBytes:      true,
			NormalizeUnicode: true,
			FieldName:        "name",
		})
		if err != nil {
			if ve, ok := err.(*ValidationError); ok {
				validationErrors.Errors = append(validationErrors.Errors, *ve)
			}
		} else {
			*ref.Name = normalizedName
		}
	}

	if ref.Description != nil {
		normalizedDescription, err := ValidateString(*ref.Description, StringValidationOptions{
			MinLength:      ProductDescriptionMinLength,
			MaxLength:      ProductDescriptionMaxLength,
			TrimWhitespace: true,
			AllowEmpty:     false,
			NoControlChars: true,
			NoHTMLTags:     true,
			NoScriptTags:   true,

			NoNullBytes:      true,
			NormalizeUnicode: true,
			FieldName:        "description",
		})
		if err != nil {
			if ve, ok := err.(*ValidationError); ok {
				validationErrors.Errors = append(validationErrors.Errors, *ve)
			}
		} else {
			*ref.Description = normalizedDescription
		}
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

type DeleteProductInput struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
}

func (ref *DeleteProductInput) Validate() error {
	var validationErrors ValidationErrors

	// Validate ID
	if err := ValidateUUID(ref.ID, 7, "id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Validate ProjectID
	if err := ValidateUUID(ref.ProjectID, 7, "project_id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

type SelectProductsInput struct {
	Sort      string
	Filter    string
	Fields    string
	Paginator Paginator
}

func (ref *SelectProductsInput) Validate() error {
	var validationErrors ValidationErrors

	// Validate paginator
	if err := ref.Paginator.Validate(); err != nil {
		if ve, ok := err.(*ValidationErrors); ok {
			validationErrors.Errors = append(validationErrors.Errors, ve.Errors...)
		} else {
			validationErrors.AddError("paginator", err.Error(), "INVALID_PAGINATION")
		}
	}

	// Validate sort expression
	if err := ValidateSortExpression(ref.Sort, "sort"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Additional sort validation using existing parser
	if ref.Sort != "" {
		if _, err := qfv.NewSortParser(ProductsSortFields).Parse(ref.Sort); err != nil {
			validationErrors.AddError("sort", err.Error(), "INVALID_SORT")
		}
	}

	// Validate filter expression
	if err := ValidateFilterExpression(ref.Filter, "filter"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Additional filter validation using existing parser
	if ref.Filter != "" {
		if _, err := qfv.NewFilterParser(ProductsFilterFields).Parse(ref.Filter); err != nil {
			validationErrors.AddError("filter", err.Error(), "INVALID_FILTER")
		}
	}

	// Validate fields expression
	if err := ValidateFieldsExpression(ref.Fields, "fields"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Additional fields validation using existing parser
	if ref.Fields != "" {
		if _, err := qfv.NewFieldsParser(ProductsPartialFields).Parse(ref.Fields); err != nil {
			validationErrors.AddError("fields", err.Error(), "INVALID_FIELDS")
		}
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

type ListProductsInput = SelectProductsInput

type SelectProductsOutput struct {
	Items     []Product `json:"items"`
	Paginator Paginator `json:"paginator"`
}

type ListProductsOutput = SelectProductsOutput

// CreateProductRequest represents the input for the CreateProduct method.
//
// @Description CreateProductRequest represents the input for the CreateProduct method.
type CreateProductRequest struct {
	ID          uuid.UUID `json:"id" example:"01980434-b7ff-7ac1-b7b0-13de306cc1cb" format:"uuid"`
	Name        string    `json:"name" example:"New product name" format:"string"`
	Description string    `json:"description" example:"This is a product" format:"string"`
}

// Validate validates the CreateProductRequest.
func (req *CreateProductRequest) Validate() error {
	var validationErrors ValidationErrors

	// Validate ID
	if err := ValidateUUID(req.ID, 7, "id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Validate Name
	normalizedName, err := ValidateString(req.Name, StringValidationOptions{
		MinLength:      ProductNameMinLength,
		MaxLength:      ProductNameMaxLength,
		TrimWhitespace: true,
		AllowEmpty:     false,
		NoControlChars: true,
		NoHTMLTags:     true,
		NoScriptTags:   true,

		NoNullBytes:      true,
		NormalizeUnicode: true,
		FieldName:        "name",
	})
	if err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	} else {
		req.Name = normalizedName
	}

	// Validate Description
	normalizedDescription, err := ValidateString(req.Description, StringValidationOptions{
		MinLength:      ProductDescriptionMinLength,
		MaxLength:      ProductDescriptionMaxLength,
		TrimWhitespace: true,
		AllowEmpty:     false,
		NoControlChars: true,
		NoHTMLTags:     true,
		NoScriptTags:   true,

		NoNullBytes:      true,
		NormalizeUnicode: true,
		FieldName:        "description",
	})
	if err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	} else {
		req.Description = normalizedDescription
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

// UpdateProductRequest represents the input for the UpdateProduct method.
//
// @Description UpdateProductRequest represents the input for the UpdateProduct method.
type UpdateProductRequest struct {
	Name        *string `json:"name" example:"Modified product name" format:"string"`
	Description *string `json:"description" example:"This is a product" format:"string"`
}

func (req *UpdateProductRequest) Validate() error {
	var validationErrors ValidationErrors

	// Check if any field is provided for update
	if reflect.DeepEqual(req, &UpdateProductRequest{}) {
		validationErrors.Errors = append(validationErrors.Errors, ValidationError{
			Field:   "request",
			Message: "at least one field must be provided for update",
		})
		return &validationErrors
	}

	// Validate Name (if provided)
	if req.Name != nil {
		normalizedName, err := ValidateString(*req.Name, StringValidationOptions{
			MinLength:      ProductNameMinLength,
			MaxLength:      ProductNameMaxLength,
			TrimWhitespace: true,
			AllowEmpty:     false,
			NoControlChars: true,
			NoHTMLTags:     true,
			NoScriptTags:   true,

			NoNullBytes:      true,
			NormalizeUnicode: true,
			FieldName:        "name",
		})
		if err != nil {
			if ve, ok := err.(*ValidationError); ok {
				validationErrors.Errors = append(validationErrors.Errors, *ve)
			}
		} else {
			*req.Name = normalizedName
		}
	}

	// Validate Description (if provided)
	if req.Description != nil {
		normalizedDescription, err := ValidateString(*req.Description, StringValidationOptions{
			MinLength:      ProductDescriptionMinLength,
			MaxLength:      ProductDescriptionMaxLength,
			TrimWhitespace: true,
			AllowEmpty:     false,
			NoControlChars: true,
			NoHTMLTags:     true,
			NoScriptTags:   true,

			NoNullBytes:      true,
			NormalizeUnicode: true,
			FieldName:        "description",
		})
		if err != nil {
			if ve, ok := err.(*ValidationError); ok {
				validationErrors.Errors = append(validationErrors.Errors, *ve)
			}
		} else {
			*req.Description = normalizedDescription
		}
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

// ListProductsResponse represents a list of users.
//
// @Description ListProductResponse represents a list of users.
type ListProductsResponse = SelectProductsOutput

// ProductPaymentProcessorRequest represents the input for linking a product to a payment processor.
//
// @Description ProductPaymentProcessorRequest represents the input for linking a product to a payment processor.
type ProductPaymentProcessorRequest struct {
	PaymentProcessorID        uuid.UUID `json:"payment_processor_id"`
	PaymentProcessorProductID string    `json:"payment_processor_product_id"`
}

// Validate validates the ProductPaymentProcessorRequest.
func (req *ProductPaymentProcessorRequest) Validate() error {
	var validationErrors ValidationErrors

	// Validate PaymentProcessorID
	if err := ValidateUUID(req.PaymentProcessorID, 7, "payment_processor_id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Validate PaymentProcessorProductID
	normalizedProductID, err := ValidateString(req.PaymentProcessorProductID, StringValidationOptions{
		MinLength:      1,
		MaxLength:      255, // reasonable limit for product IDs
		TrimWhitespace: true,
		AllowEmpty:     false,
		NoControlChars: true,
		NoHTMLTags:     true,
		NoScriptTags:   true,

		NoNullBytes:      true,
		NormalizeUnicode: true,
		FieldName:        "payment_processor_product_id",
	})
	if err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	} else {
		req.PaymentProcessorProductID = normalizedProductID
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

type LinkProductToPaymentProcessorInput struct {
	ProductID                 uuid.UUID
	PaymentProcessorID        uuid.UUID
	PaymentProcessorProductID string
}

func (ref *LinkProductToPaymentProcessorInput) Validate() error {
	var validationErrors ValidationErrors

	// Validate ProductID
	if err := ValidateUUID(ref.ProductID, 7, "product_id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Validate PaymentProcessorID
	if err := ValidateUUID(ref.PaymentProcessorID, 7, "payment_processor_id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Validate PaymentProcessorProductID
	normalizedProductID, err := ValidateString(ref.PaymentProcessorProductID, StringValidationOptions{
		MinLength:      1,
		MaxLength:      255, // reasonable limit for product IDs
		TrimWhitespace: true,
		AllowEmpty:     false,
		NoControlChars: true,
		NoHTMLTags:     true,
		NoScriptTags:   true,

		NoNullBytes:      true,
		NormalizeUnicode: true,
		FieldName:        "payment_processor_product_id",
	})
	if err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	} else {
		ref.PaymentProcessorProductID = normalizedProductID
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

type UnlinkProductFromPaymentProcessorInput = LinkProductToPaymentProcessorInput

// LinkProductToPaymentProcessorRequest represents the input for linking a product to a payment processor.
//
// @Description LinkProductToPaymentProcessorRequest represents the input for linking a product to a payment processor.
type LinkProductToPaymentProcessorRequest struct {
	PaymentProcessorID        uuid.UUID `json:"payment_processor_id"`
	PaymentProcessorProductID string    `json:"payment_processor_product_id"`
}

func (req *LinkProductToPaymentProcessorRequest) Validate() error {
	var validationErrors ValidationErrors

	// Validate PaymentProcessorID
	if err := ValidateUUID(req.PaymentProcessorID, 7, "payment_processor_id"); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	}

	// Validate PaymentProcessorProductID
	normalizedProductID, err := ValidateString(req.PaymentProcessorProductID, StringValidationOptions{
		MinLength:      1,
		MaxLength:      255, // reasonable limit for product IDs
		TrimWhitespace: true,
		AllowEmpty:     false,
		NoControlChars: true,
		NoHTMLTags:     true,
		NoScriptTags:   true,

		NoNullBytes:      true,
		NormalizeUnicode: true,
		FieldName:        "payment_processor_product_id",
	})
	if err != nil {
		if ve, ok := err.(*ValidationError); ok {
			validationErrors.Errors = append(validationErrors.Errors, *ve)
		}
	} else {
		req.PaymentProcessorProductID = normalizedProductID
	}

	if validationErrors.HasErrors() {
		return &validationErrors
	}

	return nil
}

// UnlinkProductFromPaymentProcessorRequest represents the input for unlinking a product from a payment processor.
//
// @Description UnlinkProductFromPaymentProcessorRequest represents the input for unlinking a product from a payment processor.
type UnlinkProductFromPaymentProcessorRequest = LinkProductToPaymentProcessorRequest
