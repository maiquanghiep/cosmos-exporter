package collector

import (
	"context"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

type ValidatorDelegationGauge struct {
	chainID          string
	desc             *prometheus.Desc
	grpcConn         *grpc.ClientConn
	validatorAddress string
}

func NewValidatorDelegationGauge(grpcConn *grpc.ClientConn, validatorAddress string, chainID string) *ValidatorDelegationGauge {
	return &ValidatorDelegationGauge{
		grpcConn:         grpcConn,
		validatorAddress: validatorAddress,
		chainID:          chainID,
		desc: prometheus.NewDesc(
			"validator_delegation_count",
			"Number of delegations to the validator",
			[]string{"validator_address", "chain_id"},
			nil,
		),
	}
}

func (collector *ValidatorDelegationGauge) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.desc
}

func (collector *ValidatorDelegationGauge) Collect(ch chan<- prometheus.Metric) {
	stakingClient := stakingtypes.NewQueryClient(collector.grpcConn)
	stakingRes, err := stakingClient.ValidatorDelegations(
		context.Background(),
		&stakingtypes.QueryValidatorDelegationsRequest{ValidatorAddr: collector.validatorAddress},
	)
	if err != nil {
		ch <- prometheus.NewInvalidMetric(collector.desc, err)
		return
	}

	delegationsCount := float64(len(stakingRes.DelegationResponses))

	ch <- prometheus.MustNewConstMetric(collector.desc, prometheus.GaugeValue, delegationsCount, collector.validatorAddress, collector.chainID)
}
