package collector

import (
	"context"
	"math"
	"strconv"

	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

type ValidatorCommissionGauge struct {
	chainID          string
	denom            string
	desc             *prometheus.Desc
	exponent         uint
	grpcConn         *grpc.ClientConn
	validatorAddress string
}

func NewValidatorCommissionGauge(grpcConn *grpc.ClientConn, validatorAddress string, chainID string, denom string, exponent uint) *ValidatorCommissionGauge {
	return &ValidatorCommissionGauge{
		chainID: chainID,
		denom:   denom,
		desc: prometheus.NewDesc(
			"validator_commission",
			"Commission of the validator",
			[]string{"validator_address", "chain_id", "denom"},
			nil,
		),
		exponent:         exponent,
		grpcConn:         grpcConn,
		validatorAddress: validatorAddress,
	}
}

func (collector *ValidatorCommissionGauge) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.desc
}

func (collector *ValidatorCommissionGauge) Collect(ch chan<- prometheus.Metric) {
	distributionClient := distributiontypes.NewQueryClient(collector.grpcConn)
	distributionRes, err := distributionClient.ValidatorCommission(
		context.Background(),
		&distributiontypes.QueryValidatorCommissionRequest{ValidatorAddress: collector.validatorAddress},
	)
	if err != nil {
		ch <- prometheus.NewInvalidMetric(collector.desc, err)
		return
	}

	for _, commission := range distributionRes.Commission.Commission {
		if value, err := strconv.ParseFloat(commission.Amount.String(), 64); err != nil {
			// TODO LOGGING
		} else {
			displayValue := value / math.Pow10(int(collector.exponent))
			ch <- prometheus.MustNewConstMetric(collector.desc, prometheus.GaugeValue, displayValue, collector.validatorAddress, collector.chainID, collector.denom)
		}
	}

}
