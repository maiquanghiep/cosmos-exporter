package collector

import (
	"context"
	"math"
	"strconv"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

type ValidatorStatus struct {
	chainID          string
	denom            string
	exponent         uint
	jailDesc         *prometheus.Desc
	rateDesc         *prometheus.Desc
	votingDesc       *prometheus.Desc
	grpcConn         *grpc.ClientConn
	validatorAddress string
}

func NewValidatorStatus(grpcConn *grpc.ClientConn, validatorAddress string, chainID string, denom string, exponent uint) *ValidatorStatus {
	return &ValidatorStatus{
		grpcConn:         grpcConn,
		validatorAddress: validatorAddress,
		chainID:          chainID,
		denom:            denom,
		exponent:         exponent,
		jailDesc: prometheus.NewDesc(
			"validator_jailed",
			"Return 1 if the validator is jailed",
			[]string{"validator_address", "chain_id"},
			nil,
		),
		rateDesc: prometheus.NewDesc(
			"validator_rate",
			"Commission rate of the validator",
			[]string{"validator_address", "chain_id"},
			nil,
		),
		votingDesc: prometheus.NewDesc(
			"validator_voting_power",
			"Voting power of the validator",
			[]string{"validator_address", "chain_id"},
			nil,
		),
	}
}

func (collector *ValidatorStatus) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.jailDesc
	ch <- collector.rateDesc
	ch <- collector.votingDesc
}

func (collector *ValidatorStatus) Collect(ch chan<- prometheus.Metric) {
	stakingClient := stakingtypes.NewQueryClient(collector.grpcConn)
	validator, err := stakingClient.Validator(
		context.Background(),
		&stakingtypes.QueryValidatorRequest{ValidatorAddr: collector.validatorAddress},
	)
	if err != nil {
		ch <- prometheus.NewInvalidMetric(collector.jailDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.rateDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.votingDesc, err)
		return
	}

	// Jail handle
	var jailed float64
	if validator.Validator.Jailed {
		jailed = 1
	} else {
		jailed = 0
	}
	ch <- prometheus.MustNewConstMetric(collector.jailDesc, prometheus.GaugeValue, jailed, collector.validatorAddress, collector.chainID)

	// Rate handle

	if rate, err := strconv.ParseFloat(validator.Validator.Commission.CommissionRates.Rate.String(), 64); err != nil {
		// LOG
	} else {
		ch <- prometheus.MustNewConstMetric(collector.rateDesc, prometheus.GaugeValue, rate, collector.validatorAddress, collector.chainID)
	}

	// Share handle

	if value, err := strconv.ParseFloat(validator.Validator.DelegatorShares.String(), 64); err != nil {
		// LOG
	} else {
		displayValue := value / math.Pow10(int(collector.exponent))
		ch <- prometheus.MustNewConstMetric(collector.votingDesc, prometheus.GaugeValue, displayValue, collector.validatorAddress, collector.chainID)
	}

}
