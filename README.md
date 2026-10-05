# hud-popup-merge

战斗飘字合并：把短时间内的伤害累加到同一条飘字上。只用 Go 标准库。

```
go test ./...
go run ./cmd/check --sample=cross-source
go run ./cmd/check --sample=window
go run ./cmd/check --sample=stack
go run ./cmd/check --sample=evict
go run ./cmd/check --sample=dup
go run ./cmd/check --sample=work
```

## 口径

- **合并键**：`target`、`source`、`kind` 三者都相同才算同一条飘字。
- **合并窗口**：同键距该条飘字**上一次写入**（首次创建或上一次累加）不超过 `MergeMs` 毫秒时可合并，超过则开新条。窗口随写入滑动。
- **累加上限**：一条飘字最多累计 `MaxStack` 次，首次计 1；到顶后即使仍在窗口内也必须开新条。
- **活跃上限**：同时活跃的飘字不得超过 `MaxPopups`，超出时淘汰最旧的一条并累加 `Dropped`。
- **幂等**：同一个 `eventID` 只处理一次；重复出现直接返回上一次的结果，不改动台账。

## 不变量

- `seq` 从 1 开始单调递增，淘汰不回收序号。
- `amount` 等于该条飘字已合并事件的金额之和。
- `merged` 为假，当且仅当这次调用新建了一条飘字。
- 判定按合并键索引进行，`Scanned` 不得随活跃飘字总数放大。

## 输出契约

每次写入打印一行：

```
{"id": <原样>, "seq": <序号>, "amount": <累计金额>, "merged": <是否合并>}
```

`--sample=evict` 末行打印 `{"active": N, "dropped": N}`；
`--sample=work` 末行打印 `{"popups": N, "scanned": N}`。
