package fourbyte

import "fmt"

var (
	// ErrFailedToInsertMethod is returned when we failed to insert method into ClickHouse.
	ErrFailedToInsertMethod = fmt.Errorf("failed to insert method into ClickHouse")

	// ErrFailedMarshalMethod is returned when we failed to marshal method info.
	ErrFailedMarshalMethod = fmt.Errorf("failed to marshal method")

	// ErrFailedRedisWrite is returned when we failed to write method info into Redis.
	ErrFailedRedisWrite = fmt.Errorf("failed to write method into Redis")

	// ErrFailedToCheckIfMethodCacheKeyExists is returned when we failed to check if method cache key exists.
	ErrFailedToCheckIfMethodCacheKeyExists = fmt.Errorf("failed to check if method cache key exists")

	// ErrFailedToCheckIfMethodExists is returned when we failed to check if method exists.
	ErrFailedToCheckIfMethodExists = fmt.Errorf("failed to check if method exists")

	// ErrFailedToGetLastPageNumber is returned when we failed to get last page number.
	ErrFailedToGetLastPageNumber = fmt.Errorf("failed to get last page number")

	// ErrFailedToGetPage is returned when we failed to get page.
	ErrFailedToGetPage = fmt.Errorf("failed to get page")

	// ErrFailedToConstructNewMethod is returned when we failed to construct new method.
	ErrFailedToConstructNewMethod = fmt.Errorf("failed to construct new method")

	// ErrFailedToExtractPageNum is returned when we failed to extract page number from URL.
	ErrFailedToExtractPageNum = fmt.Errorf("failed to extract page number from URL")

	// ErrFailedToSetNextPageNumber is returned when we failed to set next page number.
	ErrFailedToSetNextPageNumber = fmt.Errorf("failed to set next page number")
)
