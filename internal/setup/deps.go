package setup

import (
	"github.com/SarunasBucius/nutri-price-server/internal/api"
	"github.com/SarunasBucius/nutri-price-server/internal/repository"
	"github.com/SarunasBucius/nutri-price-server/internal/service/nutritionalvalue"
	"github.com/SarunasBucius/nutri-price-server/internal/service/product"
	"github.com/SarunasBucius/nutri-price-server/internal/service/receipt"
	"github.com/SarunasBucius/nutri-price-server/internal/service/recipe"
)

type handlers struct {
	product *api.ProductAPI
	receipt *api.ReceiptAPI
	nv      *api.NutritionalValueAPI
	recipes *api.RecipeAPI
}

func loadAPIHandlers(conf Config) handlers {
	receiptRepo := repository.NewReceiptRepo(conf.DBPool)
	productRepo := repository.NewProductRepo(conf.DBPool)
	nvRepo := repository.NewNutritionalValueRepo(conf.DBPool)
	recipesRepo := repository.NewRecipeRepo(conf.DBPool)

	receiptService := receipt.NewReceiptService(receiptRepo, productRepo)
	productService := product.NewProductService(productRepo, receiptRepo)
	nvService := nutritionalvalue.NewNutritionalValueService(nvRepo)
	recipeService := recipe.NewRecipeService(productRepo, nvRepo, recipesRepo)

	receiptAPI := api.NewReceiptAPI(receiptService)
	productAPI := api.NewProductAPI(productService)
	nvAPI := api.NewNutritionalValuesAPI(nvService)
	recipeAPI := api.NewRecipeAPI(recipeService)

	return handlers{
		receipt: receiptAPI,
		product: productAPI,
		nv:      nvAPI,
		recipes: recipeAPI,
	}
}
