# 扫链
1 **获取每个区块，获取区块中的每笔交易并判断是否数据业务需要的数据：**
* 判断交易是不是普通ETH转账还是合约转账,如果是普通转账，要判断发送地址或者接受地址是不是我们自己管理的账号
* 如果是合约，要看是合约转ETH还是普通调用合约方法，具体看看Value是否大于0，to的地址是不是自己管理的合约地址
* 如果是普通调用合约方法，就要判断to对应合约地址属于哪一个地址
* 获取交易的同时获取收据（使用eth_getBlockReceipts），交易的下标跟收据下标一样，需要通过收据来判断交易的状态，是否交易成功
* 处理收据的同时需要解析事件

2 **使用websocket监听特定事件，实时性强但是容易受网络影响，另外数据的可靠性比较差，适合一些实时要求高的场景**

3 **链重组(ReOrg)**
* 扫链可以使用确认数：如ETH 12、BSC 15、Polygon 128 
  
  具体做法：
  1. 判断扫描的区块：scanned block number <= current block number - confirmed number, 区块扫描后确认对应的状态为confirmed
  2. 经过上述扫描策略如果还是会发生ReOrg，那么会有以下两个方法兜底：
     * 上述第一个方案扫描的区块都会去判断当前区块的父区块hash跟上次落库的区块hash 是否一致。如果不一致，既发生重组，那么一般情况下需要发送告警，上游部分核心业务需要停止，
       然后需要修正数据库数据，对应被重组的数据需要更改状态，并重新在发生重组的节点获取最新的数据，并落库
     * 服务会有另外一个协程，每天定时的获取finalized 区块并修改区块状态，这个流程属于低频操作，主要被确认为finalized 则数据是不会回滚
     
4 **表核心字段设计**
     
   * 区块表(chain id、block hash)
   * 交易记录表(chain_id, tx_hash)
   * 交易回执表(chain_id, tx_hash)
   * 事件记录表(chain_id, tx_hash, log_index)

5 **数据状态**

区块状态（5-7种）              │  交易状态（6-8种）              │
│  ─────────────────────────────  │  ─────────────────────────────│
│  1. pending     - 待同步        │  1. pending    - 待确认       │
│  2. syncing     - 同步中        │  2. included   - 已入块       │
│  3. synced      - 已同步        │  3. confirmed  - 已确认       │
│  4. confirmed   - 已确认        │  4. finalized  - 最终确认     │
│  5. finalized   - 最终确认      │  5. failed     - 执行失败     │
│  6. reorged     - 已重组        │  6. replaced   - 被替换       │
│  7. orphaned    - 孤块          │  7. dropped    - 被丢弃       │
│                                 │  8. reorged    - 已重组 


6 **解析数据**

eventSignature := []byte("ItemSet(bytes32,bytes32)")
hash := crypto.Keccak256Hash(eventSignature)

for _, vLog := range logs {
     event := struct {
     Key   [32]byte
     Value [32]byte
     }{}
  if vLog.Topics[0].Hex()==eventSignature{
      err := contractAbi.Unpack(&event, "ItemSet", vLog.Data)
      if err != nil {
          log.Fatal(err)
      }
  }
}

7 **NFT订单流程**

流程：
卖家挂售：
钱包签名 → POST /api/v1/orders (side=0) → 订单簿存储

买家浏览：
GET /api/v1/listings?nft_contract=0x...&token_id=6529
→ 前端展示价格列表

买家购买（Buy Now）：
GET /api/v1/orders/{hash}  ← 取出卖家的完整签名订单
→ 前端用卖家签名 + 买家自己的参数，调用合约fulfillOrder()
→ 链上原子交换完成
→ 后端监听链上事件 → MarkFilled(orderHash)