// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httputil

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

// One of the copies, say from b to r2, could be avoided by using a more
// elaborate trick where the other copy is made during Request/Response.Write.
// This would complicate things too much, given that these functions are for
// debugging only.
func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, nil, err
	}
	if err = b.Close(); err != nil {
		return nil, nil, err
	}
	return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewBuffer(buf.Bytes())), nil
}

// dumpConn is a net.Conn which writes to Writer and reads from Reader
type dumpConn struct {
	io.Writer
	io.Reader
}

func (c *dumpConn) Close() error                       { return nil }
func (c *dumpConn) LocalAddr() net.Addr                { return nil }
func (c *dumpConn) RemoteAddr() net.Addr               { return nil }
func (c *dumpConn) SetDeadline(t time.Time) error      { return nil }
func (c *dumpConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *dumpConn) SetWriteDeadline(t time.Time) error { return nil }

// DumpRequestOut is like DumpRequest but includes
// headers that the standard http.Transport adds,
// such as User-Agent.
func DumpRequestOut(req *http.Request, body bool) ([]byte, error) {
	save := req.Body
	if !body || req.Body == nil {
		req.Body = nil
	} else {
		var err error
		save, req.Body, err = drainBody(req.Body)
		if err != nil {
			return nil, err
		}
	}

	// Since we're using the actual Transport code to write the request,
	// switch to http so the Transport doesn't try to do an SSL
	// negotiation with our dumpConn and its bytes.Buffer & pipe.
	// The wire format for https and http are the same, anyway.
	if req.URL.Scheme == "https" {
		defer func() { req.URL.Scheme = "https" }()
		req.URL.Scheme = "http"
	}

	// Use the actual Transport code to record what we would send
	// on the wire, but not using TCP.  Use a Transport with a
	// customer dialer that returns a fake net.Conn that waits
	// for the full input (and recording it), and then responds
	// with a dummy response.
	var buf bytes.Buffer // records the output
	pr, pw := io.Pipe()
	dr := &delegateReader{c: make(chan io.Reader)}
	// Wait for the request before replying with a dummy response:
	go func() {
		http.ReadRequest(bufio.NewReader(pr))
		dr.c <- strings.NewReader("HTTP/1.1 204 No Content\r\n\r\n")
	}()

	t := &http.Transport{
		Dial: func(net, addr string) (net.Conn, error) {
			return &dumpConn{io.MultiWriter(pw, &buf), dr}, nil
		},
	}

	_, err := t.RoundTrip(req)

	req.Body = save
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// delegateReader is a reader that delegates to another reader,
// once it arrives on a channel.
type delegateReader struct {
	c chan io.Reader
	r io.Reader // nil until received from c
}

func (r *delegateReader) Read(p []byte) (int, error) {
	if r.r == nil {
		r.r = <-r.c
	}
	return r.r.Read(p)
}

// Return value if nonempty, def otherwise.
func valueOrDefault(value, def string) string {
	if value != "" {
		return value
	}
	return def
}

var reqWriteExcludeHeaderDump = map[string]bool{
	"Host":              true, // not in Header map anyway
	"Content-Length":    true,
	"Transfer-Encoding": true,
	"Trailer":           true,
}

// dumpAsReceived writes req to w in the form as it was received, or
// at least as accurately as possible from the information retained in
// the request.
func dumpAsReceived(req *http.Request, w io.Writer) error {
	return nil
}

// DumpRequest returns the as-received wire representation of req,
// optionally including the request body, for debugging.
// DumpRequest is semantically a no-op, but in order to
// dump the body, it reads the body data into memory and
// changes req.Body to refer to the in-memory copy.
// The documentation for http.Request.Write details which fields
// of req are used.
func DumpRequest(req *http.Request, body bool) (dump []byte, err error) {
	save := req.Body
	if !body || req.Body == nil {
		req.Body = nil
	} else {
		save, req.Body, err = drainBody(req.Body)
		if err != nil {
			return
		}
	}

	var b bytes.Buffer

	fmt.Fprintf(&b, "%s %s HTTP/%d.%d\r\n", valueOrDefault(req.Method, "GET"),
		req.URL.RequestURI(), req.ProtoMajor, req.ProtoMinor)

	host := req.Host
	if host == "" && req.URL != nil {
		host = req.URL.Host
	}
	if host != "" {
		fmt.Fprintf(&b, "Host: %s\r\n", host)
	}

	chunked := len(req.TransferEncoding) > 0 && req.TransferEncoding[0] == "chunked"
	if len(req.TransferEncoding) > 0 {
		fmt.Fprintf(&b, "Transfer-Encoding: %s\r\n", strings.Join(req.TransferEncoding, ","))
	}
	if req.Close {
		fmt.Fprintf(&b, "Connection: close\r\n")
	}

	err = req.Header.WriteSubset(&b, reqWriteExcludeHeaderDump)
	if err != nil {
		return
	}

	io.WriteString(&b, "\r\n")

	if req.Body != nil {
		var dest io.Writer = &b
		if chunked {
			dest = NewChunkedWriter(dest)
		}
		_, err = io.Copy(dest, req.Body)
		if chunked {
			dest.(io.Closer).Close()
			io.WriteString(&b, "\r\n")
		}
	}

	req.Body = save
	if err != nil {
		return
	}
	dump = b.Bytes()
	return
}

// DumpResponse is like DumpRequest but dumps a response.
func DumpResponse(resp *http.Response, body bool) (dump []byte, err error) {
	var b bytes.Buffer
	save := resp.Body
	savecl := resp.ContentLength
	if !body || resp.Body == nil {
		resp.Body = nil
		resp.ContentLength = 0
	} else {
		save, resp.Body, err = drainBody(resp.Body)
		if err != nil {
			return
		}
	}
	err = resp.Write(&b)
	resp.Body = save
	resp.ContentLength = savecl
	if err != nil {
		return
	}
	dump = b.Bytes()
	return
}
