package stream

import (
	"math/big"
	"testing"
	"time"

	"cosmossdk.io/log/v2"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestPendingTransactionCache(t *testing.T) {
	s := NewRPCStreams(nil, log.NewNopLogger(), nil, nil)
	tx := evmtypes.NewTxWithData(&ethtypes.LegacyTx{Nonce: 1, Gas: 21000, GasPrice: big.NewInt(1), Value: big.NewInt(0)})
	tx.From = common.HexToAddress("0x1234").Bytes()
	hash := tx.Hash()
	s.ListenPendingTx(tx)
	hashes, _ := s.PendingTxStream().ReadAllNonBlocking(0)
	require.Contains(t, hashes, hash)
	cached := s.PendingTransaction(hash)
	require.NotNil(t, cached)
	require.Equal(t, hash, cached.Hash())
	tx.From[0] = 1
	cached.From[1] = 2
	require.Equal(t, common.HexToAddress("0x1234").Bytes(), s.PendingTransaction(hash).From)
	entry, _ := s.pendingTxs.Get(hash)
	entry.received = time.Now().Add(-3 * time.Second)
	s.pendingTxs.Add(hash, entry)
	require.Nil(t, s.PendingTransaction(hash))
	require.Nil(t, s.PendingTransaction(common.Hash{}))
	for i := uint64(2); i < 4100; i++ {
		s.ListenPendingTx(evmtypes.NewTxWithData(&ethtypes.LegacyTx{Nonce: i, Gas: 21000, GasPrice: big.NewInt(1), Value: big.NewInt(0)}))
	}
	require.Equal(t, 4096, s.pendingTxs.Len())
	require.Nil(t, s.PendingTransaction(hash))
}
