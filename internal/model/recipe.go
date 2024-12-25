package model

import "time"

type RecipeNew struct {
	Name         string          `json:"name"`
	Ingredients  []IngredientNew `json:"ingredients"`
	Steps        []string        `json:"steps"`
	Notes        string          `json:"notes"`
	DishMadeDate *time.Time      `json:"dishMadeDate,omitempty"`
}

type IngredientNew struct {
	Product            string   `json:"product"`
	RecipeQuantity     Quantity `json:"recipeQuantity"`
	NormalizedQuantity Quantity `json:"normalizedQuantity"`
	CutStyle           string   `json:"cutStyle"`
}

type Recipe struct {
	ID           int          `json:"id"`
	Name         string       `json:"name"`
	Ingredients  []Ingredient `json:"ingredients"`
	Steps        []string     `json:"steps"`
	Notes        string       `json:"notes"`
	DishMadeDate *time.Time   `json:"dishMadeDate,omitempty"`
}

type RecipeUpdate struct {
	ID           int             `json:"id"`
	Name         string          `json:"name"`
	Ingredients  []IngredientNew `json:"ingredients"`
	Steps        []string        `json:"steps"`
	Notes        string          `json:"notes"`
	DishMadeDate *time.Time      `json:"dishMadeDate,omitempty"`
}

type Ingredient struct {
	ID                 int      `json:"id"`
	RecipeID           int      `json:"recipeId"`
	Product            string   `json:"product"`
	RecipeQuantity     Quantity `json:"recipeQuantity"`
	NormalizedQuantity Quantity `json:"normalizedQuantity"`
	CutStyle           string   `json:"cutStyle"`
}

type Ingredients []Ingredient

func (ingredients Ingredients) GetProductNames() []string {
	var productNames []string
	for _, ingredient := range ingredients {
		productNames = append(productNames, ingredient.Product)
	}
	return productNames
}

type CalculatedMealNutritionalValue struct {
	NutritionalValue  NutritionalValue                   `json:"nutritionalValue"`
	CalculatedRecipes []CalculatedRecipeNutritionalValue `json:"calculatedRecipes"`
}

type CalculatedRecipeNutritionalValue struct {
	RecipeID           int                                 `json:"recipeId"`
	RecipeName         string                              `json:"name"`
	NutritionalValue   NutritionalValue                    `json:"nutritionalValue"`
	CalculatedProducts []CalculatedProductNutritionalValue `json:"calculatedProducts"`
}

type CalculatedProductNutritionalValue struct {
	Product          string           `json:"product"`
	Message          string           `json:"message"`
	NutritionalValue NutritionalValue `json:"nutritionalValue"`
}

type CalculatedMealPrice struct {
	Price             float64                 `json:"price"`
	CalculatedRecipes []CalculatedRecipePrice `json:"calculatedRecipes"`
}

type CalculatedRecipePrice struct {
	RecipeID           int                      `json:"recipeId"`
	RecipeName         string                   `json:"name"`
	Price              float64                  `json:"price"`
	CalculatedProducts []CalculatedProductPrice `json:"calculatedProducts"`
}

type CalculatedProductPrice struct {
	Product string  `json:"product"`
	Message string  `json:"message"`
	Price   float64 `json:"price"`
}
