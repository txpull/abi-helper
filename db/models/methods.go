package models

import (
	"context"

	"github.com/google/uuid"
	"github.com/txpull/unpack/db"
	"github.com/txpull/unpack/types"
)

func CreateMethodsTable(ctx context.Context, client *db.ClickHouse) error {
	// failure to create methods table: code: 44, message: Cannot create table with column 'arguments'
	// which type is 'Object('json')' because experimental Object type is not allowed.
	// Set allow_experimental_object_type = 1 in order to allow it
	//
	query := `
		CREATE TABLE IF NOT EXISTS methods (
			uuid UUID,
			name String,
			raw_name String,
			signature String,
			hex String,
			bytes Array(UInt8),
			is_constant bool,
			is_payable bool,
			is_partial bool,
			arguments Nullable(String),
			returns Nullable(String),
			state_mutability Nullable(String),
			type Nullable(String),
			timestamp DateTime DEFAULT now()
		) engine=MergeTree() order by (uuid, hex, timestamp)
	`

	if err := client.DB().Exec(ctx, query); err != nil {
		return err
	}

	return nil
}

func InsertMethod(ctx context.Context, client *db.ClickHouse, method *types.Method) error {
	query := `
		INSERT INTO methods (
			uuid,
			name,
			raw_name,
			signature,
			hex,
			bytes,
			is_constant,
			is_payable,
			is_partial,
			arguments,
			returns,
			state_mutability,
			type
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	err := client.DB().Exec(ctx, query,
		uuid.New().String(),
		method.Name,
		method.RawName,
		method.Signature,
		method.Hex,
		method.Bytes,
		method.IsConstant,
		method.IsPayable,
		method.IsPartial,
		method.GetArgumentsAsJSON(),
		method.GetReturnsAsJSON(),
		method.StateMutability,
		method.Type,
	)
	if err != nil {
		return err
	}

	return nil
}

func GetMethod(ctx context.Context, client *db.ClickHouse, signature string) (*types.Method, error) {
	query := `
		SELECT
			uuid,
			name,
			raw_name,
			signature,
			hex,
			bytes,
			is_constant,
			is_payable,
			is_partial,
			arguments,
			returns,
			state_mutability,
			type
		FROM methods
		WHERE signature = ?
	`

	var method types.Method
	if err := client.DB().QueryRow(ctx, query, signature).Scan(
		&method.UUID,
		&method.Name,
		&method.RawName,
		&method.Signature,
		&method.Hex,
		&method.Bytes,
		&method.IsConstant,
		&method.IsPayable,
		&method.IsPartial,
		&method.Arguments,
		&method.Returns,
		&method.StateMutability,
		&method.Type,
	); err != nil {
		return nil, err
	}

	return &method, nil
}

func MethodExists(ctx context.Context, client *db.ClickHouse, method *types.Method) (bool, error) {
	query := `SELECT COUNT(*) FROM methods WHERE hex = ?`

	var count uint64
	if err := client.DB().QueryRow(ctx, query, method.Hex).Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

func DeleteMethodById(ctx context.Context, client *db.ClickHouse, id *uuid.UUID) error {
	query := `DELETE FROM methods WHERE uuid = ?`

	if err := client.DB().Exec(ctx, query, id.String()); err != nil {
		return err
	}

	return nil
}
