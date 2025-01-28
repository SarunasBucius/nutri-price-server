package recipe

import (
	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/umath"
)

func calculateMealNutritionalValue(ingredients []model.Ingredient, productsNV []model.ProductNutritionalValue) model.CalculatedMealNutritionalValue {
	calculatedProductsNV := make(map[int][]model.CalculatedProductNutritionalValue, len(ingredients))
	var totalNV model.NutritionalValue
	for _, ingredient := range ingredients {
		calculatedProductNV := calculateIngredientNutritionalValue(ingredient, productsNV)
		calculatedProductsNV[ingredient.RecipeID] = append(calculatedProductsNV[ingredient.RecipeID], calculatedProductNV)
		totalNV = addNutritionalValues(totalNV, calculatedProductNV.NutritionalValue)
	}

	calculatedNVByRecipe := make([]model.CalculatedRecipeNutritionalValue, 0, len(calculatedProductsNV))
	for recipeID := range calculatedProductsNV {
		nvs := make([]model.NutritionalValue, 0, len(calculatedProductsNV[recipeID]))
		for _, calculatedProduct := range calculatedProductsNV[recipeID] {
			nvs = append(nvs, calculatedProduct.NutritionalValue)
		}

		calculatedNVByRecipe = append(calculatedNVByRecipe, model.CalculatedRecipeNutritionalValue{
			RecipeID:           recipeID,
			CalculatedProducts: calculatedProductsNV[recipeID],
			NutritionalValue:   addNutritionalValues(nvs...),
		})
	}
	return model.CalculatedMealNutritionalValue{
		NutritionalValue:  totalNV,
		CalculatedRecipes: calculatedNVByRecipe,
	}
}

func calculateIngredientNutritionalValue(ingredient model.Ingredient, productsNutritionalValue []model.ProductNutritionalValue) model.CalculatedProductNutritionalValue {
	for _, productNV := range productsNutritionalValue {
		if productNV.Product != ingredient.Product {
			continue
		}
		if productNV.Unit == ingredient.Unit {
			isPiece := productNV.Unit == model.Pieces
			return model.CalculatedProductNutritionalValue{
				Product:          ingredient.Product,
				NutritionalValue: calculateNutritionalValue(ingredient.Amount, productNV.NutritionalValue, isPiece),
			}
		}
	}

	return model.CalculatedProductNutritionalValue{
		Product: ingredient.Product,
		Message: "could not find nutritional value for the product",
	}
}

func calculateNutritionalValue(ingredientAmount float64, productNV model.NutritionalValue, isPiece bool) model.NutritionalValue {
	multiplier := float64(100)
	if isPiece {
		multiplier = 1
	}
	return model.NutritionalValue{
		EnergyValueKCAL:    umath.RoundFloat(productNV.EnergyValueKCAL/multiplier*ingredientAmount, 0),
		Fat:                umath.RoundFloat(productNV.Fat/multiplier*ingredientAmount, 3),
		SaturatedFat:       umath.RoundFloat(productNV.SaturatedFat/multiplier*ingredientAmount, 3),
		Carbohydrate:       umath.RoundFloat(productNV.Carbohydrate/multiplier*ingredientAmount, 3),
		CarbohydrateSugars: umath.RoundFloat(productNV.CarbohydrateSugars/multiplier*ingredientAmount, 3),
		Fibre:              umath.RoundFloat(productNV.Fibre/multiplier*ingredientAmount, 3),
		SolubleFibre:       umath.RoundFloat(productNV.SolubleFibre/multiplier*ingredientAmount, 3),
		InsolubleFibre:     umath.RoundFloat(productNV.InsolubleFibre/multiplier*ingredientAmount, 3),
		Protein:            umath.RoundFloat(productNV.Protein/multiplier*ingredientAmount, 3),
		Salt:               umath.RoundFloat(productNV.Salt/multiplier*ingredientAmount, 3),
	}
}

func addNutritionalValues(nutritionalValues ...model.NutritionalValue) model.NutritionalValue {
	var total model.NutritionalValue
	for _, nv := range nutritionalValues {
		total = model.NutritionalValue{
			EnergyValueKCAL:    total.EnergyValueKCAL + nv.EnergyValueKCAL,
			Fat:                total.Fat + nv.Fat,
			SaturatedFat:       total.SaturatedFat + nv.SaturatedFat,
			Carbohydrate:       total.Carbohydrate + nv.Carbohydrate,
			CarbohydrateSugars: total.CarbohydrateSugars + nv.CarbohydrateSugars,
			Fibre:              total.Fibre + nv.Fibre,
			SolubleFibre:       total.SolubleFibre + nv.SolubleFibre,
			InsolubleFibre:     total.InsolubleFibre + nv.InsolubleFibre,
			Protein:            total.Protein + nv.Protein,
			Salt:               total.Salt + nv.Salt,
		}
	}
	return total
}
