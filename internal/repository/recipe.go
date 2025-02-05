package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/uerror"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RecipeRepo struct {
	DB *pgxpool.Pool
}

func NewRecipeRepo(db *pgxpool.Pool) *RecipeRepo {
	return &RecipeRepo{DB: db}
}

func (r *RecipeRepo) InsertRecipe(ctx context.Context, recipe model.RecipeNew) error {
	query := `
	INSERT INTO recipes 
		(recipe_name, steps, notes, dish_made_date) 
	VALUES 
		($1, $2, $3, $4) 
	RETURNING id`
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var id int
	if err := tx.QueryRow(ctx, query, recipe.Name, recipe.Steps, recipe.Notes, recipe.DishMadeDate).Scan(&id); err != nil {
		return err
	}
	if err := insertIngredients(ctx, tx, id, recipe.Ingredients); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func insertIngredients(ctx context.Context, tx pgx.Tx, recipeID int, ingredients []model.IngredientNew) error {
	rows := make([][]interface{}, 0, len(ingredients))
	for _, ingredient := range ingredients {
		row := []interface{}{
			recipeID, ingredient.Product, ingredient.Unit, ingredient.Amount, ingredient.Notes}
		rows = append(rows, row)
	}

	_, err := tx.CopyFrom(ctx,
		pgx.Identifier{"recipe_ingredients"},
		[]string{"recipe_id", "product_name", "unit", "amount", "notes"},
		pgx.CopyFromRows(rows),
	)

	return err
}

func (r *RecipeRepo) GetRecipeSummaries(ctx context.Context) ([]model.RecipeSummary, error) {
	query := `SELECT id, recipe_name, steps, notes, dish_made_date FROM recipes ORDER BY dish_made_date DESC NULLS LAST`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipeSummaries []model.RecipeSummary
	for rows.Next() {
		var rs model.RecipeSummary
		var dishMadeDate *pgtype.Date
		if err := rows.Scan(&rs.ID, &rs.Name, &rs.Steps, &rs.Notes, &dishMadeDate); err != nil {
			return nil, err
		}
		if dishMadeDate != nil {
			rs.DishMadeDate = dishMadeDate.Time.Format(time.DateOnly)
		}
		recipeSummaries = append(recipeSummaries, rs)
	}
	return recipeSummaries, nil
}

func (r *RecipeRepo) GetRecipe(ctx context.Context, recipeID int) (model.Recipe, error) {
	query := `
	SELECT id, recipe_name, steps, notes, dish_made_date 
	FROM recipes 
	WHERE id = $1`

	var recipe model.Recipe
	err := r.DB.QueryRow(ctx, query, recipeID).Scan(&recipe.ID, &recipe.Name, &recipe.Steps, &recipe.Notes, &recipe.DishMadeDate)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Recipe{}, uerror.NewNotFound("nutritional value not found", err)
	}
	if err != nil {
		return model.Recipe{}, fmt.Errorf("query recipe: %w", err)
	}

	ingredients, err := r.getRecipeIngredients(ctx, recipeID)
	if err != nil {
		return model.Recipe{}, fmt.Errorf("get recipe ingredients: %w", err)
	}
	recipe.Ingredients = ingredients

	return recipe, nil
}

func (r *RecipeRepo) getRecipeIngredients(ctx context.Context, recipeID int) ([]model.Ingredient, error) {
	query := `
	SELECT id, recipe_id, product_name, unit, amount, notes 
	FROM recipe_ingredients 
	WHERE recipe_id = $1`

	rows, err := r.DB.Query(ctx, query, recipeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ingredients []model.Ingredient
	for rows.Next() {
		var i model.Ingredient
		if err := rows.Scan(&i.ID, &i.RecipeID, &i.Product,
			&i.Unit, &i.Amount, &i.Notes); err != nil {
			return nil, err
		}
		ingredients = append(ingredients, i)
	}
	return ingredients, nil
}

func (r *RecipeRepo) UpdateRecipe(ctx context.Context, recipe model.RecipeUpdate) error {
	query := `
	UPDATE recipes 
	SET recipe_name = $1, steps = $2, notes = $3, dish_made_date = $4
	WHERE id = $5`

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	status, err := tx.Exec(ctx, query, recipe.Name, recipe.Steps, recipe.Notes, recipe.DishMadeDate, recipe.ID)
	if err != nil {
		return err
	}

	if status.RowsAffected() != 1 {
		return uerror.NewNotFound(fmt.Sprintf("recipe with id %q does not exist", recipe.ID), nil)
	}

	if err := deleteRecipeIngredients(ctx, tx, recipe.ID); err != nil {
		return err
	}

	if err := insertIngredients(ctx, tx, recipe.ID, recipe.Ingredients); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func deleteRecipeIngredients(ctx context.Context, tx pgx.Tx, recipeID int) error {
	query := `DELETE FROM recipe_ingredients WHERE recipe_id = $1`

	if _, err := tx.Exec(ctx, query, recipeID); err != nil {
		return err
	}
	return nil
}

func (r *RecipeRepo) DeleteRecipe(ctx context.Context, recipeID int) error {

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := deleteRecipeIngredients(ctx, tx, recipeID); err != nil {
		return err
	}

	query := `DELETE FROM recipes WHERE id = $1`
	if _, err := tx.Exec(ctx, query, recipeID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *RecipeRepo) GetRecipesIngredients(ctx context.Context, recipeIDs []int) (model.Ingredients, error) {
	query := `
	SELECT id, recipe_id, product_name, unit, amount, notes 
	FROM recipe_ingredients 
	WHERE recipe_id = ANY($1)`

	rows, err := r.DB.Query(ctx, query, recipeIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ingredients []model.Ingredient
	for rows.Next() {
		var i model.Ingredient
		if err := rows.Scan(&i.ID, &i.RecipeID, &i.Product,
			&i.Unit, &i.Amount,
			&i.Notes); err != nil {
			return nil, err
		}
		ingredients = append(ingredients, i)
	}

	return ingredients, nil
}

func (r *RecipeRepo) GetRecipeIDsByDate(ctx context.Context, date time.Time) ([]int, error) {
	query := `
	SELECT id 
	FROM recipes 
	WHERE dish_made_date = $1`

	rows, err := r.DB.Query(ctx, query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipeIDs []int
	for rows.Next() {
		var recipeID int
		if err := rows.Scan(&recipeID); err != nil {
			return nil, err
		}
		recipeIDs = append(recipeIDs, recipeID)
	}

	return recipeIDs, nil
}

func (r *RecipeRepo) CloneRecipes(ctx context.Context, recipeIDs []int, date string, ingredientsByRecipeID map[int]model.Ingredients) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, recipeID := range recipeIDs {
		id, err := cloneRecipe(ctx, tx, recipeID, date)
		if err != nil {
			return fmt.Errorf("clone recipe: %w", err)
		}
		if err := insertIngredients(ctx, tx, id, ingredientsByRecipeID[recipeID].ToNewIngredients()); err != nil {
			return fmt.Errorf("insert ingredients: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func cloneRecipe(ctx context.Context, tx pgx.Tx, recipeID int, date string) (int, error) {
	query := `INSERT INTO recipes (recipe_name, steps, notes, dish_made_date)
SELECT recipe_name, steps, notes, $1
FROM recipes
WHERE id = $2 RETURNING id;
`
	var id int
	if err := tx.QueryRow(ctx, query, date, recipeID).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil

}
