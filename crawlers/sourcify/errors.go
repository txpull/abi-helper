package sourcify

import "errors"

var (
	// ErrFailedToCheckIfMethodCacheKeyExists is returned when we fail to check if a method cache key exists
	ErrFailedToCheckIfMethodCacheKeyExists = errors.New("failed to check if method cache key exists")

	// ErrFailedGetTransactionReceiptByHash is returned when we fail to get a transaction receipt by hash
	ErrFailedGetTransactionReceiptByHash = errors.New("failed to get transaction receipt by hash")

	// ErrFailedToWriteContractToDatabase is returned when we fail to write a contract to the database
	ErrFailedToWriteContractToDatabase = errors.New("failed to write contract to database")

	// ErrContractAlreadyExists is returned when we try to write a contract to the database that already exists
	ErrContractAlreadyExists = errors.New("contract already exists")

	// ErrFailedMarshalContract is returned when we fail to marshal a contract
	ErrFailedMarshalContract = errors.New("failed to marshal contract")

	// ErrFailedProcessAbi is returned when we fail to process an ABI
	ErrFailedProcessAbi = errors.New("failed to process ABI")

	// ErrFailedWriteContractToRedis is returned when we fail to write a contract to redis
	ErrFailedWriteContractToRedis = errors.New("failed to write contract to redis")

	// ErrFailedParseAbi is returned when we fail to parse an ABI
	ErrFailedParseAbi = errors.New("failed to parse ABI")

	// ErrFailedProcessAbiMethods is returned when we fail to process ABI methods
	ErrFailedProcessAbiMethods = errors.New("failed to process ABI methods")

	// ErrFailedProcessAbiEvents is returned when we fail to process ABI events
	ErrFailedProcessAbiEvents = errors.New("failed to process ABI events")

	// ErrFailedToReadRedis is returned when we fail to read from redis
	ErrFailedToReadRedis = errors.New("failed to read from redis")

	// ErrFailedToCheckIfMethodExists is returned when we fail to check if a method exists
	ErrFailedToCheckIfMethodExists = errors.New("failed to check if method exists")

	// ErrFailedToInsertMethod is returned when we fail to insert a method
	ErrFailedToInsertMethod = errors.New("failed to insert method")

	// ErrFailedToInsertMethodMapping is returned when we fail to insert a method mapping
	ErrFailedToInsertMethodMapping = errors.New("failed to insert method mapping")

	// ErrFailedMarshalMethod is returned when we fail to marshal a method
	ErrFailedMarshalMethod = errors.New("failed to marshal method")

	// ErrFailedMarshalMethodMapping is returned when we fail to marshal a method mapping info
	ErrFailedMarshalMethodMapping = errors.New("failed to marshal method mapping")

	// ErrFailedWriteMethodToRedis is returned when we fail to write a method to redis
	ErrFailedWriteMethodToRedis = errors.New("failed to write method to redis")

	// ErrFailedWriteMethodMappingToRedis is returned when we fail to write a method mapping to redis
	ErrFailedWriteMethodMappingToRedis = errors.New("failed to write method mapping to redis")

	// ErrFailedToMarshalConstructorAbi is returned when we fail to marshal a constructor ABI
	ErrFailedToMarshalConstructorAbi = errors.New("failed to marshal constructor ABI")

	// ErrFailedToScanContract is returned when we fail to scan a contract
	ErrFailedToScanContract = errors.New("failed to scan contract")

	// ErrFailedCheckEventExistenceInRedis is returned when we fail to check if an event exists in redis
	ErrFailedCheckEventExistenceInRedis = errors.New("failed to check event existence in redis")

	// ErrFailedToInsertEvent is returned when we fail to insert an event
	ErrFailedToInsertEvent = errors.New("failed to insert event")

	// ErrFailedToInsertEventMapping is returned when we fail to insert an event mapping
	ErrFailedToInsertEventMapping = errors.New("failed to insert event mapping")

	// ErrFailedToMarshalEvent is returned when we fail to marshal an event
	ErrFailedToMarshalEvent = errors.New("failed to marshal event")

	// ErrFailedToWriteEventToRedis is returned when we fail to write an event to redis
	ErrFailedToWriteEventToRedis = errors.New("failed to write event to redis")

	// ErrFailedToWriteEventMappingInfoToRedis is returned when we fail to write an event mapping info to redis
	ErrFailedToWriteEventMappingInfoToRedis = errors.New("failed to write event mapping info to redis")

	// ErrFailedContractValidationCheck is returned when we fail to validate a contract
	ErrFailedContractValidationCheck = errors.New("failed to validate contract")
)
