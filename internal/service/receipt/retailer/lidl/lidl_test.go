package lidl

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/stretchr/testify/require"
)

const receiptExample = `UAB "Lidl Lietuva" Į. k.: 111791015          
           ...          
PVM mokėtojo kodas LT117910113                        
Kvitas 64/128                                #00017499
0159177   Tamsusis šokoladas            
  0,99 X 2,000 vnt.                             1,98 A
0080505   Vynuogės žal.be kaul                  1,29 A
0080206   Obuol. Crimson Snow           
  1,89 X 1,232 KG                               2,33 A
7605416   Juod.duon.su saulėg.                  1,79 A
Taikoma nuolaida                        
Nuolaida                                       -0,54 A
------------------------------------------------------
Tarpinė suma                                     20,32
======================================================
Nuolaida                                          0,54
 - - - - - - - - - - - - - - - - - - - - - - - - - - -
                       PVM        Be PVM        Su PVM
A=21,00%              3,53         16,79         20,32
======================================================
Mokėti                                           20,32
Mokėta (Banko kortelė)                           20,32

-----------------------------------------------------#
...
-----------------------------------------------------#
Kvito Nr. 0226-017637-86-20240416                    #
-----------------------------------------------------#
                Dėkojame, kad pirkote!                
             Nemokama klientų infolinija:             
                     8 800 10011                      
          PVM sąskaitos faktūros išrašomos:           
            www.saskaitosfakturos.lidl.lt             
Kasininkas (-e) 86                                    
LF NM0000006239BC                  2023-04-16 15:47:08`

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
			name:   "parse_date",
			fields: fields{ReceiptLines: strings.Split(receiptExample, "\n"), Retailer: "lidl"},
			want:   time.Date(2023, 4, 16, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := LidlParser{
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

func TestLidlParser_ParseProducts(t *testing.T) {
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
			fields: fields{ReceiptLines: strings.Split(receiptExample, "\n"), Retailer: "lidl"},
			want: model.ReceiptProducts{
				{
					Name:  "Tamsusis šokoladas",
					Price: 1.98,
					Quantity: model.Quantity{
						Unit:   model.Pieces,
						Amount: 2,
					},
				},
				{
					Name:  "Vynuogės žal.be kaul",
					Price: 1.29,
				},
				{
					Name:  "Obuol. Crimson Snow",
					Price: 2.33,
					Quantity: model.Quantity{
						Unit:   model.Grams,
						Amount: 1232,
					},
				},
				{
					Name:  "Juod.duon.su saulėg.",
					Price: 1.25,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := LidlParser{
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
