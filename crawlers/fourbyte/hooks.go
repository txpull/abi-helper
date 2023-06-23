package fourbyte

import (
	"github.com/txpull/unpack/db"
	"github.com/txpull/unpack/signatures"
)

func PostWriteClickHouseHook(db *db.ClickHouse) PostWriteHook {
	return func(signature *signatures.Signature) error {
		panic("OK")
		return nil
	}
}
