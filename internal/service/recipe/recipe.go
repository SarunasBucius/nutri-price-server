package recipe

import (
	"context"
	"fmt"

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
}

type IRecipeRepository interface {
	InsertRecipe(ctx context.Context, recipe model.RecipeNew) error
	GetRecipesNames(ctx context.Context) ([]string, error)
	GetRecipe(ctx context.Context, recipeID int) (model.Recipe, error)
	UpdateRecipe(ctx context.Context, recipe model.RecipeUpdate) error
	DeleteRecipe(ctx context.Context, recipeID int) error
	GetRecipesIngredients(ctx context.Context, recipeIDs []int) (model.Ingredients, error)
}

type IProductRepository interface {
	GetLastBoughtProductsByNamesOrGroups(ctx context.Context, products []string) ([]model.PurchasedProduct, error)
}

func (s *Service) InsertRecipe(ctx context.Context, recipe model.RecipeNew) error {
	if err := s.RecipeRepo.InsertRecipe(ctx, recipe); err != nil {
		return fmt.Errorf("insert recipe: %w", err)
	}
	return nil
}

func (s *Service) GetRecipesNames(ctx context.Context) ([]string, error) {
	recipesNames, err := s.RecipeRepo.GetRecipesNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("get recipes names: %w", err)
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
