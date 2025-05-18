package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/uerror"
	"github.com/go-chi/chi/v5"
	"google.golang.org/genai"
)

type ReceiptAPI struct {
	Service IReceiptService
}

func NewReceiptAPI(receiptService IReceiptService) *ReceiptAPI {
	return &ReceiptAPI{Service: receiptService}
}

type IReceiptService interface {
	ProcessReceipt(ctx context.Context, receipt string) (model.ParseReceiptFromTextResponse, error)
	ProcessReceiptFromDB(ctx context.Context, receiptDate string) (model.ParseReceiptFromTextResponse, error)
	GetUnconfirmedReceiptSummaries(ctx context.Context) ([]model.UnconfirmedReceiptSummary, error)
	GetUnconfirmedReceipt(ctx context.Context, retailer, date string) ([]model.PurchasedProductNew, error)
	GetLastReceiptDates(ctx context.Context) ([]model.LastReceiptDate, error)
	GetProductsWithMissingInfo(ctx context.Context, dateFrom string) ([]model.ProductAndVarietyName, error)
}

func (rc *ReceiptAPI) ParseReceiptFromText(w http.ResponseWriter, r *http.Request) {
	receipt, err := getReceiptFromBody(r)
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	res, err := parseWithAI(r.Context(), receipt)
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	// processedReceipt, err := rc.Service.ProcessReceipt(r.Context(), receipt)
	// if err != nil {
	// 	errorResponse(r.Context(), w, err)
	// 	return
	// }

	successResponse(r.Context(), w, res)
}

// Receipt represents the structured data from a grocery receipt
type Receipt struct {
	Retailer   string    `json:"retailer"`
	Date       string    `json:"date"`
	TotalPrice float64   `json:"totalPrice"`
	Products   []Product `json:"products"`
}

// Product represents an individual item on the receipt
type Product struct {
	Name     string  `json:"name"`
	Variety  string  `json:"variety"`
	Price    float64 `json:"price"`
	Discount float64 `json:"discount,omitempty"`
	Quantity float64 `json:"quantity,omitempty"`
	Unit     string  `json:"unit,omitempty"`
}

// parseWithAI sends input text to Gemini API and returns the response text
func parseWithAI(ctx context.Context, input string) (*Receipt, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	prompt := `You are a structured data extractor for grocery receipts.
Given the receipt text, extract the following fields:
    "retailer" - the name of the store, lowercase. Example: "lidl", "maxima", "norfa".
    "date" - purchase date in the format YYYY-MM-DD.
	"totalPrice" - the total price of the receipt, as a float.
    "products" - an array of objects, each with the following keys:
        "name" - the general product type (e.g. "Duona", "Bananai").
        "variety" - descriptive attributes such as brand, flavor, fat %, size, color, or packaging.
        "price" - the price for that product, as a float.
        "quantity" - numeric quantity if present (e.g. 10, 1300), skip if not present. Convert to grams or milliliters if necessary.
        "unit" - measurement unit (e.g. "g", "kg", "L", "vnt"), skip if not present.
		"discount" - The monetary value of a discount specifically applied to this product, do not apply this to "price" field. Look for a line appearing after this product's entry that indicates a price reduction for this item. Keywords include "nuolaida", "Akcija", or "KUPONU". Extract the absolute numeric monetary value of the discount. If a line shows both a percentage and an explicit monetary discount amount for this item, prioritize the monetary amount. This value should be a float; use 0.0 if no specific discount is found for this product.
Special Parsing Rules:
    If a product line shows a price calculation like 1,3x0,98, interpret this as 1.3 kg * 0.98 €/kg = total price. Use the quantity.
	Do not make any calculations or assumptions about the "price" field of product.
    Use the original receipt language for product names and variety fields.
Input receipt: %s`
	response, err := client.Models.GenerateContent(ctx, "gemini-2.0-flash", genai.Text(fmt.Sprintf(prompt, input)),
		&genai.GenerateContentConfig{
			Temperature:      genai.Ptr[float32](0.0),
			TopP:             genai.Ptr[float32](0.0),
			TopK:             genai.Ptr[float32](1),
			ResponseMIMEType: "application/json",
			ResponseSchema:   receiptSchema,
			MaxOutputTokens:  int32(len(input)),
		})
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %v", err)
	}

	var unparsedReceipt string
	for _, candidate := range response.Candidates {
		for _, part := range candidate.Content.Parts {
			unparsedReceipt += fmt.Sprintf("%v", part.Text)
		}
	}
	fmt.Println("First call result:", unparsedReceipt)

	parsedReceipt, err := parseReceiptString(unparsedReceipt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse receipt JSON: %v", err)
	}
	for i := range parsedReceipt.Products {
		absDiscount := math.Abs(parsedReceipt.Products[i].Discount)
		if absDiscount > 0 {
			parsedReceipt.Products[i].Price -= absDiscount
		}
	}
	return parsedReceipt, nil
}

