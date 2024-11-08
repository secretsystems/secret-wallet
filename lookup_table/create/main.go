package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"

	"github.com/deroproject/derohe/walletapi"
)

// go run ./lookup_table/create

func main() {
	walletapi.Initialize_LookupTable(1, 1<<21)

	fmt.Println("Creating lookup table...")
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	t := walletapi.Balance_lookup_table
	err := enc.Encode(t)
	if err != nil {
		log.Fatal(err)

	}

	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("./lookup_table/lookup_table", buffer.Bytes(), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}
