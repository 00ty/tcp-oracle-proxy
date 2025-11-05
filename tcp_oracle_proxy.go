package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	localAddr  = ":9001"
	oracleAddr = "192.168.8.11:1521"
	bufSize    = 64 * 1024
)

var (
	activeTunnels int64
	logger        = log.New(os.Stdout, "", log.LstdFlags|log.Lmsgprefix)
)

func main() {
	ln, err := net.Listen("tcp", localAddr)
	if err != nil {
		logger.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	logger.Printf("proxy ready: %s -> %s", localAddr, oracleAddr)

	// 优雅退出
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		logger.Println("shutting down...")
		cancel()
		ln.Close()
	}()

	for {
		client, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				waitForIdle(5 * time.Second)
				return
			default:
				logger.Printf("accept: %v", err)
				continue
			}
		}
		go handle(ctx, client)
	}
}

func handle(ctx context.Context, client net.Conn) {
	defer client.Close()

	// 自动重拨后端
	var oracle net.Conn
	var err error
	backoff := time.Second
	for {
		oracle, err = net.DialTimeout("tcp", oracleAddr, 3*time.Second)
		if err == nil {
			break
		}
		logger.Printf("dial oracle: %v; retry in %v", err, backoff)
		select {
		case <-time.After(backoff):
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
		case <-ctx.Done():
			return
		}
	}
	defer oracle.Close()

	atomic.AddInt64(&activeTunnels, 1)
	defer atomic.AddInt64(&activeTunnels, -1)

	logger.Printf("tunnel start [%d active]", atomic.LoadInt64(&activeTunnels))
	defer logger.Printf("tunnel end   [%d active]", atomic.LoadInt64(&activeTunnels))

	// 双工转发
	var wg sync.WaitGroup
	wg.Add(2)
	bufPool := &sync.Pool{New: func() interface{} { return make([]byte, bufSize) }}

	go func() { // client -> oracle
		defer wg.Done()
		buf := bufPool.Get().([]byte)
		defer bufPool.Put(buf)
		_, _ = io.CopyBuffer(oracle, client, buf)
		oracle.(*net.TCPConn).CloseWrite()
	}()

	go func() { // oracle -> client
		defer wg.Done()
		buf := bufPool.Get().([]byte)
		defer bufPool.Put(buf)
		_, _ = io.CopyBuffer(client, oracle, buf)
		client.(*net.TCPConn).CloseWrite()
	}()

	// 任意方向 EOF 后，等待另一条路 5 s 内自然结束
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		logger.Println("tunnel force close after 5s idle")
	}
}

// 等所有隧道自然下线或超时
func waitForIdle(timeout time.Duration) {
	tick := time.NewTicker(500 * time.Millisecond)
	defer tick.Stop()
	for {
		if atomic.LoadInt64(&activeTunnels) == 0 {
			logger.Println("all tunnels closed, exit")
			return
		}
		select {
		case <-time.After(timeout):
			logger.Printf("exit after %v grace period", timeout)
			return
		case <-tick.C:
		}
	}
}