var receiptSchema = &genai.Schema{
	Title:       "Receipt Data",
	Description: "Schema for representing a shopping receipt.",
	Type:        genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"retailer": {
			Type:        genai.TypeString,
			Description: "Name of the retailer.",
			Example:     "",
		},
		"date": {
			Type:        genai.TypeString,
			Description: "Date of the purchase in YYYY-MM-DD format.",
			Example:     "2021-05-14",
		},
		"totalPrice": {
			Type:        genai.TypeNumber,
			Format:      "double",
			Description: "Total price of the receipt.",
			Example:     15.99,
		},
		"products": {
			Type:        genai.TypeArray,
			Description: "List of purchased products.",
			Items: &genai.Schema{
				Type:        genai.TypeObject,
				Description: "Represents a single product item.",
				Properties: map[string]*genai.Schema{
					"name": {
						Type:        genai.TypeString,
						Description: "Name of the product.",
					},
					"variety": {
						Type:        genai.TypeString,
						Description: "Specific variety or description of the product. Can be empty.",
					},
					"price": {
						Type:        genai.TypeNumber,
						Format:      "double",
						Description: "Price of the product per unit.",
						Example:     0.89,
					},
					"discount": {
						Type:        genai.TypeNumber,
						Format:      "double",
						Description: "Discount applied to the product.",
						Example:     0.10,
					},
					"quantity": {
						Type:        genai.TypeInteger,
						Format:      "int64",
						Description: "Quantity of the product. Only present if more than one or if measured by weight/volume.",
						Example:     2,
					},
					"unit": {
						Type:        genai.TypeString,
						Description: "Unit of measurement for the quantity (e.g., g, kg).",
						Enum:        []string{"vnt", "g", "ml"},
						Example:     "vnt",
					},
				},
				PropertyOrdering: []string{"name", "variety", "price", "quantity", "unit"},
			},
		},
	},
	Required: []string{"retailer", "date", "products"},
}

// parseReceiptString parses the JSON string into a Receipt struct
func parseReceiptString(jsonStr string) (*Receipt, error) {
	var receipt Receipt
	err := json.Unmarshal([]byte(jsonStr), &receipt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse receipt JSON: %v", err)
	}

	return &receipt, nil
}

func (rc *ReceiptAPI) ParseReceiptInDB(w http.ResponseWriter, r *http.Request) {
	receiptDate := r.URL.Query().Get("receiptDate")

	processedReceipt, err := rc.Service.ProcessReceiptFromDB(r.Context(), receiptDate)
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, processedReceipt)
}

func getReceiptFromBody(r *http.Request) (string, error) {
	if r.Header.Get("Content-Type") == "text/plain" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return "", uerror.NewBadRequest("unable to read text request body", err)
		}
		return string(body), nil
	}

	var receipt struct {
		Receipt string `json:"receipt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&receipt); err != nil {
		return "", uerror.NewBadRequest("invalid json request body", err)
	}

	return receipt.Receipt, nil
}

func (rc *ReceiptAPI) GetUnconfirmedReceiptSummaries(w http.ResponseWriter, r *http.Request) {
	summaries, err := rc.Service.GetUnconfirmedReceiptSummaries(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, emptyIfNil(summaries))
}

func (rc *ReceiptAPI) GetUnconfirmedReceipt(w http.ResponseWriter, r *http.Request) {
	retailerAndDate := chi.URLParam(r, "retailerAndDate")
	retailerAndDateSplit := strings.Split(retailerAndDate, "_")
	if len(retailerAndDateSplit) != 2 {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid retailerAndDate query parameter", nil))
		return
	}

	products, err := rc.Service.GetUnconfirmedReceipt(r.Context(), retailerAndDateSplit[0], retailerAndDateSplit[1])
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, products)
}

func (rc *ReceiptAPI) GetLastReceiptDates(w http.ResponseWriter, r *http.Request) {
	dates, err := rc.Service.GetLastReceiptDates(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, emptyIfNil(dates))
}

func (rc *ReceiptAPI) GetProductsWithMissingInfo(w http.ResponseWriter, r *http.Request) {
	dateFrom := r.URL.Query().Get("dateFrom")
	if dateFrom == "" {
		errorResponse(r.Context(), w, uerror.NewBadRequest("missing dateFrom query parameter", nil))
		return
	}

	products, err := rc.Service.GetProductsWithMissingInfo(r.Context(), dateFrom)
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, emptyIfNil(products))
}
