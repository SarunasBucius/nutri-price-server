package norfa

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
)

const receiptExample = `D1_
UAB NORFOS MAŽMENA
...
Visada laukiame Jūsų, Ačiū
------------------------------------------------
# Kvito numeris 00000 #
Ledai AURUM 100ml su kakaviniu glaistu 0,39 M1
*Salierų stiebai, 1kg 0,466x1,95 0,91 M1
Nuolaida 50% -0,45 EUR
Raudonieji lęšiai SKANĖJA, 500g 1,89 M1
AKCIJŲ NUOLAIDA 0,45 EUR #
******************************************* #
Lojalumo kortelė ************0000 #
******************************************* #
KVITO SUMA 4,23 EUR #
* prekėms lojalumo nuolaidos netaikomos #
NUOLAIDA PREKIŲ SUMAI 1,89 EUR #
Iš viso NORFA pinigų už kvitą 0,00 EUR #
NORFA pinigų likutis 0,00 EUR #
NORFA pinigai galioja iki 2023-09-15 #
TARPINĖ SUMA 4,23 EUR
------------------------------------------------
PVM1 (21,00 %) 0,73
------------------------------------------------
SUMA 4,23 EUR
BANKO KORTELE 4,23 EUR
2023-02-21 17:03 KAS#00000
@ LTF NV 0000000 00 00000`

func TestParser_ParseProducts(t *testing.T) {
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
			fields: fields{ReceiptLines: strings.Split(receiptExample, "\n"), Retailer: "norfa"},
			want: model.ReceiptProducts{
				{
					ProductLineInReceipt: "Ledai AURUM 100ml su kakaviniu glaistu 0,39 M1",
					PurchasedProductNew: model.PurchasedProductNew{
						Name:     "Ledai AURUM 100ml su kakaviniu glaistu",
						Price:    model.Price{Paid: 0.39, Full: 0.39},
						Quantity: model.Quantity{Unit: model.Milliliters, Amount: 100},
					},
				},
				{
					ProductLineInReceipt: "Salierų stiebai, 1kg 0,466x1,95 0,91 M1",
					PurchasedProductNew: model.PurchasedProductNew{
						Name:     "Salierų stiebai",
						Price:    model.Price{Paid: 0.46, Full: 0.91, Discount: 0.45},
						Quantity: model.Quantity{Unit: model.Grams, Amount: 466},
					},
				},
				{
					ProductLineInReceipt: "Raudonieji lęšiai SKANĖJA, 500g 1,89 M1",
					PurchasedProductNew: model.PurchasedProductNew{
						Name:     "Raudonieji lęšiai SKANĖJA, 500g",
						Price:    model.Price{Paid: 1.89, Full: 1.89},
						Quantity: model.Quantity{Unit: model.Grams, Amount: 500},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NorfaParser{
				ReceiptLines: tt.fields.ReceiptLines,
				Retailer:     tt.fields.Retailer,
			}
			got, err := p.ParseProducts()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.ParseProducts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.ParseProducts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_ParseDate(t *testing.T) {
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
			name:   "extract_date",
			fields: fields{ReceiptLines: strings.Split(receiptExample, "\n"), Retailer: "norfa"},
			want:   time.Date(2023, 2, 21, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "too_short_receipt",
			fields:  fields{ReceiptLines: []string{""}},
			want:    time.Time{},
			wantErr: true,
		},
		{
			name:    "invalid_date",
			fields:  fields{ReceiptLines: []string{"invalid_date", "line1"}},
			want:    time.Time{},
			wantErr: true,
		},
		{
			name:    "too_short_date_line",
			fields:  fields{ReceiptLines: []string{"2024-05-2", "line1"}},
			want:    time.Time{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NorfaParser{
				ReceiptLines: tt.fields.ReceiptLines,
				Retailer:     tt.fields.Retailer,
			}
			got, err := p.ParseDate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.ParseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.ParseDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
