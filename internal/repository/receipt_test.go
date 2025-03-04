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

func (s *ContainerTestSuite) TestReceiptRepo_insertParsedProducts() {
	ctx := context.Background()

	err := s.Container.Restore(ctx, postgres.WithSnapshotName("emptyTables"))
	s.Require().NoError(err)

	type args struct {
		ctx            context.Context
		parsedProducts model.ReceiptProducts
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "insert_parsed_products",
			args: args{
				ctx: ctx,
				parsedProducts: model.ReceiptProducts{
					{
						Name: "red apples",
					},
					{
						Name: "lentils",
					},
					{
						Name: "lentils",
					},
				},
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

			r := NewReceiptRepo(db)
			err = r.insertParsedProducts(tt.args.ctx, tt.args.parsedProducts)
			if tt.wantErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
		})
	}
}

func (s *ContainerTestSuite) TestReceiptRepo_GetUnconfirmedReceipt() {
	ctx := context.Background()

	err := s.Container.Restore(ctx, postgres.WithSnapshotName("emptyTables"))
	s.Require().NoError(err)
	type args struct {
		ctx      context.Context
		retailer string
		date     string
	}
	tests := []struct {
		name       string
		args       args
		insertData func(db *pgxpool.Pool)
		want       []model.PurchasedProductNew
		wantErr    bool
	}{
		{
			name: "get_unconfirmed_receipt",
			args: args{
				ctx:      ctx,
				retailer: "norfa",
				date:     "2024-01-01",
			},
			insertData: func(db *pgxpool.Pool) {
				r := NewReceiptRepo(db)

				r.InsertRawReceipt(ctx, time.Date(2024, 01, 01, 0, 0, 0, 0, time.UTC), "receipt", "norfa", model.ReceiptProducts{
					{
						Name: "red apples",
					},
					{
						Name: "lentils",
					},
				})
			},
			want: []model.PurchasedProductNew{
				{
					Name: "red apples",
				},
				{
					Name: "lentils",
				},
			},
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

			r := NewReceiptRepo(db)

			got, err := r.GetUnconfirmedReceipt(tt.args.ctx, tt.args.retailer, tt.args.date)
			if tt.wantErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(tt.want, got)
		})
	}
}
