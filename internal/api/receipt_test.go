package api

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

// abs returns the absolute value of a float64.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// normalizeUnit attempts to standardize unit strings for comparison.
// This is a simple example; a more robust solution might be needed for production.
func normalizeUnit(unit string) string {
	u := strings.ToLower(strings.TrimSpace(unit))
	switch u {
	case "lb", "lbs", "pound", "pounds":
		return "lb"
	case "kg", "kgs", "kilogram", "kilograms":
		return "kg"
	case "g", "gs", "gram", "grams":
		return "g"
	case "oz", "ounce", "ounces":
		return "oz"
	case "pc", "pcs", "piece", "pieces", "ea", "each", "unit", "un":
		return "item" // Consolidate various "each" type units to "item"
	case "gallon", "gallons", "gal":
		return "gallon"
	case "liter", "liters", "l":
		return "liter"
	case "ml", "milliliter", "milliliters":
		return "ml"
	case "meter", "meters", "m":
		return "meter"
	// Add more normalizations as needed
	default:
		return u
	}
}

// TestParseWithAI_Integration performs integration tests against the live Gemini API.
// It requires the GEMINI_API_KEY environment variable to be set.
// WARNING: This test makes actual API calls and may incur costs and is subject to API flakiness.
func TestParseWithAI_Integration(t *testing.T) {
	godotenv.Load("../../.env") // Load environment variables from .env file
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Fatal("Skipping integration test: GEMINI_API_KEY environment variable is not set.")
	}

	// This timeout is for the entire test case, including the API call.
	// The genai client might have its own internal timeouts for requests.
	defaultTestTimeout := 45 * time.Second // Increased timeout for potentially slower API responses

	testCases := []struct {
		name               string
		receiptInputText   string    // Raw text input for the AI
		expectedProducts   []Product // Validate only Price, Quantity, Unit after normalization
		expectError        bool      // True if we expect an error from parseWithAI or parseReceiptString
		errorContains      string    // If expectError is true, check if error message contains this substring
		allowEmptyProducts bool      // True if an empty product list is an acceptable outcome for the input
		skip               bool      // Skip this test case if true
	}{
		{
			name:             "maxima_1",
			receiptInputText: receipts[0],
			expectedProducts: []Product{
				{Price: 0.89},
				{Price: 2.70, Quantity: 2, Unit: "vnt"},
				{Price: 2.49},
				{Price: 1.71, Quantity: 1422, Unit: "g"},
				{Price: 2.15},
				{Price: 1.79},
				{Price: 1.64, Quantity: 784, Unit: "g"},
				{Price: 1.64, Quantity: 588, Unit: "g"},
			},
			expectError: false,
		},
		{
			name:             "norfa_1",
			receiptInputText: receipts[1],
			expectedProducts: []Product{
				{Price: 1.69, Quantity: 470, Unit: "g"},
				{Price: 1.11, Quantity: 375, Unit: "g"},
				{Price: 1.79, Quantity: 10, Unit: "vnt"},
				{Price: 1.98, Quantity: 500, Unit: "g"},
				{Price: 1.27, Quantity: 1300, Unit: "g"},
				{Price: 2.22, Quantity: 1266, Unit: "g"},
				{Price: 1.39, Quantity: 130, Unit: "g"},
			},
			skip: true, // Skip this test case for now due to API issues
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {

			if tc.skip {
				t.Skipf("Skipping test case: %s", tc.name)
			}

			ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
			defer cancel()

			t.Logf("Sending to AI: \n%s\n", tc.receiptInputText)
			receipt, err := parseWithAI(ctx, tc.receiptInputText)

			if err != nil {
				t.Fatalf("parseWithAI() failed unexpectedly: %v. Input was: %s", err, tc.receiptInputText)
			}

			if receipt == nil {
				t.Fatalf("parseReceiptString() returned nil receipt, but no error was expected")
			}

			if !tc.allowEmptyProducts && len(receipt.Products) == 0 && len(tc.expectedProducts) > 0 {
				t.Fatalf("Received no products from AI, but expected %d", len(tc.expectedProducts))
			}

			if len(receipt.Products) != len(tc.expectedProducts) {
				t.Fatalf("Number of products mismatch: got %d, want %d.\nGot Products: %+v\nExpected Products: %+v",
					len(receipt.Products), len(tc.expectedProducts), receipt.Products, tc.expectedProducts)
			}

			// Use a small epsilon for float comparisons
			const priceEpsilon = 0.01     // For currency, e.g., +/- 1 cent
			const quantityEpsilon = 0.001 // For general quantities

			for i, gotProd := range receipt.Products {
				if i >= len(tc.expectedProducts) {
					t.Errorf("More products received than expected. Index %d, Got: %+v", i, gotProd)
					continue
				}
				wantProd := tc.expectedProducts[i]

				if abs(gotProd.Price-wantProd.Price) > priceEpsilon {
					t.Errorf("Product %d (Name: '%s'): Price got = %.2f, want = %.2f.",
						i, gotProd.Name, gotProd.Price, wantProd.Price)
				}

				if abs(gotProd.Quantity-wantProd.Quantity) > quantityEpsilon {
					t.Errorf("Product %d (Name: '%s'): Quantity got = %.3f, want = %.3f.",
						i, gotProd.Name, gotProd.Quantity, wantProd.Quantity)
				}

				normalizedGotUnit := normalizeUnit(gotProd.Unit)
				normalizedWantUnit := normalizeUnit(wantProd.Unit)

				// Allow "item" as a fallback if wantUnit is more specific but quantity is 1
				// or if AI returns empty string for unit and we expect "item" for quantity 1.
				if normalizedGotUnit != normalizedWantUnit {
					isAcceptableUnitMismatch := false
					if (normalizedWantUnit == "item" && (normalizedGotUnit == "" || normalizedGotUnit == "item")) ||
						(normalizedGotUnit == "item" && (normalizedWantUnit == "" || normalizedWantUnit == "item")) {
						// If one is "item" and other is "" (empty string often means 1 item), consider it okay for quantity 1.
						if abs(gotProd.Quantity-1.0) < quantityEpsilon || (gotProd.Quantity == 0 && wantProd.Quantity == 0) {
							isAcceptableUnitMismatch = true
						}
					} else if normalizedWantUnit == "loaf" && normalizedGotUnit == "item" { // Specific case from example
						isAcceptableUnitMismatch = true
					}

					if !isAcceptableUnitMismatch {
						t.Errorf("Product %d (Name: '%s'): Unit got = '%s' (norm: '%s'), want = '%s' (norm: '%s').",
							i, gotProd.Name, gotProd.Unit, normalizedGotUnit, wantProd.Unit, normalizedWantUnit)
					}
				}
			}
			if t.Failed() {
				// Log the full AI response again if any assertion within the loop failed
				t.Logf("Review AI Response due to test failures: %v", receipt)
			}
		})
	}
}

