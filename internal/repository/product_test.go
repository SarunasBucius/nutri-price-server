package repository

import (
	"context"
	"testing"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/jackc/pgx/v5"
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
						Name:     "red apples",
						Price:    model.Price{Discount: 0.10, Paid: 1, Full: 1.1},
						Quantity: model.Quantity{Unit: model.Grams, Amount: 500},
						Group:    "apples",
						Notes:    "best ones",
					},
					{
						Name:     "lentils",
						Price:    model.Price{Paid: 1, Full: 1},
						Quantity: model.Quantity{Unit: model.Grams, Amount: 500},
						Group:    "lentils",
					},
				},
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
			err = r.InsertProducts(tt.args.ctx, tt.args.retailer, tt.args.purchaseDate, tt.args.products)
			if tt.wantErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
		})
	}
}

func (s *ContainerTestSuite) TestProductRepo_GetProductGroups() {
	ctx := context.Background()

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name       string
		args       args
		insertData func(db *pgxpool.Pool)
		want       []string
		wantErr    bool
	}{
		{
			name: "get_product_groups",
			args: args{ctx: ctx},
			insertData: func(db *pgxpool.Pool) {
				_, err := db.CopyFrom(ctx,
					pgx.Identifier{"purchased_products"},
					[]string{"product_name", "retailer", "product_group", "measurement_unit", "quantity", "full_price", "paid_price", "discount", "notes", "purchase_date"},
					pgx.CopyFromRows([][]interface{}{
						{"red apples", "norfa", "apples", "pieces", "1", 1, 0.90, 0.1, "", "2024-10-14"},
					}),
				)
				s.Require().NoError(err)
			},
			want: []string{"apples"},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			err := s.Container.Restore(ctx, postgres.WithSnapshotName("emptyTables"))
			s.Require().NoError(err)

			db, err := pgxpool.New(ctx, s.Container.MustConnectionString(ctx))
			require.NoError(t, err)
			defer db.Close()

			if tt.insertData != nil {
				tt.insertData(db)
			}

			r := NewProductRepo(db)
			got, err := r.GetProductGroups(tt.args.ctx)
			if tt.wantErr {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(got, tt.want)
		})
	}
}
