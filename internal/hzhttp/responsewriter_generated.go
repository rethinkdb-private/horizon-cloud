// NOTE: This file was automatically generated by mkresponsewriter, do not edit directly.

package hzhttp

import (
	"bufio"
	"net"
	"net/http"
	"time"
)

var responseWriterConstructors = map[responseWriterType]func(http.ResponseWriter, *responseWriterState) http.ResponseWriter{}

type responseWriterWrapHijackFlushCloseNotify struct {
	rws   *responseWriterState
	inner http.ResponseWriter
}

func (w *responseWriterWrapHijackFlushCloseNotify) Header() http.Header {
	return w.inner.Header()
}

func (w *responseWriterWrapHijackFlushCloseNotify) Write(p []byte) (int, error) {
	start := time.Now()
	n, err := w.inner.Write(p)
	w.rws.Transfer.Duration += time.Now().Sub(start)
	w.rws.Transfer.Bytes += int64(n)
	return n, err
}

func (w *responseWriterWrapHijackFlushCloseNotify) WriteHeader(status int) {
	w.rws.Status = status
	w.inner.WriteHeader(status)
}

var _ http.ResponseWriter = &responseWriterWrapHijackFlushCloseNotify{}

func (w *responseWriterWrapHijackFlushCloseNotify) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	c, rw, err := w.inner.(http.Hijacker).Hijack()
	if err != nil && w.rws.Status == 0 {
		w.rws.Status = http.StatusSwitchingProtocols
	}
	return c, rw, err
}

var _ http.Hijacker = &responseWriterWrapHijackFlushCloseNotify{}

func (w *responseWriterWrapHijackFlushCloseNotify) Flush() {
	w.inner.(http.Flusher).Flush()
}

var _ http.Flusher = &responseWriterWrapHijackFlushCloseNotify{}

func (w *responseWriterWrapHijackFlushCloseNotify) CloseNotify() <-chan bool {
	return w.inner.(http.CloseNotifier).CloseNotify()
}

var _ http.CloseNotifier = &responseWriterWrapHijackFlushCloseNotify{}

func init() {
	typ := responseWriterType{
		Hijacker:      true,
		Flusher:       true,
		CloseNotifier: true,
	}
	responseWriterConstructors[typ] = func(w http.ResponseWriter, rws *responseWriterState) http.ResponseWriter {
		return &responseWriterWrapHijackFlushCloseNotify{rws, w}
	}
}

type responseWriterWrapHijackFlush struct {
	rws   *responseWriterState
	inner http.ResponseWriter
}

func (w *responseWriterWrapHijackFlush) Header() http.Header {
	return w.inner.Header()
}

func (w *responseWriterWrapHijackFlush) Write(p []byte) (int, error) {
	start := time.Now()
	n, err := w.inner.Write(p)
	w.rws.Transfer.Duration += time.Now().Sub(start)
	w.rws.Transfer.Bytes += int64(n)
	return n, err
}

func (w *responseWriterWrapHijackFlush) WriteHeader(status int) {
	w.rws.Status = status
	w.inner.WriteHeader(status)
}

var _ http.ResponseWriter = &responseWriterWrapHijackFlush{}

func (w *responseWriterWrapHijackFlush) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	c, rw, err := w.inner.(http.Hijacker).Hijack()
	if err != nil && w.rws.Status == 0 {
		w.rws.Status = http.StatusSwitchingProtocols
	}
	return c, rw, err
}

var _ http.Hijacker = &responseWriterWrapHijackFlush{}

func (w *responseWriterWrapHijackFlush) Flush() {
	w.inner.(http.Flusher).Flush()
}

var _ http.Flusher = &responseWriterWrapHijackFlush{}

func init() {
	typ := responseWriterType{
		Hijacker:      true,
		Flusher:       true,
		CloseNotifier: false,
	}
	responseWriterConstructors[typ] = func(w http.ResponseWriter, rws *responseWriterState) http.ResponseWriter {
		return &responseWriterWrapHijackFlush{rws, w}
	}
}

