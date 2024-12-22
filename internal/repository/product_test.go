package repository

import (
	"context"
	"testing"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/migrations"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type ProductTestSuite struct {
	suite.Suite
	Container *postgres.PostgresContainer
}

func (s *ProductTestSuite) SetupSuite() {
	ctx := context.Background()

	var err error
	s.Container, err = initPostgresContainer(ctx)
	s.Require().NoError(err)

	s.createDBSnapshotWithEmptyTables(ctx)
}

func (s *ProductTestSuite) TearDownSuite() {
	err := s.Container.Terminate(context.Background())
	s.Require().NoError(err)
}

func TestProductTestSuite(t *testing.T) {
	suite.Run(t, new(ProductTestSuite))
}

func initPostgresContainer(ctx context.Context) (*postgres.PostgresContainer, error) {
	dbName := "nutrients"
	dbUser := "user"
	dbPassword := "password"
	ctr, err := postgres.Run(
		ctx,
		"docker.io/postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)
	return ctr, err
}

func (s *ProductTestSuite) createDBSnapshotWithEmptyTables(ctx context.Context) {
	s.migrateDB(ctx)

	err := s.Container.Snapshot(ctx, postgres.WithSnapshotName("emptyTables"))
	s.Require().NoError(err)

}

func (s *ProductTestSuite) migrateDB(ctx context.Context) {
	dbPool, err := pgxpool.New(ctx, s.Container.MustConnectionString(ctx))
	s.Require().NoError(err)
	defer dbPool.Close()

	err = migrations.MigrateDB(ctx, dbPool)
	s.Require().NoError(err)
}

func (s *ProductTestSuite) TestProductRepo_InsertProducts() {
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

func (s *ProductTestSuite) TestProductRepo_GetProductGroups() {
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
						{"red apples", "norfa", "apples", "pieces", "1", 1, 0.90, 0.1, "", time.Date(2024, 10, 14, 0, 0, 0, 0, time.UTC)},
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
