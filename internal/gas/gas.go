package gas

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/ethclient"
)

/* EIP1559标准：
交易花费的gas主要考虑两方面：
 .GasPrice
    GasPrice=BaseFee+PriorityFee
    BaseFee=(当前Block Base Fee)*2,PriorityFee=市场建议值 :BaseFee,Priority Fee 都需要根据实际公司业务来决定，可以配置
    在链上执行的时候，BaseFee会按照实际的算，并不一定会按照Base Fee*2 来收取，Priority Fee 就是给多少算多少，影响交易的优先级
 .GasLimit
    基于RPC建议的估算值基础之上上调10%-20%，也可根据具体配置不同的上涨幅度.
判断交易状态
  交易是在节点内存池Or链上。处于Pending状态，通过交易hash查找该笔交易，没有对应的block number，没有上链，
  需要重新发起一笔交易覆盖旧交易(nonce 跟上一笔一致)，Gas Price需要上调至少10%，建议直接提高 20%-30%。
  如果上链，状态为Failed，需要判断具体原因，如果是‘Out of gas’,GasLimit不够，需要重新发起交易，GasLimit上调
*/

type GsEstimator struct {
	client *ethclient.Client
}
type GsProfile struct {
	GasLimit  uint64
	GasFeeGap *big.Int
	GasTipCap *big.Int
}

func NewGsEstimator(client *ethclient.Client) *GsEstimator {
	return &GsEstimator{client: client}
}
func (estimator *GsEstimator) setClient(client *ethclient.Client) {
	estimator.client = client
}

func (estimator *GsEstimator) GetProfile(ctx context.Context, msg *ethereum.CallMsg) (*GsProfile, error) {
	gasLimit, err := estimator.getGasLimit(msg, 1.2)
	if err != nil {
		return nil, err
	}
	maxFee, tipGap, err := estimator.getPrice()
	if err != nil {
		return nil, err
	}
	profile := new(GsProfile)
	profile.GasLimit = gasLimit
	profile.GasFeeGap = maxFee
	profile.GasTipCap = tipGap
	return profile, nil
}
func (estimator *GsEstimator) getPrice() (*big.Int, *big.Int, error) {
	// base fee
	header, err := estimator.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, nil, err
	}
	baseFee := header.BaseFee
	bufferBaseFee := new(big.Int).Mul(baseFee, big.NewInt(2))
	tipCap, err := estimator.client.SuggestGasTipCap(context.Background())
	if err != nil {
		return nil, nil, err
	}
	maxFeePerGas := new(big.Int).Add(bufferBaseFee, tipCap)

	return maxFeePerGas, tipCap, nil
}
func (estimator *GsEstimator) getGasLimit(msg *ethereum.CallMsg, buffer float64) (uint64, error) {
	estimateGas, err := estimator.client.EstimateGas(context.Background(), *msg)
	if err != nil {
		//todo print out logs
		return 0, err
	}
	//buffer默认为1.2
	buf := buffer
	if buf == 0 {
		buf = 1.2
	}
	gasLimit := uint64(float64(estimateGas) * buf)
	return gasLimit, nil
}
