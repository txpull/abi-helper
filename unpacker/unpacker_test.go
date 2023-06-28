package unpacker

import (
	"context"
	"math/big"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/txpull/unpack/clients"
	"github.com/txpull/unpack/db"
	"github.com/txpull/unpack/options"
	"github.com/txpull/unpack/readers"
	"github.com/txpull/unpack/scanners"
)

func TestGenericUnpacker(t *testing.T) {
	//t.Skip("Requires series of credentials in order to run properly.")
	tAssert := assert.New(t)

	_, err := options.NewFromEnv("../.unpack.toml")
	tAssert.NoError(err, "failure to initialize options")

	ctx := context.TODO()

	// Redis is used to store cached unpacked contracts.
	rdb, err := clients.NewRedis(ctx, options.G().Database.Redis)
	tAssert.NoError(err, "failure to initialize redis client")
	tAssert.NotNil(rdb, "redis client is nil")

	// ClickHouse database is used to store unpacked contracts.
	cdb, err := db.NewClickHouse(ctx, options.G().Database.Clickhouse)
	tAssert.NoError(err, "failure to initialize clickhouse client")
	tAssert.NotNil(cdb, "clickhouse client is nil")

	// In order to be able nicely iterate through multiple datasets, we are going to use Manger.
	redisReader, err := readers.NewRedisReader(ctx, rdb)
	tAssert.NoError(err, "failure to initialize redis reader")
	tAssert.NotNil(redisReader, "redis reader is nil")

	clickhouseReader, err := readers.NewClickHouseReader(ctx, cdb)
	tAssert.NoError(err, "failure to initialize clickhouse reader")
	tAssert.NotNil(clickhouseReader, "clickhouse reader is nil")

	manager, err := readers.NewManager(ctx,
		readers.WithReader("redis", redisReader),
		readers.WithReader("clickhouse", clickhouseReader),
	)
	tAssert.NoError(err, "failure to initialize readers manager")
	tAssert.NotNil(manager, "readers manager is nil")

	// Set the priority reader at this point to be Redis.
	err = manager.SetPriorityReader("clickhouse")
	tAssert.NoError(err, "failure to set priority reader")

	// We are going to use actual real network to fetch necessary information to complete test.
	// You can find free rpc urls at:
	// BSC - https://chainlist.org/chain/56
	// ETH - https://chainlist.org/chain/1
	client, err := clients.NewEthClient(ctx, options.G().Networks.Binance.ArchiveNode)
	tAssert.NoError(err)
	tAssert.NotNil(client)

	// BitQuery is used to fetch contract block and transaction information if necessary.
	bitquery := scanners.NewBitQueryProvider(
		os.Getenv("BITQUERY_API_URL"),
		os.Getenv("BITQUERY_API_KEY"),
	)

	bscscan := scanners.NewBscScanProvider(ctx, options.G().Clients.Bscscan)

	unpacker, err := NewUnpacker(
		ctx,
		WithReaderManager(manager),
		WithEthClient(client),
		WithBitQuery(bitquery),
		WithBscScan(bscscan),
	)
	tAssert.NoError(err, "failure to initialize unpacker")
	tAssert.NotNil(unpacker, "unpacker is nil")

	// TESTS START HERE UNTIL NOW WE WERE JUST SETTING UP THE ENVIRONMENT

	// First test will look for contract that we are sure it exists in both redis and clickhouse.
	contract, complete, err := unpacker.UnpackContract(big.NewInt(56), common.HexToAddress("0x33fDd11397Bf41CceA71572db4C2AE2F276f84EE"), nil)
	tAssert.NoError(err, "failure to unpack contract")
	tAssert.NotNil(contract, "contract is nil")
	tAssert.True(complete, "contract is not complete")

	// Second test will look for contract that we are sure it does not exist anywhere but we know we can find information for it
	// on the blockchain.
	contractNew, completeNew, err := unpacker.UnpackContract(big.NewInt(56), common.HexToAddress("0xDFf99132777B82DFf71059FeF413689d340e1e4A"), nil)
	tAssert.NoError(err, "failure to unpack contract")
	tAssert.NotNil(contractNew, "contract is nil")
	tAssert.True(completeNew, "contract is not complete")

	spew.Dump(contractNew)
}
