package sealing

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/api"
	builtin0 "github.com/filecoin-project/specs-actors/actors/builtin"
	miner3 "github.com/filecoin-project/specs-actors/v3/actors/builtin/miner"
	"golang.org/x/xerrors"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var GwDebug = false

func sectorPreCommitExpectExpired(sector SectorInfo, epoch abi.ChainEpoch) bool {
	if GwDebug {
		log.Debugf("%v preCommitExpectExpired: currentEpoch： %v ,TicketEpock: %v ", GwLogFilterFlag, epoch, sector.TicketEpoch)
		return epoch-sector.TicketEpoch > 100+miner3.ChainFinality
	}
	// todo maybe more precise
	return epoch-sector.TicketEpoch > builtin0.EpochsInDay+miner3.ChainFinality-500
}

func sectorProveCommitExpectExpired(sector SectorInfo, epoch abi.ChainEpoch) bool {
	if GwDebug {
		log.Debugf("%v proveCommitExpectExpired: currentEpoch： %v ,PreCommitEpoch: %v ", GwLogFilterFlag, epoch, sector.PreCommitEpoch)
		return epoch-sector.PreCommitEpoch > 100+miner3.ChainFinality
	}
	// todo,about 5 hour
	return epoch-sector.PreCommitEpoch > builtin0.EpochsInDay-500

}

/**
PreCommit 提交处理
返回值
	第一个代表：当期gas费率满足要求可以提交
	第二个代表：扇区即将到期，必须进行处理
*/
func LoopPreCommitCheckGas(sector SectorInfo) (bool, bool) {
	count := 1
	for {
		if GwSchedulerManger.CanSubCommitted() {
			return true, false
		}
		expectExpired := sectorPreCommitExpectExpired(sector, GwSchedulerManger.currentEpoch)
		if expectExpired {
			return false, true
		}
		// other error, timeout
		if count > builtin0.EpochsInDay+100 {
			return false, false
		}
		time.Sleep(builtin0.EpochDurationSeconds * time.Second)
		count++
	}

}

/**
ProveCommit  提交处理
返回值
	第一个代表：当期gas费率满足要求可以提交
	第二个代表：扇区即将到期，必须进行处理
*/

func LoopProveCommitCheckGas(sector SectorInfo) (bool, bool) {
	count := 1
	for {
		if GwSchedulerManger.CanSubCommitted() {
			return true, false
		}
		if sectorProveCommitExpectExpired(sector, GwSchedulerManger.currentEpoch) {
			return false, true
		}
		if count > builtin0.EpochsInDay {
			return false, false
		}
		time.Sleep(builtin0.EpochDurationSeconds * time.Second)
		count++
	}
}

type GwSchedulerManager struct {
	fullNodeApi      api.FullNode
	apiClose         jsonrpc.ClientCloser
	currentBaseFee   abi.TokenAmount
	thresholdBaseFee abi.TokenAmount
	heartTimer       *time.Ticker
	closeSign        chan struct{}
	switchOpen       bool
	currentEpoch     abi.ChainEpoch
	baseFeeLock      sync.Mutex
}

var GwSchedulerManger = &GwSchedulerManager{}

func (m *GwSchedulerManager) Run() error {

	gwDebug := os.Getenv(GwDebugKey)
	if gwDebug == "1" {
		GwDebug = true
	}
	log.Warnf("%v gwdebug info,gwdebug value is %v ", GwLogFilterFlag, GwDebug)

	baseFee := os.Getenv(GwDefaultThresholdBaseFeeKey)
	baseFeeValue, err := big.FromString(baseFee)
	if err != nil {
		return xerrors.Errorf("%v GwDefaultThresholdBaseFeeKey must config, please set value %v ", GwLogFilterFlag, err)
	}
	m.thresholdBaseFee = baseFeeValue

	m.closeSign = make(chan struct{}, 1)
	// default switch open
	m.switchOpen = true

	fullNodeAPI, closer, err := GetFullNodeAPI()
	if err != nil {
		return xerrors.Errorf("%v: init fullNodeApi error %v ", GwLogFilterFlag, err)
	}
	m.fullNodeApi = fullNodeAPI

	ctx, _ := context.WithTimeout(context.Background(), 25*time.Second)
	chainHead, err := fullNodeAPI.ChainHead(ctx)
	if err != nil {
		return xerrors.Errorf("%v get chainHead error %v :", GwLogFilterFlag, err)
	}
	m.currentBaseFee = chainHead.MinTicketBlock().ParentBaseFee

	m.thresholdBaseFee = abi.NewTokenAmount(DefaultThresholdBaseFee)
	log.Infof("%v get DefaultThresholdBaseFee is %v ", GwLogFilterFlag, m.thresholdBaseFee.String())

	m.apiClose = closer
	go func() {
		err = m.runWebServer()
		if err != nil {
			panic(xerrors.Errorf("%v: init runWebServer error %v ", GwLogFilterFlag, err))
		}
	}()

	go m.runBaseFee()

	// todo can config
	m.heartTimer = time.NewTicker(builtin0.EpochDurationSeconds * time.Second)

	time.Sleep(2 * time.Second)
	return nil
}