type responseWriterWrapHijackCloseNotify struct {
	rws   *responseWriterState
	inner http.ResponseWriter
}

func (w *responseWriterWrapHijackCloseNotify) Header() http.Header {
	return w.inner.Header()
}

func (w *responseWriterWrapHijackCloseNotify) Write(p []byte) (int, error) {
	start := time.Now()
	n, err := w.inner.Write(p)
	w.rws.Transfer.Duration += time.Now().Sub(start)
	w.rws.Transfer.Bytes += int64(n)
	return n, err
}

func (w *responseWriterWrapHijackCloseNotify) WriteHeader(status int) {
	w.rws.Status = status
	w.inner.WriteHeader(status)
}

var _ http.ResponseWriter = &responseWriterWrapHijackCloseNotify{}

func (w *responseWriterWrapHijackCloseNotify) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	c, rw, err := w.inner.(http.Hijacker).Hijack()
	if err != nil && w.rws.Status == 0 {
		w.rws.Status = http.StatusSwitchingProtocols
	}
	return c, rw, err
}

var _ http.Hijacker = &responseWriterWrapHijackCloseNotify{}

func (w *responseWriterWrapHijackCloseNotify) CloseNotify() <-chan bool {
	return w.inner.(http.CloseNotifier).CloseNotify()
}

var _ http.CloseNotifier = &responseWriterWrapHijackCloseNotify{}

func init() {
	typ := responseWriterType{
		Hijacker:      true,
		Flusher:       false,
		CloseNotifier: true,
	}
	responseWriterConstructors[typ] = func(w http.ResponseWriter, rws *responseWriterState) http.ResponseWriter {
		return &responseWriterWrapHijackCloseNotify{rws, w}
	}
}

type responseWriterWrapHijack struct {
	rws   *responseWriterState
	inner http.ResponseWriter
}

func (w *responseWriterWrapHijack) Header() http.Header {
	return w.inner.Header()
}

func (w *responseWriterWrapHijack) Write(p []byte) (int, error) {
	start := time.Now()
	n, err := w.inner.Write(p)
	w.rws.Transfer.Duration += time.Now().Sub(start)
	w.rws.Transfer.Bytes += int64(n)
	return n, err
}

func (w *responseWriterWrapHijack) WriteHeader(status int) {
	w.rws.Status = status
	w.inner.WriteHeader(status)
}

var _ http.ResponseWriter = &responseWriterWrapHijack{}

func (w *responseWriterWrapHijack) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	c, rw, err := w.inner.(http.Hijacker).Hijack()
	if err != nil && w.rws.Status == 0 {
		w.rws.Status = http.StatusSwitchingProtocols
	}
	return c, rw, err
}

var _ http.Hijacker = &responseWriterWrapHijack{}

func init() {
	typ := responseWriterType{
		Hijacker:      true,
		Flusher:       false,
		CloseNotifier: false,
	}
	responseWriterConstructors[typ] = func(w http.ResponseWriter, rws *responseWriterState) http.ResponseWriter {
		return &responseWriterWrapHijack{rws, w}
	}
}

type responseWriterWrapFlushCloseNotify struct {
	rws   *responseWriterState
	inner http.ResponseWriter
}

func (w *responseWriterWrapFlushCloseNotify) Header() http.Header {
	return w.inner.Header()
}

func (w *responseWriterWrapFlushCloseNotify) Write(p []byte) (int, error) {
	start := time.Now()
	n, err := w.inner.Write(p)
	w.rws.Transfer.Duration += time.Now().Sub(start)
	w.rws.Transfer.Bytes += int64(n)
	return n, err
}

func (w *responseWriterWrapFlushCloseNotify) WriteHeader(status int) {
	w.rws.Status = status
	w.inner.WriteHeader(status)
}

var _ http.ResponseWriter = &responseWriterWrapFlushCloseNotify{}

func (w *responseWriterWrapFlushCloseNotify) Flush() {
	w.inner.(http.Flusher).Flush()
}

var _ http.Flusher = &responseWriterWrapFlushCloseNotify{}

