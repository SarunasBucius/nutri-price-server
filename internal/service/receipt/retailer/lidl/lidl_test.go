package lidl

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

const receiptExample = `
UAB "Lidl Lietuva" Į. k.: 111791015          
           ...          
PVM mokėtojo kodas LT117910113                        
Kvitas 64/128                                #00017499
7602643   Pop.pirk.maišelis                     0,20 A
0082231   Ilgavaisis agurkas                    0,99 A
0048087   Feta sūris                            2,69 A
7801405   Vištų kiaušiniai L                    1,75 A
5900018   Virti avinžirniai                     1,49 A
0083836   Tamsūs smulk. slyv. pomidorai         2,49 A
5528568   Brokoliai                             1,75 A
0159177   Tamsusis šokoladas            
  0,99 X 2,000 vnt.                             1,98 A
0082615   Paprikos raudonosios          
  3,49 X 0,298 KG                               1,04 A
0080505   Vynuogės žal.be kaul                  1,29 A
0080206   Obuol. Crimson Snow           
  1,89 X 1,232 KG                               2,33 A
7605416   Juod.duon.su saulėg.                  1,79 A
Taikoma nuolaida                        
Nuolaida                                       -0,54 A
0082180   Baklažanai                    
  2,99 X 0,354 KG                               1,06 A
0003010   Vienkart. maišelis                    0,01 A
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
