package sealing

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-state-types/abi"
	"testing"
)

func TestGwApi(t *testing.T) {
	api, closer, err := GetFullNodeAPI()
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	defer closer()
	head, err := api.ChainHead(context.Background())
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	fmt.Print(head.MinTicketBlock().ParentBaseFee)
}

func TestBaseApi(t *testing.T) {
	a := abi.NewTokenAmount(100)
	b := abi.NewTokenAmount(99)
	lessThan := b.LessThan(a)
	fmt.Print(lessThan, a, b)
}
