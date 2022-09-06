package databaseclient

import (
	"codepix/example-bank-api/lib/repositories"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
)

func MapError(tx *gorm.DB) error {
	if tx.Error == nil {
		return nil
	}
	switch err := tx.Error.(type) {
	case sqlite3.Error:
		return mapSqliteError(tx, err)
	case *pgconn.PgError:
		return mapPostgresError(tx, err)
	default:
		return mapGormError(tx, err)
	}
}

func mapGormError(tx *gorm.DB, err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		model := GetSchemaName(tx)
		return &repositories.NotFoundError{model}
	default:
		operation := getOperation(tx)
		return &repositories.InternalError{operation, tx.Statement.Table, err.Error()}
	}
}

func mapSqliteError(tx *gorm.DB, err sqlite3.Error) error {
	switch {
	case err.ExtendedCode == sqlite3.ErrConstraintUnique:
		model := GetSchemaName(tx)
		return &repositories.AlreadyExistsError{model}
	default:
		operation := getOperation(tx)
		return &repositories.InternalError{operation, tx.Statement.Table, err.Error()}
	}
}

func mapPostgresError(tx *gorm.DB, err *pgconn.PgError) error {
	switch {
	case err.Code == pgerrcode.UniqueViolation:
		model := GetSchemaName(tx)
		return &repositories.AlreadyExistsError{model}
	default:
		operation := getOperation(tx)
		return &repositories.InternalError{operation, tx.Statement.Table, err.Error()}
	}
}

func getOperation(tx *gorm.DB) string {
	clauses := []string{}
	for clause := range tx.Statement.Clauses {
		clauses = append(clauses, strings.ToLower(clause))
	}
	return fmt.Sprintf("[%s]", strings.Join(clauses, ", "))
}

func GetSchemaName(tx *gorm.DB) string {
	return strings.ToLower(splitCamelCase(tx.Statement.Schema.Name))
}

var splitCamelCaseRegex = regexp.MustCompile("([A-Z])([A-Z])([a-z])|([a-z])([A-Z])")

func splitCamelCase(s string) string {
	return splitCamelCaseRegex.ReplaceAllString(s, "$1$4 $2$3$5")
}