func (m *GwSchedulerManager) runBaseFee() abi.TokenAmount {
	for {
		select {
		case <-m.heartTimer.C:
			ctx, _ := context.WithTimeout(context.Background(), 25*time.Second)
			chainHead, err := m.fullNodeApi.ChainHead(ctx)
			if err != nil {
				log.Errorf("%v get chainHead error %v : maybe fullNNodeApi is error ", GwLogFilterFlag, err)
				// todo check api 错误
				continue
			}
			m.baseFeeLock.Lock()
			m.currentBaseFee = chainHead.MinTicketBlock().ParentBaseFee
			m.currentEpoch = chainHead.Height()
			m.baseFeeLock.Unlock()
			log.Infof("%v get current BaseFee  %v,currentEpock %v ", GwLogFilterFlag, m.currentBaseFee, m.currentEpoch)
		case <-m.closeSign:
			return abi.TokenAmount{}
		}
	}
}

func (m *GwSchedulerManager) CanSubCommitted() bool {
	m.baseFeeLock.Lock()
	defer m.baseFeeLock.Unlock()
	if !m.switchOpen {
		return true
	}
	return m.currentBaseFee.LessThan(m.thresholdBaseFee)
}

func (m *GwSchedulerManager) Close() {
	close(m.closeSign)
	if m.apiClose != nil {
		m.apiClose()
	}
	m.heartTimer.Stop()

}

func (m *GwSchedulerManager) runWebServer() error {
	http.HandleFunc(SetBaseFeeThresholdApiPath, m.ConfigThresholdBaseFeeHandler)
	http.HandleFunc(SetAutoSwitchStatusApiPath, m.ConfigAutoCommitSwitchHandler)
	http.HandleFunc(GetAutoConfigInfo, m.GetAutoSubmitConfigHandler)
	schedulerPort := os.Getenv(GwSchedulerPortKey)
	if schedulerPort != "" {
		_, err := strconv.ParseInt(schedulerPort, 10, 64)
		if err != nil {
			return xerrors.Errorf("%v config scheduler server port is error,value is %v ,please correct config: %v ", GwLogFilterFlag, schedulerPort, err)
		}
	} else {
		schedulerPort = DefaultSchedulerPort
	}

	// only local host
	err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%v", schedulerPort), nil)
	if err != nil {
		return err
	}
	return nil
}

func (m *GwSchedulerManager) GetAutoSubmitConfigHandler(writer http.ResponseWriter, request *http.Request) {
	// todo check token
	autoSubmitConfig := AutoSubmitConfig{
		SwitchStatus:     m.switchOpen,
		ThresholdBaseFee: m.thresholdBaseFee.String(),
		CurrentBaseFee:   m.currentBaseFee.String(),
	}
	Response(writer, StatusOK, "success", autoSubmitConfig)
}

func (m *GwSchedulerManager) ConfigThresholdBaseFeeHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusForbidden)
		return
	}
	thresholdBaseFeeForm := &ThresholdBaseFeeForm{}
	err := MarshalJson(request, thresholdBaseFeeForm)
	if err != nil {
		Response(writer, StatusFail, fmt.Sprintf("thresholdBaseFeeForm is error please confirm data format%v ", err), nil)
		return
	}
	baseFee, err := thresholdBaseFeeForm.Valid()
	if err != nil {
		Response(writer, StatusFail, fmt.Sprintf("request param is error please config request data  %v ", err), nil)
		return
	}
	m.baseFeeLock.Lock()
	m.thresholdBaseFee = baseFee
	m.baseFeeLock.Unlock()
	log.Infof("%v config ThresholdBaseFee is success,value is %v ", GwLogFilterFlag, m.thresholdBaseFee)
	Response(writer, StatusOK, fmt.Sprintf("success  current thresholdBaseFee is %v ", m.thresholdBaseFee), nil)

}

func (m *GwSchedulerManager) ConfigAutoCommitSwitchHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusForbidden)
		return
	}
	autoCommitSwitchForm := &AutoCommitSwitchForm{}
	err := MarshalJson(request, autoCommitSwitchForm)
	if err != nil {
		Response(writer, StatusFail, fmt.Sprintf("autoCommitSwitchFrom is error pleae config request data format%v ", err), nil)
		return
	}
	err = autoCommitSwitchForm.Valid()
	if err != nil {
		Response(writer, StatusFail, fmt.Sprintf("request params is error %v please confirm request data ", err), nil)
		return
	}
	m.switchOpen = autoCommitSwitchForm.AutoSwitch == AutoSubmitSwitchOpen
	log.Infof("%v config AutosSwitch, current switch is %v ", GwLogFilterFlag, m.switchOpen)
	Response(writer, StatusOK, fmt.Sprintf("success current switch status is %v ", m.switchOpen), nil)

}
