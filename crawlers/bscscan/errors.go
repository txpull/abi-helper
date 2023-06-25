package bscscan

import "errors"

var (
	// ErrNoContractsToProcess is returned when there are no contracts to process.
	ErrNoContractsToProcess = errors.New("no contracts to process")

	// ErrMaxRateLimitReached is returned when the maximum rate limit is reached and retry is needed.
	ErrMaxRateLimitReached = errors.New("max rate limit reached, retrying...")

	// ErrFailedGetContractInfo is returned when failed to get contract information from BSCScan.
	ErrFailedGetContractInfo = errors.New("failed to get contract information from BSCScan")

	// ErrFailedMarshalContractInfo is returned when failed to marshal contract information to binary.
	ErrFailedMarshalContractInfo = errors.New("failed to marshal contract information to binary")

	// ErrFailedCheckExistenceInBadger is returned when failed to check existence of contract information in redis.
	ErrFailedCheckExistenceInBadger = errors.New("failed to check existence of contract information in redis")

	// ErrFailedWriteContractInfoToDB is returned when failed to write contract information to redis.
	ErrFailedWriteContractInfoToDB = errors.New("failed to write contract information to redis")

	// ErrExceededMaxRetryAttempts is returned when the maximum number of retry attempts is exceeded.
	ErrExceededMaxRetryAttempts = errors.New("exceeded max retry attempts")

	// ErrFailedGetTransactionByHash is returned when failed to get transaction by hash.
	ErrFailedGetTransactionByHash = errors.New("failed to get transaction by hash")

	// ErrFailedGetTransactionReceiptByHash is returned when failed to get transaction receipt by hash.
	ErrFailedGetTransactionReceiptByHash = errors.New("failed to get transaction receipt by hash")

	// ErrFailedProcessAbi is returned when failed to process ABI.
	ErrFailedProcessAbi = errors.New("failed to process ABI")

	// ErrFailedParseAbi is returned when failed to parse ABI.
	ErrFailedParseAbi = errors.New("failed to parse ABI")

	// ErrFailedProcessAbiMethods is returned when failed to process ABI methods.
	ErrFailedProcessAbiMethods = errors.New("failed to process ABI methods")

	// ErrFailedToInsertMethod is returned when failed to insert method.
	ErrFailedToInsertMethod = errors.New("failed to insert method")

	// ErrFailedToInsertMethodMapping is returned when failed to insert method mapping.
	ErrFailedToInsertMethodMapping = errors.New("failed to insert method mapping")

	// ErrFailedMarshalMethodInfo is returned when failed to marshal method information to binary.
	ErrFailedMarshalMethodInfo = errors.New("failed to marshal method information to binary")

	// ErrFailedMarshalMethodMappingInfo is returned when failed to marshal method mapping information to binary.
	ErrFailedMarshalMethodMappingInfo = errors.New("failed to marshal method mapping information to binary")

	// ErrFailedWriteMethodInfoToDB is returned when failed to write method information to redis.
	ErrFailedWriteMethodInfoToDB = errors.New("failed to write method information to redis")

	// ErrFailedWriteMethodMappingInfoToDB is returned when failed to write method mapping information to redis.
	ErrFailedWriteMethodMappingInfoToDB = errors.New("failed to write method mapping information to redis")

	// ErrFailedProcessAbiEvents is returned when failed to process ABI events.
	ErrFailedProcessAbiEvents = errors.New("failed to process ABI events")

	// ErrFailedCheckEventExistenceInBadger is returned when failed to check existence of event information in redis.
	ErrFailedCheckEventExistenceInBadger = errors.New("failed to check existence of event information in redis")

	// ErrFailedToMarshalEventInfo is returned when failed to marshal event information to binary.
	ErrFailedToMarshalEventInfo = errors.New("failed to marshal event information to binary")

	// ErrFailedToWriteEventInfoToredis is returned when failed to write event information to redis.
	ErrFailedToWriteEventInfoToRedis = errors.New("failed to write event information to redis")

	// ErrFailedToInsertEvent is returned when failed to insert event.
	ErrFailedToInsertEvent = errors.New("failed to insert event")

	// ErrFailedToInsertEventMapping is returned when failed to insert event mapping.
	ErrFailedToInsertEventMapping = errors.New("failed to insert event mapping")

	// ErrFailedToWriteEventMappingInfoToredis is returned when failed to write event mapping information to redis.
	ErrFailedToWriteEventMappingInfoToRedis = errors.New("failed to write event mapping information to redis")

	// ErrFailedToCheckIfMethodExists is returned when failed to check if method exists.
	ErrFailedToCheckIfMethodExists = errors.New("failed to check if method exists")

	// ErrFailedToCheckIfMethodCacheKeyExists is returned when failed to check if method cache key exists.
	ErrFailedToCheckIfMethodCacheKeyExists = errors.New("failed to check if method cache key exists")
)
