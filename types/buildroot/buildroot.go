package buildroot

import (
	"github.com/apex-fusion/nexus/helper/keccak"
	itrie "github.com/apex-fusion/nexus/state/immutable-trie"
	"github.com/apex-fusion/nexus/types"
	"github.com/umbracle/fastrlp"
)

var arenaPool fastrlp.ArenaPool

// CalculateUncleRoot calculates the root of a list of uncles
func CalculateUncleRoot(uncles []*types.Header) types.Hash {
	if len(uncles) == 0 {
		return types.EmptyUncleHash
	}

	a := arenaPool.Get()
	v := a.NewArray()

	for _, i := range uncles {
		v.Set(i.MarshalRLPWith(a))
	}

	root := keccak.Keccak256Rlp(nil, v)

	arenaPool.Put(a)

	return types.BytesToHash(root)
}

func calculateRootWithRlp(num int, h func(indx int) *fastrlp.Value) types.Hash {
	hF := func(indx int) []byte {
		return h(indx).MarshalTo(nil)
	}

	return CalculateRoot(num, hF)
}

// CalculateRoot calculates a root with a callback
func CalculateRoot(num int, h func(indx int) []byte) types.Hash {
	if num == 0 {
		return types.EmptyRootHash
	}

	if num <= 128 {
		fastH := acquireFastHasher()
		dst, ok := fastH.Hash(num, h)

		// important to copy the return before releasing the hasher
		res := types.BytesToHash(dst)

		releaseFastHasher(fastH)

		if ok {
			return res
		}
	}

	// fallback to slow hash
	return types.BytesToHash(deriveSlow(num, h))
}

var numArenaPool fastrlp.ArenaPool

func deriveSlow(num int, h func(indx int) []byte) []byte {
	t := itrie.NewTrie()
	txn := t.Txn(nil)

	ar := numArenaPool.Get()
	for i := 0; i < num; i++ {
		indx := ar.NewUint(uint64(i))
		txn.Insert(indx.MarshalTo(nil), h(i))
		ar.Reset()
	}

	numArenaPool.Put(ar)

	x, _ := txn.Hash()

	return x
}
