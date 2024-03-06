package chainiterator

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"

	"github.com/taikoxyz/taiko-mono/packages/taiko-client/internal/testutils"
)

type BlockBatchIteratorTestSuite struct {
	testutils.ClientTestSuite
}

func (s *BlockBatchIteratorTestSuite) TestIter() {
	var maxBlocksReadPerEpoch uint64 = 2

	headHeight, err := s.RPCClient.L1.BlockNumber(context.Background())
	s.Nil(err)
	s.Greater(headHeight, uint64(0))

	lastEnd := common.Big0

	iter, err := NewBlockBatchIterator(context.Background(), &BlockBatchIteratorConfig{
		Client:                s.RPCClient.L1,
		MaxBlocksReadPerEpoch: &maxBlocksReadPerEpoch,
		StartHeight:           common.Big0,
		EndHeight:             new(big.Int).SetUint64(headHeight),
		OnBlocks: func(
			ctx context.Context,
			start, end *types.Header,
			updateCurrentFunc UpdateCurrentFunc,
			endIterFunc EndIterFunc,
		) error {
			s.Equal(lastEnd.Uint64(), start.Number.Uint64())
			lastEnd = end.Number
			return nil
		},
	})

	s.Nil(err)
	s.Nil(iter.Iter())
	s.Equal(headHeight, lastEnd.Uint64())
}

func (s *BlockBatchIteratorTestSuite) TestIterEndFunc() {
	var maxBlocksReadPerEpoch uint64 = 2

	headHeight, err := s.RPCClient.L1.BlockNumber(context.Background())
	s.Nil(err)
	s.Greater(headHeight, maxBlocksReadPerEpoch)

	lastEnd := common.Big0

	iter, err := NewBlockBatchIterator(context.Background(), &BlockBatchIteratorConfig{
		Client:                s.RPCClient.L1,
		MaxBlocksReadPerEpoch: &maxBlocksReadPerEpoch,
		StartHeight:           common.Big0,
		EndHeight:             new(big.Int).SetUint64(headHeight),
		OnBlocks: func(
			ctx context.Context,
			start, end *types.Header,
			updateCurrentFunc UpdateCurrentFunc,
			endIterFunc EndIterFunc,
		) error {
			s.Equal(lastEnd.Uint64(), start.Number.Uint64())
			lastEnd = end.Number
			endIterFunc()
			return nil
		},
	})

	s.Nil(err)
	s.Nil(iter.Iter())
	s.Equal(lastEnd.Uint64(), maxBlocksReadPerEpoch)
}

func (s *BlockBatchIteratorTestSuite) TestIterCtxCancel() {
	lastEnd := common.Big0
	headHeight, err := s.RPCClient.L1.BlockNumber(context.Background())
	s.Nil(err)
	ctx, cancel := context.WithCancel(context.Background())
	retry := 5 * time.Second

	itr, err := NewBlockBatchIterator(ctx, &BlockBatchIteratorConfig{
		Client:                s.RPCClient.L1,
		MaxBlocksReadPerEpoch: nil,
		RetryInterval:         &retry,
		StartHeight:           common.Big0,
		EndHeight:             new(big.Int).SetUint64(headHeight),
		OnBlocks: func(
			ctx context.Context,
			start, end *types.Header,
			updateCurrentFunc UpdateCurrentFunc,
			endIterFunc EndIterFunc,
		) error {
			s.Equal(lastEnd.Uint64(), start.Number.Uint64())
			lastEnd = end.Number
			endIterFunc()
			return nil
		},
	})

	s.Nil(err)
	cancel()
	// should output a log.Warn and context cancel error
	err8 := itr.Iter()
	s.ErrorContains(err8, "context canceled")
}

func (s *BlockBatchIteratorTestSuite) TestBlockBatchIteratorConfig() {
	_, err := NewBlockBatchIterator(context.Background(), &BlockBatchIteratorConfig{
		Client: nil,
	})
	s.ErrorContains(err, "invalid RPC client")

	_, err2 := NewBlockBatchIterator(context.Background(), &BlockBatchIteratorConfig{
		Client:   s.RPCClient.L1,
		OnBlocks: nil,
	})
	s.ErrorContains(err2, "invalid callback")

	lastEnd := common.Big0
	_, err3 := NewBlockBatchIterator(context.Background(), &BlockBatchIteratorConfig{
		Client: s.RPCClient.L1,
		OnBlocks: func(
			ctx context.Context,
			start, end *types.Header,
			updateCurrentFunc UpdateCurrentFunc,
			endIterFunc EndIterFunc,
		) error {
			s.Equal(lastEnd.Uint64(), start.Number.Uint64())
			lastEnd = end.Number
			endIterFunc()
			return nil
		},
		StartHeight: nil,
	})
	s.ErrorContains(err3, "invalid start height")

	_, err4 := NewBlockBatchIterator(context.Background(), &BlockBatchIteratorConfig{
		Client: s.RPCClient.L1,
		OnBlocks: func(
			ctx context.Context,
			start, end *types.Header,
			updateCurrentFunc UpdateCurrentFunc,
			endIterFunc EndIterFunc,
		) error {
			s.Equal(lastEnd.Uint64(), start.Number.Uint64())
			lastEnd = end.Number
			endIterFunc()
			return nil
		},
		StartHeight: common.Big2,
		EndHeight:   common.Big0,
	})
	s.ErrorContains(err4, "start height (2) > end height (0)")

	_, err6 := NewBlockBatchIterator(context.Background(), &BlockBatchIteratorConfig{
		Client: s.RPCClient.L1,
		OnBlocks: func(
			ctx context.Context,
			start, end *types.Header,
			updateCurrentFunc UpdateCurrentFunc,
			endIterFunc EndIterFunc,
		) error {
			s.Equal(lastEnd.Uint64(), start.Number.Uint64())
			lastEnd = end.Number
			endIterFunc()
			return nil
		},
		StartHeight: big.NewInt(1000), // use very high number
		EndHeight:   big.NewInt(1000),
	})
	s.ErrorContains(err6, "failed to get start header")

	_, err7 := NewBlockBatchIterator(context.Background(), &BlockBatchIteratorConfig{
		Client: s.RPCClient.L1,
		OnBlocks: func(
			ctx context.Context,
			start, end *types.Header,
			updateCurrentFunc UpdateCurrentFunc,
			endIterFunc EndIterFunc,
		) error {
			s.Equal(lastEnd.Uint64(), start.Number.Uint64())
			lastEnd = end.Number
			endIterFunc()
			return nil
		},
		StartHeight: common.Big0,
		EndHeight:   big.NewInt(1000), // use very high number
	})
	s.ErrorContains(err7, "failed to get end header")
}

func TestBlockBatchIteratorTestSuite(t *testing.T) {
	suite.Run(t, new(BlockBatchIteratorTestSuite))
}