var receipts = []string{
	`
MAXIMA LT, UAB
J0                                               #00048706
Kasininkas (-ė): 00001AA68D5A                          #
Kvitas bazėje: 111/175                                 #
Kramtomoji guma ORBIT WHITE SPEARMINT             0,89 A
Juoda raikyta AGOTOS duona su saulėgrąžomis       3,38 A
 1,69 X 2 vnt.
Nuolaida:Juoda raikyta AGOTOS duona su saulėgrąžo
                                                -0,68 A
Ekologiškas sojų gėrimas WELL DONE                2,49 A
Bananai                                           1,92 A
 1,35 X 1,422 kg
Nuolaida prekei                                  -0,21 A
Sūris salotoms GRIKIOS, 45 % rieb. s. m.          2,69 A
Nuolaida:Sūris salotoms GRIKIOS, 45 % rieb. s. m.
                                                -0,54 A
Valgomasis natūralus jogurtas GRAIKIŠKA AMFORA, 3,9 %
rieb                                              1,99 A
Nuolaida:Valgomasis natūralus jogurtas GRAIKIŠKA -0,20 A
Pomidorai su šakelėmis, 47-57 mm                  2,34 A
 2,99 X 0,784 kg
Nuolaida:Pomidorai su šakelėmis, 47-57 mm        -0,70 A
Lietuviški trumpavaisiai agurkai                  2,93 A
 4,99 X 0,588 kg
Nuolaida:Lietuviški trumpavaisiai agurkai        -1,29 A
====================================================== #
Suteiktos naudos                                       #
Su AČIŪ kortele sutaupėte                        -3,40 #
Pritaikytos nuolaidos                            -3,42 #
========================================================
Nuolaidos                                           6,82
--------------------------------------------------------
                     Be PVM           PVM        Su PVM
- - - - - - - - - - - - - - - - - - - - - - - - - - - - 
A= 21,00%              14,96          3,14         18,10
========================================================
Kvito suma                                         18,10
- - - - - - - - - - - - - - - - - - - - - - - - - - - - 
MAXIMA`,
	`D1_UAB NORFOS MAŽMENA
------------------------------------------------
# Kvito numeris 27424 #
Sumuštinių duona AGOTOS su grūdais ir sėkl., 470
g 1,69 M1
*Juoda AGOTOS duona su saulėg 375g 1,11 M1
*Kiaušiniai A M rudi, 10 vnt 1,79 M1
Varškė N 9% rieb., 500g 1,98 M1
*Apelsinai 4/5dyd., 1kg 1,3x0,98 1,27 M1
Bananai, 1kg 1,266x1,75 2,22 M1
*Trašk. TAFFEL S.CREAM & ONION CHIPS, 130g
1,99 M1
30% nuolaida su KUPONU -0,60 EUR
AKCIJŲ NUOLAIDA 0,60 EUR #
******************************************* #
KVITO SUMA 16,10 EUR #
2025-04-27 15:50 KAS#45075
`,
}
