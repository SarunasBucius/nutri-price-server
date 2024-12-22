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
	r.Post("/purchased-products", h.product.InsertProducts)
	r.Get("/purchased-products/product-groups", h.product.GetProductGroups)
	r.Get("/purchased-products", h.product.GetProducts)
	r.Get("/purchased-products/{productID}", h.product.GetProduct)
	r.Put("/purchased-products/{productID}", h.product.UpdateProduct)
	r.Delete("/purchased-products/{productID}", h.product.DeleteProduct)

	r.Post("/nutritional-values", h.nv.InsertNutritionalValues)
	r.Get("/nutritional-values", h.nv.GetNutritionalValues)
	r.Get("/nutritional-values/{nutritionalValueID}", h.nv.GetNutritionalValue)
	r.Put("/nutritional-values/{nutritionalValueID}", h.nv.UpdateNutritionalValue)
	r.Delete("/nutritional-values/{nutritionalValueID}", h.nv.DeleteNutritionalValues)

	r.Post("/recipes", h.recipes.InsertRecipe)
	r.Get("/recipes/names", h.recipes.GetRecipeNames)
	r.Get("/recipes/{recipeID}", h.recipes.GetRecipe)
	r.Put("/recipes/{recipeID}", h.recipes.UpdateRecipe)
	r.Get("/recipes/{recipeIDs}/meal-nutritional-value", h.recipes.GetMealNutritionalValue)
	r.Get("/recipes/{recipeIDs}/meal-price", h.recipes.GetMealPrice)
	r.Delete("/recipes/{recipeID}", h.recipes.DeleteRecipe)

	return r
}
