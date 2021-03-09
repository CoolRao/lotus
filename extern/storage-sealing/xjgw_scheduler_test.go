package sealing

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/go-state-types/abi"
	"testing"
)



func TestForm(t *testing.T){
	autoCommitSwitchForm := AutoCommitSwitchForm{}
	err := json.Unmarshal([]byte("{\n    \"autoSwitch\":1,\n    \"option\":\"ConfigAutoSwitch\"\n}"), &autoCommitSwitchForm)
	if err!=nil{
		fmt.Println(err.Error())
		return
	}
	fmt.Println(autoCommitSwitchForm.Option)
}




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
