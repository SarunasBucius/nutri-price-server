package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/uerror"
	"github.com/go-chi/chi/v5"
)

type RecipeAPI struct {
	Service IRecipeService
}

func NewRecipeAPI(recipeService IRecipeService) *RecipeAPI {
	return &RecipeAPI{Service: recipeService}
}

type IRecipeService interface {
	InsertRecipe(ctx context.Context, recipe model.RecipeNew) error
	GetRecipeSummaries(ctx context.Context) ([]model.RecipeSummary, error)
	GetRecipe(ctx context.Context, recipeID int) (model.Recipe, error)
	UpdateRecipe(ctx context.Context, recipe model.RecipeUpdate) error
	GetMealPrice(ctx context.Context, recipeIDs []int) (model.CalculatedMealPrice, error)
	GetMealNutritionalValue(ctx context.Context, recipeIDs []int) (model.CalculatedMealNutritionalValue, error)
	DeleteRecipe(ctx context.Context, recipeID int) error
	GetMealPriceByDate(ctx context.Context, date time.Time) (model.CalculatedMealPrice, error)
	GetMealNutritionalValueByDate(ctx context.Context, date time.Time) (model.CalculatedMealNutritionalValue, error)
	CloneRecipes(ctx context.Context, recipeIDs []int, date string) error
}

func (rc *RecipeAPI) InsertRecipe(w http.ResponseWriter, r *http.Request) {
	var recipe model.RecipeNew
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid request body", err))
		return
	}

	if err := rc.Service.InsertRecipe(r.Context(), recipe); err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, newSuccessMessage("successfully inserted recipe"))
}

func (rc *RecipeAPI) GetRecipeSummaries(w http.ResponseWriter, r *http.Request) {
	recipeNames, err := rc.Service.GetRecipeSummaries(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, emptyIfNil(recipeNames))
}

func (rc *RecipeAPI) GetRecipe(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "recipeID")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid id", err))
		return
	}

	recipe, err := rc.Service.GetRecipe(r.Context(), id)
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, recipe)
}

func (rc *RecipeAPI) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "recipeID")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid id", err))
		return
	}

	var recipe model.RecipeUpdate
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid request body", err))
		return
	}
	recipe.ID = id

	if err := rc.Service.UpdateRecipe(r.Context(), recipe); err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, newSuccessMessage("successfully updated recipe"))
}

func (rc *RecipeAPI) GetMealNutritionalValue(w http.ResponseWriter, r *http.Request) {
	idsParam := chi.URLParam(r, "recipeIDs")

	ids, err := numbersParamToInts(idsParam)
	if err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid ids", err))
		return
	}

	calculatedMeal, err := rc.Service.GetMealNutritionalValue(r.Context(), ids)
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, calculatedMeal)
}

func (rc *RecipeAPI) GetMealPrice(w http.ResponseWriter, r *http.Request) {
	idsParam := chi.URLParam(r, "recipeIDs")

	ids, err := numbersParamToInts(idsParam)
	if err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid ids", err))
		return
	}

	calculatedMeal, err := rc.Service.GetMealPrice(r.Context(), ids)
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, calculatedMeal)
}

func (rc *RecipeAPI) GetMealNutritionalValueByDate(w http.ResponseWriter, r *http.Request) {
	dateParam := chi.URLParam(r, "date")
	date, err := time.Parse(time.DateOnly, dateParam)
	if err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid date", err))
		return
	}

	calculatedMeal, err := rc.Service.GetMealNutritionalValueByDate(r.Context(), date)
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, calculatedMeal)
}

func (rc *RecipeAPI) GetMealPriceByDate(w http.ResponseWriter, r *http.Request) {
	dateParam := chi.URLParam(r, "date")
	date, err := time.Parse(time.DateOnly, dateParam)
	if err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid date", err))
		return
	}

	calculatedMeal, err := rc.Service.GetMealPriceByDate(r.Context(), date)
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, calculatedMeal)
}

func (rc *RecipeAPI) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "recipeID")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid id", err))
		return
	}

	if err := rc.Service.DeleteRecipe(r.Context(), id); err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, newSuccessMessage("successfully deleted recipe"))
}

func (rc *RecipeAPI) CloneRecipes(w http.ResponseWriter, r *http.Request) {
	var cloneRecipesReq model.CloneRecipesRequest
	if err := json.NewDecoder(r.Body).Decode(&cloneRecipesReq); err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid request body", err))
		return
	}

	if err := rc.Service.CloneRecipes(r.Context(), cloneRecipesReq.RecipeIDs, cloneRecipesReq.Date); err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, newSuccessMessage("successfully cloned recipes"))
}
