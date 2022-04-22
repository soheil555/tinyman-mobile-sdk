package types

import algoTypes "github.com/algorand/go-algorand-sdk/types"

type LogicSig struct {
	Logic []byte
	Sig   algoTypes.Signature
	Msig  algoTypes.MultisigSig
}
