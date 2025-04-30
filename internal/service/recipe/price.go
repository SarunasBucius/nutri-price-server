package recipe

import (
	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/umath"
)

func calculateMealPrice(ingredients []model.Ingredient, purchasedProducts []model.PurchasedProduct) model.CalculatedMealPrice {
	calculatedProducts := make(map[int][]model.CalculatedProductPrice, len(ingredients))
	var totalPrice float64
	for _, ingredient := range ingredients {
		calculatedProduct := calculateIngredientPrice(ingredient, purchasedProducts)
		calculatedProducts[ingredient.RecipeID] = append(calculatedProducts[ingredient.RecipeID], calculatedProduct)
		totalPrice += calculatedProduct.Price
	}

	calculatedNVByRecipe := make([]model.CalculatedRecipePrice, 0, len(calculatedProducts))
	for recipeID := range calculatedProducts {
		var totalRecipePrice float64
		for _, calculatedProduct := range calculatedProducts[recipeID] {
			totalRecipePrice += calculatedProduct.Price
		}

		calculatedNVByRecipe = append(calculatedNVByRecipe, model.CalculatedRecipePrice{
			RecipeID:           recipeID,
			CalculatedProducts: calculatedProducts[recipeID],
			Price:              totalRecipePrice,
		})
	}
	return model.CalculatedMealPrice{
		Price:             totalPrice,
		CalculatedRecipes: calculatedNVByRecipe,
	}
}

func calculateIngredientPrice(ingredient model.Ingredient, purchasedProducts []model.PurchasedProduct) model.CalculatedProductPrice {
	for _, product := range purchasedProducts {
		if product.Name != ingredient.Product {
			continue
		}

		if product.Quantity.Unit == ingredient.Unit {
			unroundedProductPrice := product.Price / product.Quantity.Amount * ingredient.Amount
			productPrice := umath.RoundFloat(unroundedProductPrice, 2)
			return model.CalculatedProductPrice{
				Product: ingredient.Product,
				Price:   productPrice,
			}
		}
	}
	return model.CalculatedProductPrice{
		Product: ingredient.Product,
		Message: "could not find price for the product",
	}
}
