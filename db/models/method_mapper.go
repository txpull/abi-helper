package models

import (
	"context"

	"github.com/txpull/unpack/db"
	"github.com/txpull/unpack/types"
)

func CreateMethodMapperTable(ctx context.Context, client *db.ClickHouse) error {
	query := `
		CREATE TABLE IF NOT EXISTS methods_mapper (
			uuid UUID,
			contract_uuid UUID,
			method_uuid UUID,
			timestamp DateTime DEFAULT now()
		) engine=MergeTree() order by (uuid, contract_uuid, method_uuid, timestamp)
	`

	if err := client.DB().Exec(ctx, query); err != nil {
		return err
	}

	return nil
}

func InsertMethodMapping(ctx context.Context, client *db.ClickHouse, mapping *types.MethodMapping) error {
	query := `
		INSERT INTO methods_mapper (
			uuid,
			contract_uuid,
			method_uuid
		) VALUES (?, ?, ?)
	`

	err := client.DB().Exec(ctx, query,
		mapping.UUID.String(),
		mapping.ContractUUID.String(),
		mapping.MethodUUID.String(),
	)
	if err != nil {
		return err
	}

	return nil
}
