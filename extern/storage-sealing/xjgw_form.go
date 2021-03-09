package sealing

import (
	"fmt"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"golang.org/x/xerrors"
)

type IFormValid interface {
	Valid() error
}

type ThresholdBaseFeeForm struct {
	ThresholdBaseFee string `json:"thresholdBaseFee"`
	Option           string `json:"option"`
}

func (t ThresholdBaseFeeForm) Valid() (abi.TokenAmount, error) {
	bigBaseFee, err := big.FromString(t.ThresholdBaseFee)
	if err != nil {
		return abi.TokenAmount{}, xerrors.Errorf("baseFee parse big int error %v  value is %v ", err, t.ThresholdBaseFee)
	}
	if bigBaseFee.LessThanEqual(big.NewInt(0)) {
		return abi.TokenAmount{}, xerrors.Errorf("baseFee is less than zero,value is %v ", t.ThresholdBaseFee)
	}
	if t.Option != "ConfigBaseFee" {
		return abi.TokenAmount{}, xerrors.Errorf("config baseFee: option is no correct current option is %v ", t.Option)
	}

	return bigBaseFee, nil

}

type AutoCommitSwitchForm struct {
	AutoSwitch int64  `json:"autoSwitch"`
	Option     string `json:"option"`
}

func (a AutoCommitSwitchForm) Valid() error {
	if a.AutoSwitch != AutoSubmitSwitchOpen && a.AutoSwitch != AutoSubmitSwitchClose {
		return fmt.Errorf("autoSwitch value is zero or less than zero")
	}
	if a.Option != "ConfigAutoSwitch" {
		return xerrors.Errorf("config baseFee option is no correct current option is %v ", a.Option)
	}
	return nil
}

type AutoSubmitConfig struct {
	SwitchStatus     bool   `json:"switchStatus"`
	ThresholdBaseFee string `json:"thresholdBaseFee"`
}