func (w *responseWriterWrapFlushCloseNotify) CloseNotify() <-chan bool {
	return w.inner.(http.CloseNotifier).CloseNotify()
}

var _ http.CloseNotifier = &responseWriterWrapFlushCloseNotify{}

func init() {
	typ := responseWriterType{
		Hijacker:      false,
		Flusher:       true,
		CloseNotifier: true,
	}
	responseWriterConstructors[typ] = func(w http.ResponseWriter, rws *responseWriterState) http.ResponseWriter {
		return &responseWriterWrapFlushCloseNotify{rws, w}
	}
}

type responseWriterWrapFlush struct {
	rws   *responseWriterState
	inner http.ResponseWriter
}

func (w *responseWriterWrapFlush) Header() http.Header {
	return w.inner.Header()
}

func (w *responseWriterWrapFlush) Write(p []byte) (int, error) {
	start := time.Now()
	n, err := w.inner.Write(p)
	w.rws.Transfer.Duration += time.Now().Sub(start)
	w.rws.Transfer.Bytes += int64(n)
	return n, err
}

func (w *responseWriterWrapFlush) WriteHeader(status int) {
	w.rws.Status = status
	w.inner.WriteHeader(status)
}

var _ http.ResponseWriter = &responseWriterWrapFlush{}

func (w *responseWriterWrapFlush) Flush() {
	w.inner.(http.Flusher).Flush()
}

var _ http.Flusher = &responseWriterWrapFlush{}

func init() {
	typ := responseWriterType{
		Hijacker:      false,
		Flusher:       true,
		CloseNotifier: false,
	}
	responseWriterConstructors[typ] = func(w http.ResponseWriter, rws *responseWriterState) http.ResponseWriter {
		return &responseWriterWrapFlush{rws, w}
	}
}

type responseWriterWrapCloseNotify struct {
	rws   *responseWriterState
	inner http.ResponseWriter
}

func (w *responseWriterWrapCloseNotify) Header() http.Header {
	return w.inner.Header()
}

func (w *responseWriterWrapCloseNotify) Write(p []byte) (int, error) {
	start := time.Now()
	n, err := w.inner.Write(p)
	w.rws.Transfer.Duration += time.Now().Sub(start)
	w.rws.Transfer.Bytes += int64(n)
	return n, err
}

func (w *responseWriterWrapCloseNotify) WriteHeader(status int) {
	w.rws.Status = status
	w.inner.WriteHeader(status)
}

var _ http.ResponseWriter = &responseWriterWrapCloseNotify{}

func (w *responseWriterWrapCloseNotify) CloseNotify() <-chan bool {
	return w.inner.(http.CloseNotifier).CloseNotify()
}

var _ http.CloseNotifier = &responseWriterWrapCloseNotify{}

func init() {
	typ := responseWriterType{
		Hijacker:      false,
		Flusher:       false,
		CloseNotifier: true,
	}
	responseWriterConstructors[typ] = func(w http.ResponseWriter, rws *responseWriterState) http.ResponseWriter {
		return &responseWriterWrapCloseNotify{rws, w}
	}
}

type responseWriterWrap struct {
	rws   *responseWriterState
	inner http.ResponseWriter
}

func (w *responseWriterWrap) Header() http.Header {
	return w.inner.Header()
}

func (w *responseWriterWrap) Write(p []byte) (int, error) {
	start := time.Now()
	n, err := w.inner.Write(p)
	w.rws.Transfer.Duration += time.Now().Sub(start)
	w.rws.Transfer.Bytes += int64(n)
	return n, err
}

func (w *responseWriterWrap) WriteHeader(status int) {
	w.rws.Status = status
	w.inner.WriteHeader(status)
}

var _ http.ResponseWriter = &responseWriterWrap{}

func init() {
	typ := responseWriterType{
		Hijacker:      false,
		Flusher:       false,
		CloseNotifier: false,
	}
	responseWriterConstructors[typ] = func(w http.ResponseWriter, rws *responseWriterState) http.ResponseWriter {
		return &responseWriterWrap{rws, w}
	}
}
