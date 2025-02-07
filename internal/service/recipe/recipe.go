package recipe

import (
	"context"
	"fmt"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
)

type Service struct {
	ProductRepo          IProductRepository
	NutritionalValueRepo INutritionalValueRepository
	RecipeRepo           IRecipeRepository
}

func NewRecipeService(productRepo IProductRepository, nutritionalValueRepo INutritionalValueRepository, recipeRepo IRecipeRepository) *Service {
	return &Service{
		ProductRepo:          productRepo,
		NutritionalValueRepo: nutritionalValueRepo,
		RecipeRepo:           recipeRepo,
	}
}

type INutritionalValueRepository interface {
	GetProductsNutritionalValueByProductNames(ctx context.Context, productNames []string) ([]model.ProductNutritionalValue, error)
	InsertEmptyProducts(ctx context.Context, products []string) error
}

type IRecipeRepository interface {
	InsertRecipe(ctx context.Context, recipe model.RecipeNew) error
	GetRecipeSummaries(ctx context.Context) (model.RecipeSummaries, error)
	GetRecipe(ctx context.Context, recipeID int) (model.Recipe, error)
	UpdateRecipe(ctx context.Context, recipe model.RecipeUpdate) error
	DeleteRecipe(ctx context.Context, recipeID int) error
	GetRecipesIngredients(ctx context.Context, recipeIDs []int) (model.Ingredients, error)
	GetRecipeIDsByDate(ctx context.Context, date time.Time) ([]int, error)
	CloneRecipes(ctx context.Context, recipeIDs []int, date string, ingredientsByRecipeID map[int]model.Ingredients) error
}

type IProductRepository interface {
	GetLastBoughtProductsByNamesOrGroups(ctx context.Context, products []string) ([]model.PurchasedProduct, error)
}

func (s *Service) InsertRecipe(ctx context.Context, recipe model.RecipeNew) error {
	if err := s.insertEmptyProducts(ctx, recipe.Ingredients); err != nil {
		return err
	}

	if err := s.RecipeRepo.InsertRecipe(ctx, recipe); err != nil {
		return fmt.Errorf("insert recipe: %w", err)
	}
	return nil
}

func (s *Service) GetRecipeSummaries(ctx context.Context) (model.RecipeSummaries, error) {
	recipesNames, err := s.RecipeRepo.GetRecipeSummaries(ctx)
	if err != nil {
		return model.RecipeSummaries{}, fmt.Errorf("get recipes names: %w", err)
	}
	return recipesNames, nil
}

func (s *Service) GetRecipe(ctx context.Context, recipeID int) (model.Recipe, error) {
	recipe, err := s.RecipeRepo.GetRecipe(ctx, recipeID)
	if err != nil {
		return model.Recipe{}, fmt.Errorf("get recipe: %w", err)
	}
	return recipe, nil
}

func (s *Service) UpdateRecipe(ctx context.Context, recipe model.RecipeUpdate) error {
	if err := s.insertEmptyProducts(ctx, recipe.Ingredients); err != nil {
		return err
	}

	if err := s.RecipeRepo.UpdateRecipe(ctx, recipe); err != nil {
		return fmt.Errorf("update recipe: %w", err)
	}
	return nil
}

func (s *Service) DeleteRecipe(ctx context.Context, recipeID int) error {
	if err := s.RecipeRepo.DeleteRecipe(ctx, recipeID); err != nil {
		return fmt.Errorf("delete recipe: %w", err)
	}
	return nil
}

func (s *Service) GetMealPrice(ctx context.Context, recipeIDs []int) (model.CalculatedMealPrice, error) {
	ingredients, err := s.RecipeRepo.GetRecipesIngredients(ctx, recipeIDs)
	if err != nil {
		return model.CalculatedMealPrice{}, fmt.Errorf("get recipes ingredients: %w", err)
	}

	products, err := s.ProductRepo.GetLastBoughtProductsByNamesOrGroups(ctx, ingredients.GetProductNames())
	if err != nil {
		return model.CalculatedMealPrice{}, fmt.Errorf("get last bought products by names: %w", err)
	}

	return calculateMealPrice(ingredients, products), nil
}

func (s *Service) GetMealPriceByDate(ctx context.Context, date time.Time) (model.CalculatedMealPrice, error) {
	recipeIDs, err := s.RecipeRepo.GetRecipeIDsByDate(ctx, date)
	if err != nil {
		return model.CalculatedMealPrice{}, fmt.Errorf("get recipe IDs by date: %w", err)
	}

	return s.GetMealPrice(ctx, recipeIDs)
}

func (s *Service) GetMealNutritionalValue(ctx context.Context, recipeIDs []int) (model.CalculatedMealNutritionalValue, error) {
	ingredients, err := s.RecipeRepo.GetRecipesIngredients(ctx, recipeIDs)
	if err != nil {
		return model.CalculatedMealNutritionalValue{}, fmt.Errorf("get recipes ingredients: %w", err)
	}

	productsNutritionalValue, err := s.NutritionalValueRepo.GetProductsNutritionalValueByProductNames(ctx, ingredients.GetProductNames())
	if err != nil {
		return model.CalculatedMealNutritionalValue{}, fmt.Errorf("get products nutritional value: %w", err)
	}

	return calculateMealNutritionalValue(ingredients, productsNutritionalValue), nil
}

func (s *Service) GetMealNutritionalValueByDate(ctx context.Context, date time.Time) (model.CalculatedMealNutritionalValue, error) {
	recipeIDs, err := s.RecipeRepo.GetRecipeIDsByDate(ctx, date)
	if err != nil {
		return model.CalculatedMealNutritionalValue{}, fmt.Errorf("get recipe IDs by date: %w", err)
	}

	return s.GetMealNutritionalValue(ctx, recipeIDs)
}

func (s *Service) CloneRecipes(ctx context.Context, recipeIDs []int, date string) error {
	ingredients, err := s.RecipeRepo.GetRecipesIngredients(ctx, recipeIDs)
	if err != nil {
		return fmt.Errorf("get recipes ingredients: %w", err)
	}

	ingredientsByRecipeID := make(map[int]model.Ingredients)
	for _, ingredient := range ingredients {
		ingredientsByRecipeID[ingredient.RecipeID] = append(ingredientsByRecipeID[ingredient.RecipeID], ingredient)
	}

	if err := s.RecipeRepo.CloneRecipes(ctx, recipeIDs, date, ingredientsByRecipeID); err != nil {
		return fmt.Errorf("clone recipes: %w", err)
	}
	return nil
}

func (s *Service) insertEmptyProducts(ctx context.Context, ingredients []model.IngredientNew) error {
	ingredientNames := make([]string, 0, len(ingredients))
	for _, ingredient := range ingredients {
		ingredientNames = append(ingredientNames, ingredient.Product)
	}

	if err := s.NutritionalValueRepo.InsertEmptyProducts(ctx, ingredientNames); err != nil {
		return fmt.Errorf("insert empty products: %w", err)
	}
	return nil
}
