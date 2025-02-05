package model

import "time"

type RecipeNew struct {
	Name         string          `json:"name"`
	Ingredients  []IngredientNew `json:"ingredients"`
	Steps        []string        `json:"steps"`
	Notes        string          `json:"notes"`
	DishMadeDate *string         `json:"dishMadeDate,omitempty"`
}

type IngredientNew struct {
	Product string  `json:"product"`
	Unit    Unit    `json:"unit"`
	Amount  float64 `json:"amount"`
	Notes   string  `json:"notes"`
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
	ID       int     `json:"id"`
	RecipeID int     `json:"recipeId"`
	Product  string  `json:"product"`
	Unit     Unit    `json:"unit"`
	Amount   float64 `json:"amount"`
	Notes    string  `json:"notes"`
}

type Ingredients []Ingredient

func (ingredients Ingredients) GetProductNames() []string {
	var productNames []string
	for _, ingredient := range ingredients {
		productNames = append(productNames, ingredient.Product)
	}
	return productNames
}

func (ingredients Ingredients) ToNewIngredients() []IngredientNew {
	newIngredients := make([]IngredientNew, 0, len(ingredients))
	for _, ingredient := range ingredients {
		newIngredients = append(newIngredients, IngredientNew{
			Product: ingredient.Product,
			Unit:    ingredient.Unit,
			Amount:  ingredient.Amount,
			Notes:   ingredient.Notes,
		})
	}
	return newIngredients
}

type RecipeSummary struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	Steps        []string `json:"steps"`
	Notes        string   `json:"notes"`
	DishMadeDate string   `json:"dishMadeDate,omitempty"`
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

type CloneRecipesRequest struct {
	RecipeIDs []int  `json:"recipeIds"`
	Date      string `json:"date"`
}
