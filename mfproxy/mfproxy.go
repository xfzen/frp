package mfproxy

import (
	"bytes"
	"io"
	"net"
	"strings"
)

// HTTPDetectingReadWriteCloser wraps an io.ReadWriteCloser and detects HTTP requests.
type HTTPDetectingReadWriteCloser struct {
	conn          net.Conn
	buf           *bytes.Buffer
	isHTTPRequest bool
}

// NewHTTPDetectingReadWriteCloser creates a new HTTPDetectingReadWriteCloser.
func NewHTTPDetectingReadWriteCloser(conn net.Conn, bufferSize int) *HTTPDetectingReadWriteCloser {
	return &HTTPDetectingReadWriteCloser{
		conn: conn,
		buf:  bytes.NewBuffer(make([]byte, 0, bufferSize)),
	}
}

// Read reads data from the connection and detects if it's an HTTP request.
func (d *HTTPDetectingReadWriteCloser) Read(p []byte) (int, error) {
	n, err := d.conn.Read(p)
	if err != nil && err != io.EOF {
		return n, err
	}

	// Buffer the data for HTTP request detection
	d.buf.Write(p[:n])

	// If not already detected, try to detect HTTP request
	if !d.isHTTPRequest {
		data := d.buf.Bytes()
		if isHTTPRequest(data) {
			d.isHTTPRequest = true
		}
	}

	// Copy the buffered data to the provided buffer
	copy(p, d.buf.Bytes())
	if len(d.buf.Bytes()) > len(p) {
		d.buf = bytes.NewBuffer(d.buf.Bytes()[len(p):])
	} else {
		d.buf.Reset()
	}

	return n, err
}

// Write writes data to the underlying connection.
func (d *HTTPDetectingReadWriteCloser) Write(p []byte) (int, error) {
	return d.conn.Write(p)
}

// Close closes the underlying connection.
func (d *HTTPDetectingReadWriteCloser) Close() error {
	return d.conn.Close()
}

// isHTTPRequest checks if the provided data is an HTTP request.
func isHTTPRequest(data []byte) bool {
	datalen := len(data)
	if datalen < 1024 {
		// logx.Debugf("\n%v", string(data))
	}

	// HTTP request must start with a request line
	lines := bytes.SplitN(data, []byte("\r\n"), 2)
	if len(lines) < 1 {
		return false
	}

	// Check if the request line starts with an HTTP method
	requestLine := string(lines[0])
	parts := strings.SplitN(requestLine, " ", 3)
	if len(parts) < 3 {
		return false
	}

	// HTTP request line looks like "METHOD /path HTTP/1.1"
	method := parts[0]
	version := parts[2]

	// Simple check for HTTP method and version
	if (method == "GET" || method == "POST" || method == "PUT" || method == "DELETE" || method == "HEAD" || method == "OPTIONS") &&
		(strings.HasPrefix(version, "HTTP/")) {
		return true
	}

	return false
}
