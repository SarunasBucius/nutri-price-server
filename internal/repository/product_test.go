package repository

import (
	"context"
	"testing"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func (s *ContainerTestSuite) TestProductRepo_InsertProducts() {
	ctx := context.Background()

	err := s.Container.Restore(ctx, postgres.WithSnapshotName("emptyTables"))
	s.Require().NoError(err)

	type args struct {
		ctx          context.Context
		retailer     string
		purchaseDate time.Time
		products     []model.PurchasedProductNew
		productNames []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "insert_products",
			args: args{
				ctx:          ctx,
				retailer:     "norfa",
				purchaseDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				products: []model.PurchasedProductNew{
					{
						ProductID:   "1",
						Name:        "apples",
						VarietyName: "red",
						Price:       1,
						Quantity:    model.Quantity{Unit: model.Grams, Amount: 500},
						Notes:       "best ones",
					},
					{
						ProductID:   "1",
						Name:        "apples",
						VarietyName: "green",
						Price:       1,
						Quantity:    model.Quantity{Unit: model.Grams, Amount: 500},
					},
				},
				productNames: []string{"apples"},
			},
		},
		{
			name: "insert_products_empty",
			args: args{
				ctx:          ctx,
				retailer:     "norfa",
				purchaseDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				products:     []model.PurchasedProductNew{},
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := s.Container.Restore(ctx, postgres.WithSnapshotName("emptyTables"))
				s.Require().NoError(err)
			})
			db, err := pgxpool.New(ctx, s.Container.MustConnectionString(ctx))
			require.NoError(t, err)
			defer db.Close()

			r := NewProductRepo(db)
			err = r.InsertProducts(tt.args.ctx, tt.args.productNames)
			if tt.wantErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			err = r.InsertPurchases(tt.args.ctx, tt.args.retailer, tt.args.purchaseDate, tt.args.products)
			if tt.wantErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
		})
	}
}
