package mfproxy

import (
	"io"
	"net"
	"sync"

	"github.com/zeromicro/go-zero/core/logx"
)

type WrapReadWriteCloserConn struct {
	io.ReadWriteCloser
	underConn          net.Conn
	remoteAddr         net.Addr
	SignalChan         chan struct{}   // Channel for signaling
	Wg                 *sync.WaitGroup // WaitGroup for synchronization
	IntermediateReader net.Conn        // PipeReader for intermediate data
	IntermediateWriter net.Conn        // PipeWriter for intermediate data
	UserWorkConn       net.Conn        // workconn of proxy

	Mutex sync.Mutex // 用于锁定读写的互斥锁
}

// WrapReadWriteCloserToConn remains the same
func WrapReadWriteCloserToConn(rwc io.ReadWriteCloser, underConn net.Conn) *WrapReadWriteCloserConn {
	intermediateReader, intermediateWriter := net.Pipe()

	return &WrapReadWriteCloserConn{
		ReadWriteCloser:    rwc,
		underConn:          underConn,
		remoteAddr:         underConn.RemoteAddr(),
		SignalChan:         make(chan struct{}), // Initialize the channel
		Wg:                 &sync.WaitGroup{},
		IntermediateReader: intermediateReader,
		IntermediateWriter: intermediateWriter,
		UserWorkConn:       nil,
	}
}

// Intercept the Read method
func (conn *WrapReadWriteCloserConn) Read(p []byte) (int, error) {
	// Read data from the underlying ReadWriteCloser
	n, err := conn.ReadWriteCloser.Read(p)
	if err != nil {
		logx.Debugf("conn.ReadWriteCloser.Read err: %v", err)
		return n, err
	}

	if n <= 0 {
		return n, nil
	}

	url, err := getHTTPUrl(p[:n])
	if err == nil {
		mfname, err1 := getMFName(url)
		if err1 == nil {
			mfName = mfname
		}
	}

	return n, nil
}

// Intercept the Write method
func (conn *WrapReadWriteCloserConn) Write(p []byte) (int, error) {
	return conn.ReadWriteCloser.Write(p)
}
