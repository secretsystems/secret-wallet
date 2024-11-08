package lookup_table

import (
	"bytes"
	_ "embed"
	"encoding/gob"

	"github.com/deroproject/derohe/walletapi"
)

//go:embed lookup_table
var LOOKUP_TABLE []byte

func Load() error {

	var (
		lookupTable walletapi.LookupTable
		buffer      bytes.Buffer
		enc         = gob.NewEncoder(&buffer)
		err         = enc.Encode(lookupTable)
	)
	if err != nil {
		return err
	}
	walletapi.Balance_lookup_table = &lookupTable
	return nil
}
