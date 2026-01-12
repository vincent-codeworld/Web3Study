package transcation

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func Test01() {
	// 1. 连接到以太坊节点 (这里使用 Infura 的 Sepolia 测试网为例)
	client, err := ethclient.Dial("https://sepolia.infura.io/v3/YOUR_INFURA_PROJECT_ID")
	if err != nil {
		log.Fatal(err)
	}

	// 2. 加载私钥
	// 注意：实际开发中不要把私钥硬编码在代码里，应该从环境变量或安全存储读取
	privateKeyHex := "YOUR_PRIVATE_KEY_WITHOUT_0x"
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatal(err)
	}

	// 从私钥推导出公钥和发送方地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("无法转换公钥")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 3. 获取账户 Nonce (防止重放攻击)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	// 4. 设置交易参数
	value := big.NewInt(10000000000000000)                    // 0.01 ETH (单位是 wei)
	toAddress := common.HexToAddress("0xRecipientAddress...") // 接收方地址
	gasLimit := uint64(21000)                                 // 普通转账固定为 21000

	// 获取建议的小费 (Gas Tip Cap)
	gasTipCap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// 获取建议的基础费 (Base Fee) 并计算 Gas Fee Cap
	head, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	// MaxFeePerGas = (BaseFee * 2) + Tip
	gasFeeCap := new(big.Int).Add(
		new(big.Int).Mul(head.BaseFee, big.NewInt(2)),
		gasTipCap,
	)

	// 5. 创建 EIP-1559 交易对象
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   head.Number, // 注意：这里通常需要准确的 ChainID，下面签名时会用到
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &toAddress,
		Value:     value,
		Data:      nil, // 普通转账 Data 为空
	})

	// 获取准确的 ChainID 用于签名
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// 6. 签名交易
	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	// 7. 广播交易
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("交易已发送! Hash: %s\n", signedTx.Hash().Hex())
}

func Test02() {
	// 1. 连接节点
	client, err := ethclient.Dial("https://sepolia.infura.io/v3/YOUR_INFURA_ID")
	if err != nil {
		log.Fatal(err)
	}

	// 2. 准备私钥
	privateKey, err := crypto.HexToECDSA("YOUR_PRIVATE_KEY_WITHOUT_0x")
	if err != nil {
		log.Fatal(err)
	}

	// 3. 准备合约地址和接收方地址
	// 这里填写真实的代币合约地址
	contractAddress := common.HexToAddress("0xContractAddress...")
	// 这里填写代币接收方
	toAddress := common.HexToAddress("0xRecipientAddress...")

	// 4. 初始化合约实例 (这就是你问的第一行代码)
	// token.NewErc20Token 是 abigen 自动生成的构造函数
	// "token" 是包名，"Erc20Token" 是我们在命令行指定的 --type
	instance, err := token.NewErc20Token(contractAddress, client)
	if err != nil {
		log.Fatal(err)
	}

	// 5. 构建交易鉴权信息 (Transactor)
	// 获取 ChainID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// 创建授权对象 (这就是你问的第二行代码)
	// 这个对象包含了 nonce, gasPrice, gasLimit, value 等交易元数据
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatal(err)
	}

	// 可选：手动设置 Gas 参数，如果不设置，go-ethereum 会自动估算
	// auth.GasLimit = 300000
	// auth.GasPrice = big.NewInt(20000000000)

	// 6. 调用合约方法 (这就是你问的第三行代码)
	amount := big.NewInt(1000000000000000000) // 1 个代币 (假设精度18)

	// 直接像调用普通 Go 函数一样调用 Transfer
	// 第一个参数必须是 auth，后面的参数对应 Solidity 函数的参数
	tx, err := instance.Transfer(auth, toAddress, amount)
	if err != nil {
		log.Fatal("交易发送失败:", err)
	}
	fmt.Printf("交易已发送! Hash: %s\n", tx.Hash().Hex())
}

func Test03() {
	// 1. 连接节点
	client, err := ethclient.Dial("https://sepolia.infura.io/v3/YOUR_INFURA_ID")
	if err != nil {
		log.Fatal(err)
	}

	// 2. 准备私钥
	privateKey, err := crypto.HexToECDSA("YOUR_PRIVATE_KEY_WITHOUT_0x")
	if err != nil {
		log.Fatal(err)
	}

	// 3. 准备合约地址和接收方地址
	// 这里填写真实的代币合约地址
	contractAddress := common.HexToAddress("0xContractAddress...")
	// 这里填写代币接收方
	toAddress := common.HexToAddress("0xRecipientAddress...")

	// 4. 初始化合约实例 (这就是你问的第一行代码)
	// token.NewErc20Token 是 abigen 自动生成的构造函数
	// "token" 是包名，"Erc20Token" 是我们在命令行指定的 --type
	instance, err := token.NewErc20Token(contractAddress, client)
	if err != nil {
		log.Fatal(err)
	}

	// 5. 构建交易鉴权信息 (Transactor)
	// 获取 ChainID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// 创建授权对象 (这就是你问的第二行代码)
	// 这个对象包含了 nonce, gasPrice, gasLimit, value 等交易元数据
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatal(err)
	}

	// 可选：手动设置 Gas 参数，如果不设置，go-ethereum 会自动估算
	// auth.GasLimit = 300000
	// auth.GasPrice = big.NewInt(20000000000)

	// 6. 调用合约方法 (这就是你问的第三行代码)
	amount := big.NewInt(1000000000000000000) // 1 个代币 (假设精度18)

	// 直接像调用普通 Go 函数一样调用 Transfer
	// 第一个参数必须是 auth，后面的参数对应 Solidity 函数的参数
	tx, err := instance.Transfer(auth, toAddress, amount)
	if err != nil {
		log.Fatal("交易发送失败:", err)
	}

	fmt.Printf("交易已发送! Hash: %s\n", tx.Hash().Hex())
}
