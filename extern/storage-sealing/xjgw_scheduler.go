package sealing

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/api"
	"golang.org/x/xerrors"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Manager struct {
	fullNodeApi      api.FullNode
	apiClose         jsonrpc.ClientCloser
	currentBaseFee   abi.TokenAmount
	thresholdBaseFee abi.TokenAmount
	heartTimer       *time.Ticker
	closeSign        chan struct{}
	switchOpen       bool            // 是否开启自动提交的功能， false 走原来的逻辑，自动提交
	sync.Mutex
}

var GwSchedulerManger = &Manager{}

func (m *Manager) Run() error {
	m.closeSign = make(chan struct{}, 1)

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

	// todo 环境变量获取
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

	// todo 时间可配置，
	m.heartTimer = time.NewTicker(25 * time.Second)

	time.Sleep(2 * time.Second)
	return nil
}

func (m *Manager) runBaseFee() abi.TokenAmount {
	for {
		select {
		case <-m.heartTimer.C:
			ctx, _ := context.WithTimeout(context.Background(), 25*time.Second)
			chainHead, err := m.fullNodeApi.ChainHead(ctx)
			if err != nil {
				log.Errorf("%v get chainHead error %v :", GwLogFilterFlag, err)
				// todo check api 错误
				continue
			}
			m.Lock()
			m.currentBaseFee = chainHead.MinTicketBlock().ParentBaseFee
			m.Unlock()
			log.Infof("%v get current BaseFee  %v ", GwLogFilterFlag, m.currentBaseFee)
		case <-m.closeSign:
			return abi.TokenAmount{}
		}
	}
}

func (m *Manager) runWebServer() error {

	http.HandleFunc(SetBaseFeeThresholdApiPath, m.ConfigThresholdBaseFeeHandler)

	http.HandleFunc(SetAutoSwitchStatusApiPath, m.ConfigThresholdBaseFeeHandler)

	// todo 默认端口环境变量获取
	err := http.ListenAndServe(":1100", nil)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) CanSubCommitted() bool {
	m.Lock()
	defer m.Unlock()
	if !m.switchOpen {
		return true
	}
	return m.currentBaseFee.LessThan(m.thresholdBaseFee)
}

func (m *Manager) Close() {
	close(m.closeSign)
	if m.apiClose != nil {
		m.apiClose()
	}
	m.heartTimer.Stop()

}

func (m *Manager) ConfigThresholdBaseFeeHandler(writer http.ResponseWriter, request *http.Request) {
	baseFee := request.FormValue("baseFee")
	option := request.FormValue("option")
	log.Infof("%v get base Config: set baseFee= %v ,option=%v ", GwLogFilterFlag, baseFee, option)
	if option == "ConfigBaseFee" {
		bigBaseFee, err := big.FromString(baseFee)
		if err != nil {
			log.Errorf("%v baseFee parse big int error %v, value is %v ", GwLogFilterFlag, err, baseFee)
		}
		// todo
		m.thresholdBaseFee = bigBaseFee
		log.Infof("%v config baseFee success,value is %v ", GwLogFilterFlag, m.thresholdBaseFee)
		_, err = writer.Write([]byte(fmt.Sprintf("config thresholdBaseFee is success, value is %v ", m.thresholdBaseFee)))
		if err != nil {
			log.Errorf("%v write http body is error %v ", GwLogFilterFlag, err)
		}
	} else {
		_, err := writer.Write([]byte(fmt.Sprintf("request param is error,option is %v ", option)))
		if err != nil {
			log.Errorf("%v write http body is error %v ", GwLogFilterFlag, err)
		}
	}
}

func (m *Manager) ConfigAutoCommitSwitchHandler(writer http.ResponseWriter, request *http.Request) {
	autoSwitchValue := request.FormValue("autoSwitch")
	option := request.FormValue("option")
	log.Infof("%v get autoSwitch Config: set autoSwitch= %v ,option=%v ", GwLogFilterFlag, autoSwitchValue, option)
	if option == "ConfigAutoSwitch" {
		autoSwitchFlag, err := strconv.ParseInt(autoSwitchValue, 10, 64)
		if err != nil {
			log.Errorf("%v baseFee parse big int error %v, value is %v ", GwLogFilterFlag, err, autoSwitchValue)
		}

		// todo
		m.switchOpen = autoSwitchFlag == 1
		log.Infof("%v config autoSwitch is success,value is %v ", GwLogFilterFlag, m.switchOpen)
		_, err = writer.Write([]byte(fmt.Sprintf("config autoSwitch is success, value is %v ", m.switchOpen)))
		if err != nil {
			log.Errorf("%v write http body is error %v ", GwLogFilterFlag, err)
		}
	} else {
		_, err := writer.Write([]byte(fmt.Sprintf("request param is error,option is %v ", option)))
		if err != nil {
			log.Errorf("%v write http body is error %v ", GwLogFilterFlag, err)
		}
	}
}
