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
		if product.Name != ingredient.Product && product.Group != ingredient.Product {
			continue
		}

		if product.Quantity.Unit == ingredient.NormalizedQuantity.Unit {
			unroundedProductPrice := product.Price.Paid / product.Quantity.Amount * ingredient.NormalizedQuantity.Amount
			productPrice := umath.RoundFloat(unroundedProductPrice, 2)
			return model.CalculatedProductPrice{
				Product: ingredient.Product,
				Price:   productPrice,
			}
		}

		if product.Quantity.Unit == ingredient.RecipeQuantity.Unit {
			unroundedProductPrice := product.Price.Paid / product.Quantity.Amount * ingredient.RecipeQuantity.Amount
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
