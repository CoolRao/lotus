package sealing

import (
	"github.com/filecoin-project/go-state-types/abi"
	builtin0 "github.com/filecoin-project/specs-actors/actors/builtin"
)

// 自定义日志过滤
const GwLogFilterFlag = "xjgw"

// 默认baseFee阈值
const DefaultThresholdBaseFee = 110

// commit 消息上链时间,代码里面暂未处理，根据经验得大概24h
const ProveCommitMsgMaxAge = abi.ChainEpoch(builtin0.EpochsInDay)

// 服务默认端口
const DefaultSchedulerPort = "9999"

// 调度默认端口环境变量 key
const GwSchedulerPortKey = "GwSchedulerServerPort"

// 接口地址
const BaseApiPath = "/api/v1/"
const SetBaseFeeThresholdApiPath = BaseApiPath + "setBaseFeeThreshold"
const SetAutoSwitchStatusApiPath = BaseApiPath + "SetAutoSwitchStatus"
