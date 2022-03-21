package collector

import (
	"context"
	"math"
	"strconv"

	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

type RewardGauge struct {
	chainID       string
	denom         string
	desc          *prometheus.Desc
	exponent      uint
	grpcConn      *grpc.ClientConn
	rewardAddress string
}

func NewRewardGauge(grpcConn *grpc.ClientConn, rewardAddress string, chainID string, denom string, exponent uint) *RewardGauge {
	return &RewardGauge{
		chainID: chainID,
		denom:   denom,
		desc: prometheus.NewDesc(
			"reward_balance",
			"Rewards of the address",
			[]string{"address", "validator_address", "chain_id", "denom"},
			nil,
		),
		exponent:      exponent,
		grpcConn:      grpcConn,
		rewardAddress: rewardAddress,
	}
}

func (collector *RewardGauge) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.desc
}

func (collector *RewardGauge) Collect(ch chan<- prometheus.Metric) {
	distributionClient := distributiontypes.NewQueryClient(collector.grpcConn)
	distributionRes, err := distributionClient.DelegationTotalRewards(
		context.Background(),
		&distributiontypes.QueryDelegationTotalRewardsRequest{DelegatorAddress: collector.rewardAddress},
	)
	if err != nil {
		ch <- prometheus.NewInvalidMetric(collector.desc, err)
		return
	}

	for _, reward := range distributionRes.Rewards {
		for _, entry := range reward.Reward {
			if value, err := strconv.ParseFloat(entry.Amount.String(), 64); err != nil {
				// TODO LOGGING
			} else {
				displayValue := value / math.Pow10(int(collector.exponent))
				ch <- prometheus.MustNewConstMetric(collector.desc, prometheus.GaugeValue, displayValue, collector.rewardAddress, reward.ValidatorAddress, collector.chainID, collector.denom)
			}
		}
	}

}
