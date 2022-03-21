package collector

import (
	"context"
	"math"

	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

type ValidatorsStatus struct {
	chainID         string
	denom           string
	exponent        uint
	totalVotingDesc *prometheus.Desc
	grpcConn        *grpc.ClientConn
}

func NewValidatorsStatus(grpcConn *grpc.ClientConn, chainID string, denom string, exponent uint) *ValidatorsStatus {
	return &ValidatorsStatus{
		grpcConn: grpcConn,
		chainID:  chainID,
		denom:    denom,
		exponent: exponent,
		totalVotingDesc: prometheus.NewDesc(
			"total_voting_power",
			"Total voting power of validators",
			[]string{"chain_id"},
			nil,
		),
	}
}

func (collector *ValidatorsStatus) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.totalVotingDesc
}

func (collector *ValidatorsStatus) Collect(ch chan<- prometheus.Metric) {
	stakingClient := stakingtypes.NewQueryClient(collector.grpcConn)
	validatorsResponse, err := stakingClient.Validators(
		context.Background(),
		&stakingtypes.QueryValidatorsRequest{
			Pagination: &querytypes.PageRequest{
				Limit: 1000,
			},
		},
	)
	if err != nil {
		ch <- prometheus.NewInvalidMetric(collector.totalVotingDesc, err)
		return
	}

	// Total voting power handle
	var totalVotingPower float64
	validators := validatorsResponse.Validators
	for _, validator := range validators {
		totalVotingPower += float64(validator.DelegatorShares.RoundInt64())
	}

	displayValue := totalVotingPower / math.Pow10(int(collector.exponent))
	ch <- prometheus.MustNewConstMetric(collector.totalVotingDesc, prometheus.GaugeValue, displayValue, collector.chainID)
}
