package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/SarunasBucius/nutri-price-server/internal/utils/uerror"
)

type successMessage struct {
	Message string `json:"message"`
}

func errorResponse(ctx context.Context, w http.ResponseWriter, err error) {
	slog.ErrorContext(ctx, "responding to request with an error", "error", err)

	responseMessage, statusCode := uerror.SanitizeError(err)

	w.Header().Add("Content-Type", "application/json")
	response := map[string]string{"error": responseMessage}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.ErrorContext(ctx, "encoding error message to json", "error", err)
	}
	w.WriteHeader(statusCode)
}

func successResponse(ctx context.Context, w http.ResponseWriter, response any) {
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		errorResponse(ctx, w, err)
		return
	}
}

func newSuccessMessage(message string) successMessage {
	return successMessage{Message: message}
}

func emptyIfNil[S ~[]E, E any](s S) S {
	if s == nil {
		return []E{}
	}
	return s
}

func numbersParamToInts(numbers string) ([]int, error) {
	splitNumbers := strings.Split(numbers, ",")

	convertedNumbers := make([]int, 0, len(splitNumbers))
	for _, num := range splitNumbers {
		convertedNum, err := strconv.Atoi(num)
		if err != nil {
			return nil, err
		}
		convertedNumbers = append(convertedNumbers, convertedNum)
	}
	return convertedNumbers, nil
}
