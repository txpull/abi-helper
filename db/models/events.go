package models

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/txpull/unpack/db"
	"github.com/txpull/unpack/types"
)

func CreateEventsTable(ctx context.Context, client *db.ClickHouse) error {
	// failure to create methods table: code: 44, message: Cannot create table with column 'arguments'
	// which type is 'Object('json')' because experimental Object type is not allowed.
	// Set allow_experimental_object_type = 1 in order to allow it
	query := `
		CREATE TABLE IF NOT EXISTS events (
			uuid UUID,
			name String,
			raw_name String,
			signature String,
			hash String,
			is_anonymous bool,
			is_partial bool,
			arguments Nullable(String),
			timestamp DateTime DEFAULT now()
		) engine=MergeTree() order by (uuid, hash, timestamp)
	`

	if err := client.DB().Exec(ctx, query); err != nil {
		return err
	}

	return nil
}

func InsertEvent(ctx context.Context, client *db.ClickHouse, method *types.Event) error {
	query := `
		INSERT INTO events (
			uuid,
			name,
			raw_name,
			signature,
			hash,
			is_anonymous,
			is_partial,
			arguments
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	err := client.DB().Exec(ctx, query,
		method.UUID.String(),
		method.Name,
		method.RawName,
		method.Signature,
		method.Hash.Hex(),
		method.IsAnonymous,
		method.IsPartial,
		method.GetArgumentsAsJSON(),
	)
	if err != nil {
		return err
	}

	return nil
}

func GetEvent(ctx context.Context, client *db.ClickHouse, chainId *big.Int, hash common.Hash) (*types.Event, error) {
	query := `
		SELECT
			uuid,
			name,
			raw_name,
			signature,
			hash,
			is_anonymous,
			is_partial,
			arguments
		FROM events
		WHERE hash = ?
		LIMIT 1
	`

	var method types.Event
	if err := client.DB().QueryRow(ctx, query, hash.Hex()).Scan(
		&method.UUID,
		&method.Name,
		&method.RawName,
		&method.Signature,
		&method.Hash,
		&method.IsAnonymous,
		&method.IsPartial,
		&method.Arguments,
	); err != nil {
		return nil, err
	}

	return &method, nil
}

func DeleteEventById(ctx context.Context, client *db.ClickHouse, id *uuid.UUID) error {
	query := `DELETE FROM events WHERE uuid = ?`

	if err := client.DB().Exec(ctx, query, id.String()); err != nil {
		return err
	}

	return nil
}

func EventExist(ctx context.Context, client *db.ClickHouse, hash common.Hash) (bool, error) {
	query := `SELECT COUNT(*) FROM events WHERE hash = ?`

	var count uint64
	if err := client.DB().QueryRow(ctx, query, hash.Hex()).Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}
