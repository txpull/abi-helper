package bscscan

import (
	"context"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/txpull/unpack/db"
	"github.com/txpull/unpack/scanners"
	"go.uber.org/zap"
)

// VerifiedContractsReader encapsulates the context, the BadgerDB instance,
// and a map of contracts data.
type VerifiedContractsReader struct {
	ctx      context.Context
	badgerDb *db.BadgerDB
	data     map[common.Address]*scanners.Result
}

// NewVerifiedContractsReader creates and returns a new VerifiedContractsReader instance.
//
// Example:
//
//	reader := bscscan.NewVerifiedContractsReader(ctx, badgerDb)
func NewVerifiedContractsReader(ctx context.Context, badgerDb *db.BadgerDB) *VerifiedContractsReader {
	return &VerifiedContractsReader{
		ctx:      ctx,
		badgerDb: badgerDb,
		data:     make(map[common.Address]*scanners.Result),
	}
}

// Read loads data from BadgerDB into the data map of the reader.
// Each item in the map represents a verified contract.
//
// Example:
//
//	err := reader.Discover()
//	if err != nil {
//	    log.Fatal(err)
//	}
func (r *VerifiedContractsReader) Discover() error {
	return r.badgerDb.DB().View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(databaseKeyPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			keyAddr := common.HexToAddress(string(k[len(databaseKeyPrefix):]))

			err := item.Value(func(v []byte) error {
				contract := &scanners.Result{}

				if err := contract.UnmarshalBytes(v); err != nil {
					zap.L().Error(
						"failed to unmarshal contract data",
						zap.String("contract_address", keyAddr.Hex()),
						zap.Error(err),
					)
					return err
				}

				// Add contract to the data map, using contract address as the key
				r.data[keyAddr] = contract
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// GetContractByAddress retrieves a contract by its address from the data map.
//
// Example:
//
//	contract, ok := reader.GetContractByAddress("address")
//	if !ok {
//	    fmt.Println("Contract not found")
//	}
func (r *VerifiedContractsReader) GetContractByAddress(address common.Address) (*scanners.Result, bool) {
	key := databaseKeyPrefix + address.Hex()

	exists, err := r.badgerDb.Exists(key)
	if err != nil {
		zap.L().Error(
			ErrFailedCheckExistenceInBadger.Error(),
			zap.String("contract_address", address.Hex()),
			zap.Error(err),
		)
		return nil, false
	} else if !exists {
		return nil, false
	}

	resBytes, err := r.badgerDb.Get(key)
	if err != nil {
		zap.L().Error("failed to get contract data from BadgerDB",
			zap.String("contract_address", address.Hex()),
			zap.Error(err),
		)
		return nil, false
	}

	contract := &scanners.Result{}

	if err := contract.UnmarshalBytes(resBytes); err != nil {
		zap.L().Error(
			"failed to unmarshal contract data",
			zap.String("contract_address", address.Hex()),
			zap.Error(err),
		)
		return nil, false
	}

	return contract, true
}

// GetContracts retrieves all contracts from the data map.
//
// Example:
//
//	contracts := reader.GetContracts()
//	for address, contract := range contracts {
//	    fmt.Println("Address:", address)
//	    fmt.Println("Contract Name:", contract.Name)
//	    fmt.Println()
//	}
func (r *VerifiedContractsReader) GetContracts() map[common.Address]*scanners.Result {
	return r.data
}
