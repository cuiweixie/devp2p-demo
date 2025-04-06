package main

import (
	"crypto/ecdsa"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/nat"
)

var (
	listenAddr  = flag.String("addr", ":30303", "监听地址")
	nodeKeyFile = flag.String("nodekey", "nodekey", "节点私钥文件")
	netrestrict = flag.String("netrestrict", "", "限制网络 CIDR 范围")
	bootnodes   = flag.String("bootnodes", "", "引导节点 enode URLs")
)

// 加载或生成节点私钥
func loadOrGenerateNodeKey(path string) *ecdsa.PrivateKey {
	if _, err := os.Stat(path); err == nil {
		// 文件存在，加载私钥
		key, err := crypto.LoadECDSA(path)
		if err != nil {
			log.Fatalf("加载节点密钥失败: %v", err)
		}
		return key
	} else if os.IsNotExist(err) {
		// 文件不存在，生成新私钥
		key, err := crypto.GenerateKey()
		if err != nil {
			log.Fatalf("生成节点密钥失败: %v", err)
		}
		// 确保目录存在
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			log.Fatalf("创建节点密钥目录失败: %v", err)
		}
		// 保存私钥到文件
		if err := crypto.SaveECDSA(path, key); err != nil {
			log.Fatalf("保存节点密钥失败: %v", err)
		}
		return key
	} else {
		log.Fatalf("检查节点密钥文件失败: %v", err)
		return nil
	}
}

// 解析引导节点
func parseBootnodes(urls string) []*enode.Node {
	if urls == "" {
		return nil
	}
	var nodes []*enode.Node
	for _, url := range strings.Split(urls, ",") {
		if url == "" {
			continue
		}
		node, err := enode.ParseV4(url)
		if err != nil {
			log.Printf("无效的引导节点 URL: %v", err)
			continue
		}
		nodes = append(nodes, node)
	}
	return nodes
}

func main() {
	flag.Parse()

	// 加载或生成节点私钥
	nodeKey := loadOrGenerateNodeKey(*nodeKeyFile)
	nodeID := enode.PubkeyToIDV4(&nodeKey.PublicKey)
	log.Printf("节点 ID: %s", nodeID.String())

	// 创建本地节点配置
	cfg := p2p.Config{
		PrivateKey:     nodeKey,
		MaxPeers:       50,
		Name:           "minimal-devp2p-node",
		ListenAddr:     *listenAddr,
		NAT:            nat.Any(),
		NoDiscovery:    false,
		DiscoveryV4:    true,
		BootstrapNodes: parseBootnodes(*bootnodes),
	}

	// 创建 P2P 服务器
	srv := p2p.Server{Config: cfg}

	// 启动 P2P 服务器
	if err := srv.Start(); err != nil {
		log.Fatalf("启动 P2P 服务器失败: %v", err)
	}
	defer srv.Stop()

	// 打印节点信息
	localNode := srv.LocalNode()
	log.Printf("启动成功，enode: %s", localNode.Node().URLv4())

	// 定期打印连接的对等节点信息
	go func() {
		for {
			log.Printf("当前连接的对等节点数量: %d", srv.PeerCount())
			time.Sleep(10 * time.Second)
		}
	}()

	// 等待中断信号退出
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interrupt
	log.Println("关闭节点...")
}
