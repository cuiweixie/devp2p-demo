# 1. start bootnode
```shell
go run main.go
```
# 2. start node2
```go
go run main.go --bootnodes "enode://6a7dec0d36c65bc44fb24ad09427c8b901fb623db1f8d05db8f95a155ec8497548b453d1b92e661b1398f79710ff4b39fa2a2c1c1072eb2a49ea473fc5c1ffb6@127.0.0.1:30303" --addr ":30304" --nodekey nodekey2
```
