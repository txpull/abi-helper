package models

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/txpull/unpack/db"
	"github.com/txpull/unpack/helpers"
	"github.com/txpull/unpack/types"
)

func CreateContractsTable(ctx context.Context, client *db.ClickHouse) error {
	query := `
		CREATE TABLE IF NOT EXISTS contracts (
			uuid UUID,
			chain_id Int64,
			block_hash Nullable(String),
			transaction_hash Nullable(String),
			contract_address String,
			name String,
			language String,
			compiler_version String,
			optimization_used String,
			runs String,
			constructor_arguments String,
			evm_version String,
			library String,
			license_type String,
			proxy String DEFAULT 0,
			source_code String,
			constructor_abi String,
			abi String,
			metadata String,
			source_urls Array(String),
			verification_type Nullable(Int16),
			verification_status Nullable(String),
			process_status Int8 DEFAULT 0,
			timestamp DateTime DEFAULT now()
		) engine=MergeTree() order by (uuid, chain_id, contract_address, timestamp)
	`

	if err := client.DB().Exec(ctx, query); err != nil {
		return err
	}

	return nil
}

func InsertContract(ctx context.Context, client *db.ClickHouse, contract *types.Contract) error {
	query := `
		INSERT INTO contracts (
			uuid,
			chain_id,
			block_hash,
			transaction_hash,
			contract_address,
			name,
			language,
			compiler_version,
			optimization_used,
			runs,
			constructor_arguments,
			evm_version,
			library,
			license_type,
			proxy,
			source_code,
			constructor_abi,
			abi,
			metadata,
			source_urls,
			verification_type,
			verification_status,
			process_status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	err := client.DB().Exec(ctx, query,
		contract.UUID.String(),
		contract.ChainID.Int64(),
		contract.BlockHash.Hex(),
		contract.TransactionHash.Hex(),
		contract.Address.Hex(),
		contract.Name,
		contract.Language,
		contract.CompilerVersion,
		contract.OptimizationUsed,
		contract.Runs,
		contract.ConstructorArguments,
		contract.EVMVersion,
		contract.Library,
		contract.LicenseType,
		contract.Proxy,
		contract.SourceCode,
		contract.ConstructorABI,
		contract.ABI,
		contract.MetaData,
		contract.SourceUrls,
		contract.VerificationType,
		contract.VerificationStatus,
		helpers.CONTRACT_PROCESS_STATUS_PENDING,
	)
	if err != nil {
		return err
	}

	return nil
}

func GetContract(ctx context.Context, client *db.ClickHouse, chainId *big.Int, addr common.Address) (*types.Contract, error) {
	query := `
		SELECT 
			uuid,
			chain_id,
			block_hash,
			transaction_hash,
			contract_address,
			name,
			language,
			compiler_version,
			optimization_used,
			runs,
			constructor_arguments,
			evm_version,
			library,
			license_type,
			proxy,
			source_code,
			constructor_abi,
			abi,
			metadata,
			source_urls,
			verification_type,
			verification_status,
			process_status
		FROM contracts WHERE contract_address = ? AND chain_id = ?
	`

	row := client.DB().QueryRow(ctx, query, addr.Hex(), chainId.Int64())

	contract := &types.Contract{}

	var rawChainId int64
	var blockHash, transactionHash, address, language string
	var verificationType int16

	err := row.Scan(
		&contract.UUID,
		&rawChainId,
		&blockHash,
		&transactionHash,
		&address,
		&contract.Name,
		&language,
		&contract.CompilerVersion,
		&contract.OptimizationUsed,
		&contract.Runs,
		&contract.ConstructorArguments,
		&contract.EVMVersion,
		&contract.Library,
		&contract.LicenseType,
		&contract.Proxy,
		&contract.SourceCode,
		&contract.ConstructorABI,
		&contract.ABI,
		&contract.MetaData,
		&contract.SourceUrls,
		&verificationType,
		&contract.VerificationStatus,
		&contract.ProcessStatus,
	)

	contract.ChainID = big.NewInt(rawChainId)
	contract.BlockHash = common.HexToHash(blockHash)
	contract.TransactionHash = common.HexToHash(transactionHash)
	contract.Address = common.HexToAddress(address)
	contract.VerificationType = types.VerificationType(verificationType)

	if err != nil {
		return nil, err
	}

	return contract, nil
}

func ContractExists(ctx context.Context, client *db.ClickHouse, address common.Address) (bool, error) {
	query := `SELECT COUNT(*) FROM contracts WHERE contract_address = ?`

	var count uint64
	if err := client.DB().QueryRow(ctx, query, address.Hex()).Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

func DeleteContractById(ctx context.Context, client *db.ClickHouse, id *uuid.UUID) error {
	query := `DELETE FROM contracts WHERE uuid = ?`

	if err := client.DB().Exec(ctx, query, id.String()); err != nil {
		return err
	}

	return nil
}

func DeleteContractByAddress(ctx context.Context, client *db.ClickHouse, addr *common.Address) error {
	query := `DELETE FROM contracts WHERE contract_address = ?`

	if err := client.DB().Exec(ctx, query, addr.Hex()); err != nil {
		return err
	}

	return nil
}
