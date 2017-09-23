// build ignore

package coin

import (
	"errors"
	"fmt"
	"testing"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func _badFeeCalc(t *Transaction) (uint64, error) {
	return 0, errors.New("Bad")
}

func makeNewBlock(uxHash cipher.SHA256) (*Block, error) {
	body := BlockBody{
		Transactions: Transactions{Transaction{}},
	}

	prev := Block{
		Body: body,
		Head: BlockHeader{
			Version:  0x02,
			Time:     100,
			BkSeq:    0,
			Fee:      10,
			PrevHash: cipher.SHA256{},
			BodyHash: body.Hash(),
		}}
	return NewBlock(prev, 100+20, uxHash, Transactions{Transaction{}}, _feeCalc)
}

func addTransactionToBlock(t *testing.T, b *Block) Transaction {
	tx := makeTransaction(t)
	b.Body.Transactions = append(b.Body.Transactions, tx)
	return tx
}

func TestNewBlock(t *testing.T) {
	// TODO -- update this test for newBlock changes
	prev := Block{Head: BlockHeader{Version: 0x02, Time: 100, BkSeq: 98}}
	uxHash := randSHA256(t)
	txns := Transactions{Transaction{}}
	// invalid txn fees panics
	_, err := NewBlock(prev, 133, uxHash, txns, _badFeeCalc)
	require.EqualError(t, err, fmt.Sprintf("Invalid transaction fees: Bad"))

	// no txns panics
	_, err = NewBlock(prev, 133, uxHash, nil, _feeCalc)
	require.EqualError(t, err, "Refusing to create block with no transactions")

	_, err = NewBlock(prev, 133, uxHash, Transactions{}, _feeCalc)
	require.EqualError(t, err, "Refusing to create block with no transactions")

	// valid block is fine
	fee := uint64(121)
	currentTime := uint64(133)
	b, err := NewBlock(prev, currentTime, uxHash, txns, _makeFeeCalc(fee))
	require.NoError(t, err)
	assert.Equal(t, b.Body.Transactions, txns)
	assert.Equal(t, b.Head.Fee, fee*uint64(len(txns)))
	assert.Equal(t, b.Body, BlockBody{Transactions: txns})
	assert.Equal(t, b.Head.PrevHash, prev.HashHeader())
	assert.Equal(t, b.Head.Time, currentTime)
	assert.Equal(t, b.Head.BkSeq, prev.Head.BkSeq+1)
	assert.Equal(t, b.Head.UxHash, uxHash)
}

func TestBlockHashHeader(t *testing.T) {
	uxHash := randSHA256(t)
	b, err := makeNewBlock(uxHash)
	require.NoError(t, err)
	assert.Equal(t, b.HashHeader(), b.Head.Hash())
	assert.NotEqual(t, b.HashHeader(), cipher.SHA256{})
}

func TestBlockHashBody(t *testing.T) {
	uxHash := randSHA256(t)
	b, err := makeNewBlock(uxHash)
	require.NoError(t, err)
	assert.Equal(t, b.HashBody(), b.Body.Hash())
	hb := b.HashBody()
	hashes := b.Body.Transactions.Hashes()
	tx := addTransactionToBlock(t, b)
	assert.NotEqual(t, b.HashBody(), hb)
	hashes = append(hashes, tx.Hash())
	assert.Equal(t, b.HashBody(), cipher.Merkle(hashes))
	assert.Equal(t, b.HashBody(), b.Body.Hash())
}

func TestNewGenesisBlock(t *testing.T) {
	gb, err := NewGenesisBlock(genAddress, _genCoins, _genTime)
	require.NoError(t, err)

	require.Equal(t, cipher.SHA256{}, gb.Head.PrevHash)
	require.Equal(t, _genTime, gb.Head.Time)
	require.Equal(t, uint64(0), gb.Head.BkSeq)
	require.Equal(t, uint32(0), gb.Head.Version)
	require.Equal(t, uint64(0), gb.Head.Fee)
	require.Equal(t, cipher.SHA256{}, gb.Head.UxHash)

	require.Equal(t, 1, len(gb.Body.Transactions))
	tx := gb.Body.Transactions[0]
	require.Len(t, tx.In, 0)
	require.Len(t, tx.Sigs, 0)
	require.Len(t, tx.Out, 1)

	require.Equal(t, genAddress, tx.Out[0].Address)
	require.Equal(t, _genCoins, tx.Out[0].Coins)
	require.Equal(t, _genCoins, tx.Out[0].Hours)
}