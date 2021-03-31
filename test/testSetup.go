package test

import (
	"log"
	"net/http"
	"testing"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/factly/dega-api/config"
	"github.com/factly/dega-api/graph/generated"
	"github.com/factly/dega-api/graph/loaders"
	"github.com/factly/dega-api/graph/resolvers"
	"github.com/factly/dega-api/graph/validator"
	"github.com/factly/x/loggerx"
	"github.com/gavv/httpexpect/v2"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var path string = "/query"

type Query struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName interface{}            `json:"operationName"`
}

func CheckJSON(obj *httpexpect.Object, data interface{}, qry string) {
	obj.ContainsMap(map[string]interface{}{
		"data": map[string]interface{}{
			qry: data,
		},
	})
}

func TestRouter() http.Handler {

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(loggerx.Init())
	router.Use(validator.CheckSpace())
	router.Use(middleware.RealIP)
	// router.Use(util.GormRequestID)

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers.Resolver{}}))

	router.Handle("/query", loaders.DataloaderMiddleware(srv))

	return router
}

func SetupMockDB() sqlmock.Sqlmock {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Println(err)
	}

	dialector := postgres.New(postgres.Config{
		DSN:                  "sqlmock_db_0",
		DriverName:           "postgres",
		Conn:                 db,
		PreferSimpleProtocol: true,
	})

	config.DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Println(err)
	}

	return mock
}

// ExpectationsMet checks if all the expectations are fulfilled
func ExpectationsMet(t *testing.T, mock sqlmock.Sqlmock) {
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
