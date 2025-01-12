package maxima

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/stretchr/testify/require"
)

const receiptExample = `MAXIMA LT, UAB
...
PVM mokėtojo kodas LT230335113

Kvitas 198/1582                                #00408751
Raudonėliai SALDVA
  0,65 X 2 vnt.                                   1,30 A
Nuolaida:prieskonių pakuotėms -50%               -0,66 A
Visų grūdo dalių avižiniai dribsniai WELL DONE    1,29 A
Juodasis šokoladas (72 %) PERGALĖ                 4,99 A
Nuolaida:Juodasis šokoladas (72 %) PERGALĖ[      -2,00 A
Raudonos saldžiosios paprikos, 80-100 mm
  2,99 X 0,300 kg                                 0,90 A
Nuolaida:Raudonos saldžiosios paprikos, 80-100 mm
                                                 -0,45 A
Lietuviški trumpavaisiai agurkai
  2,49 X 0,514 kg                                 1,28 A
====================================================== #
LAIKAS             2024-09-12 19:26:07                 #`

func TestMaximaParser_ParseDate(t *testing.T) {
	type fields struct {
		ReceiptLines []string
		Retailer     string
	}
	tests := []struct {
		name    string
		fields  fields
		want    time.Time
		wantErr bool
	}{
		{
			name:   "parse_date",
			fields: fields{ReceiptLines: strings.Split(receiptExample, "\n"), Retailer: "maxima"},
			want:   time.Date(2024, 9, 12, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := MaximaParser{
				ReceiptLines: tt.fields.ReceiptLines,
				Retailer:     tt.fields.Retailer,
			}
			got, err := p.ParseDate()
			if (err != nil) != tt.wantErr {
				t.Errorf("MaximaParser.ParseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MaximaParser.ParseDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaximaParser_ParseProducts(t *testing.T) {
	type fields struct {
		ReceiptLines []string
		Retailer     string
	}
	tests := []struct {
		name    string
		fields  fields
		want    model.ReceiptProducts
		wantErr bool
	}{
		{
			name:   "parse_products",
			fields: fields{ReceiptLines: strings.Split(receiptExample, "\n"), Retailer: "maxima"},
			want: model.ReceiptProducts{
				{
					Name: "Raudonėliai SALDVA",
					Price: model.Price{
						Discount: 0.66,
						Paid:     0.64,
						Full:     1.30,
					},
					Quantity: model.Quantity{
						Unit:   model.Pieces,
						Amount: 2,
					},
				},
				{
					Name: "Visų grūdo dalių avižiniai dribsniai WELL DONE",
					Price: model.Price{
						Discount: 0,
						Paid:     1.29,
						Full:     1.29,
					},
				},
				{
					Name: "Juodasis šokoladas (72 %) PERGALĖ",
					Price: model.Price{
						Discount: 2.00,
						Paid:     2.99,
						Full:     4.99,
					},
				},
				{
					Name: "Raudonos saldžiosios paprikos, 80-100 mm",
					Price: model.Price{
						Discount: 0.45,
						Paid:     0.45,
						Full:     0.90,
					},
					Quantity: model.Quantity{
						Unit:   model.Grams,
						Amount: 300,
					},
				},
				{
					Name: "Lietuviški trumpavaisiai agurkai",
					Price: model.Price{
						Discount: 0,
						Paid:     1.28,
						Full:     1.28,
					},
					Quantity: model.Quantity{
						Unit:   model.Grams,
						Amount: 514,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := MaximaParser{
				ReceiptLines: tt.fields.ReceiptLines,
				Retailer:     tt.fields.Retailer,
			}
			got, err := p.ParseProducts()
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
