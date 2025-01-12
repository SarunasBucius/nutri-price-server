package barbora

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/stretchr/testify/require"
)

const receiptExample = `Barbora
2023-04-16
1 Nektarinai, 1 kg 0.612 kg €1.6569 €1.3693 21,00 €0.84 €1.01
2 Salotos ROMAINE, 300 g 1 vnt. €1.9900 €1.6446 21,00 €1.64 €1.99
Pritaikytos nuolaidos
Nektarinai, 1 kg -€1.10`

func TestBarboraParser_ParseDate(t *testing.T) {
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
			fields: fields{ReceiptLines: strings.Split(receiptExample, "\n"), Retailer: "barbora"},
			want:   time.Date(2023, 4, 16, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := BarboraParser{
				ReceiptLines: tt.fields.ReceiptLines,
				Retailer:     tt.fields.Retailer,
			}
			got, err := p.ParseDate()
			if (err != nil) != tt.wantErr {
				t.Errorf("BarboraParser.ParseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BarboraParser.ParseDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBarboraParser_ParseProducts(t *testing.T) {
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
			fields: fields{ReceiptLines: strings.Split(receiptExample, "\n"), Retailer: "barbora"},
			want: model.ReceiptProducts{
				{
					Name: "Nektarinai, 1 kg",
					Price: model.Price{
						Discount: 1.10,
						Paid:     1.01,
						Full:     2.11,
					},
					Quantity: model.Quantity{
						Amount: 612,
						Unit:   model.Grams,
					},
				},
				{
					Name: "Salotos ROMAINE, 300 g",
					Price: model.Price{
						Discount: 0,
						Paid:     1.99,
						Full:     1.99,
					},
					Quantity: model.Quantity{
						Amount: 1,
						Unit:   model.Pieces,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := BarboraParser{
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
