package models

import (
	"context"

	"github.com/txpull/unpack/db"
	"github.com/txpull/unpack/types"
)

func CreateEventMapperTable(ctx context.Context, client *db.ClickHouse) error {
	query := `
		CREATE TABLE IF NOT EXISTS events_mapper (
			uuid UUID,
			contract_uuid UUID,
			event_uuid UUID,
			timestamp DateTime DEFAULT now()
		) engine=MergeTree() order by (uuid, contract_uuid, event_uuid, timestamp)
	`

	if err := client.DB().Exec(ctx, query); err != nil {
		return err
	}

	return nil
}

func InsertEventMapping(ctx context.Context, client *db.ClickHouse, mapping *types.EventMapping) error {
	query := `
		INSERT INTO events_mapper (
			uuid,
			contract_uuid,
			event_uuid
		) VALUES (?, ?, ?)
	`

	err := client.DB().Exec(ctx, query,
		mapping.UUID.String(),
		mapping.ContractUUID.String(),
		mapping.EventUUID.String(),
	)
	if err != nil {
		return err
	}

	return nil
}
