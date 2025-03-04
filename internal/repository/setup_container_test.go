package repository

import (
	"context"
	"testing"

	"github.com/SarunasBucius/nutri-price-server/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type ContainerTestSuite struct {
	suite.Suite
	Container *postgres.PostgresContainer
}

func (s *ContainerTestSuite) SetupSuite() {
	ctx := context.Background()

	var err error
	s.Container, err = initPostgresContainer(ctx)
	s.Require().NoError(err)

	s.createDBSnapshotWithEmptyTables(ctx)
}

func (s *ContainerTestSuite) TearDownSuite() {
	err := s.Container.Terminate(context.Background())
	s.Require().NoError(err)
}

func TestProductTestSuite(t *testing.T) {
	suite.Run(t, new(ContainerTestSuite))
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

func (s *ContainerTestSuite) createDBSnapshotWithEmptyTables(ctx context.Context) {
	s.migrateDB(ctx)

	err := s.Container.Snapshot(ctx, postgres.WithSnapshotName("emptyTables"))
	s.Require().NoError(err)

}

func (s *ContainerTestSuite) migrateDB(ctx context.Context) {
	dbPool, err := pgxpool.New(ctx, s.Container.MustConnectionString(ctx))
	s.Require().NoError(err)
	defer dbPool.Close()

	err = migrations.MigrateDB(ctx, dbPool)
	s.Require().NoError(err)
}
