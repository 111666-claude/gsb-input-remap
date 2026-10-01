# input-remap

按键重映射：冲突按优先级取最优先、重绑原子、按动作索引查询，只用 Go 标准库。

```
go test ./...
go run ./cmd/input-remap --sample conflict
go run ./cmd/input-remap --sample rebind
go run ./cmd/input-remap --sample scan
```

## 口径（README 为准）

- **冲突取最优先**：同一个键有多条绑定时取 `Priority` 最小的一条，同优先级取 `Action` 字典序最小。
- **重绑原子**：`key` 或 `action` 为空时返回 `false`，并且键位表与之前完全一致（不许先删后加）。
- **按索引查询**：`Action` 不许每次全扫绑定表，`Scanned` 不随绑定数线性增长。
- **不变量**：同一份键位表重复查询结果一致；`Bound` 只数不同按键。
- **规模**：单角色 5000 条绑定、每秒 10 万次查询，单次摊还 O(1)，内存 O(绑定数)。

## 输出契约（不改格式）

```
action=jump
ok=false action=jump
scanned<=3000
```

## 目录

```
remap.go                键位绑定
cmd/input-remap         命令行入口
remap_test.go           go test 用例
```
