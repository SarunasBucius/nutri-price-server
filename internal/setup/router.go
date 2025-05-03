package setup

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func LoadRouter(conf Config) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	h := loadAPIHandlers(conf)

	r.Post("/purchased-products/parse-from-receipt-text", h.receipt.ParseReceiptFromText)
	r.Post("/purchased-products/parse-from-receipt-in-db", h.receipt.ParseReceiptInDB)
	r.Get("/purchased-products/unconfirmed-receipts/summary", h.receipt.GetUnconfirmedReceiptSummaries)
	r.Get("/purchased-products/unconfirmed-receipts/{retailerAndDate}", h.receipt.GetUnconfirmedReceipt)
	r.Post("/purchased-products/confirm", h.product.ConfirmPurchasedProducts)
	r.Get("/purchased-products/last-receipt-dates", h.receipt.GetLastReceiptDates)
	r.Get("/purchased-products/with-missing-info", h.receipt.GetProductsWithMissingInfo)

	r.Post("/nutritional-values", h.nv.InsertNutritionalValues)
	r.Get("/nutritional-values", h.nv.GetNutritionalValues)
	r.Get("/nutritional-values/available-units", h.nv.GetNutritionalValuesUnits)
	r.Get("/nutritional-values/{nutritionalValueID}", h.nv.GetNutritionalValue)
	r.Put("/nutritional-values/{nutritionalValueID}", h.nv.UpdateNutritionalValue)
	r.Delete("/nutritional-values/{nutritionalValueID}", h.nv.DeleteNutritionalValues)

	r.Post("/recipes", h.recipes.InsertRecipe)
	r.Get("/recipes/summary", h.recipes.GetRecipeSummaries)
	r.Get("/recipes/names", h.recipes.GetRecipeNames)
	r.Get("/recipes/{recipeID}", h.recipes.GetRecipe)
	r.Put("/recipes/{recipeID}", h.recipes.UpdateRecipe)
	r.Get("/recipes/{recipeIDs}/meal-nutritional-value", h.recipes.GetMealNutritionalValue)
	r.Get("/recipes/{recipeIDs}/meal-price", h.recipes.GetMealPrice)
	r.Get("/recipes/meal-nutritional-value-by-date/{date}", h.recipes.GetMealNutritionalValueByDate)
	r.Get("/recipes/meal-price-by-date/{date}", h.recipes.GetMealPriceByDate)
	r.Delete("/recipes/{recipeID}", h.recipes.DeleteRecipe)
	r.Post("/recipes/clone", h.recipes.CloneRecipes)

	return r
}
