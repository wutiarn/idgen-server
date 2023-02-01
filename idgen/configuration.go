package idgen

import (
	"errors"
	"fmt"
	"math"
)

//goland:noinspection GoVetStructTag
type Config struct {
	InstanceId           uint64 `env:"NODE_ID" yaml:"instanceId" env-required`
	TimestampBits        uint8  `env-default:"35" yaml:"timestampBits"`
	DomainBits           uint8  `env-default:"8" yaml:"domainBits"`
	CounterBits          uint8  `env-default:"14" yaml:"counterBits"`
	InstanceIdBits       uint8  `env-default:"6" yaml:"instanceIdBits"`
	EpochStartSecond     uint64 `env-default:"1672531200" yaml:"epochStartSecond"`
	ReservedSecondsCount uint64 `env-default:"60" yaml:"reservedSecondsCount"`
	StartupSecondOffset  int64  `env-default:"0" yaml:"startupSecondOffset"`
}

type configWrapper struct {
	config          Config
	maxTimestamp    uint64
	maxInstanceId   uint64
	maxCounterValue uint64
	maxDomain       uint64
}

func newConfigWrapper(c Config) (configWrapper, error) {
	maxDomain := calculateMaxValue(c.DomainBits)
	if maxDomain > math.MaxInt {
		return configWrapper{}, errors.New(
			fmt.Sprintf("domain count (%d) must be less then MaxInt", maxDomain),
		)
	}
	return configWrapper{
		config:          c,
		maxTimestamp:    calculateMaxValue(c.TimestampBits),
		maxInstanceId:   calculateMaxValue(c.InstanceIdBits),
		maxCounterValue: calculateMaxValue(c.CounterBits),
		maxDomain:       maxDomain,
	}, nil
}

func calculateMaxValue(bits uint8) uint64 {
	return uint64(math.Pow(2, float64(bits)) - 1)
}
