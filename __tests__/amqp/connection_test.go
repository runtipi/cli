package amqp_test

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type scriptedTransport struct {
	input        *bytes.Reader
	blockAtEnd   bool
	closed       chan struct{}
	closeOnce    sync.Once
	payloadReads atomic.Int32
	largestRead  atomic.Int64
}

func newTransport(input []byte, blockAtEnd bool) *scriptedTransport {
	return &scriptedTransport{input: bytes.NewReader(input), blockAtEnd: blockAtEnd, closed: make(chan struct{})}
}

func (transport *scriptedTransport) Read(output []byte) (int, error) {
	if requested := int64(len(output)); requested > transport.largestRead.Load() {
		transport.largestRead.Store(requested)
	}

	if transport.input.Len() > 0 {
		return transport.input.Read(output)
	}

	if transport.blockAtEnd {
		<-transport.closed
	} else {
		transport.payloadReads.Add(1)
	}

	return 0, io.EOF
}

func (*scriptedTransport) Write(input []byte) (int, error) {
	return len(input), nil
}

func (transport *scriptedTransport) Close() error {
	transport.closeOnce.Do(func() { close(transport.closed) })
	return nil
}

func TestRejectsOversizedFramesBeforeAuthentication(t *testing.T) {
	tests := []struct {
		name      string
		frameType byte
		size      uint32
		frameSize int
	}{
		{name: "body just beyond the initial frame limit", frameType: 3, size: 4089},
		{name: "large body", frameType: 3, size: 2 * 1024 * 1024},
		{name: "large method", frameType: 1, size: 2 * 1024 * 1024},
		{name: "large content header", frameType: 2, size: 2 * 1024 * 1024},
		{name: "larger configured limit", frameType: 3, size: 2 * 1024 * 1024, frameSize: 131072},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			header := []byte{test.frameType, 0, 0}
			header = binary.BigEndian.AppendUint32(header, test.size)
			transport := newTransport(header, false)
			t.Cleanup(func() { _ = transport.Close() })

			_, err := amqp.Open(transport, amqp.Config{FrameSize: test.frameSize})
			if err == nil {
				t.Fatal("accepted an oversized frame before authentication")
			}
			if reads := transport.payloadReads.Load(); reads != 0 {
				t.Fatalf("attempted %d payload reads before rejecting an oversized frame", reads)
			}
		})
	}
}

func methodFrame(method uint16, fields []byte) []byte {
	payload := binary.BigEndian.AppendUint16(nil, 10)
	payload = binary.BigEndian.AppendUint16(payload, method)
	payload = append(payload, fields...)
	frame := []byte{1, 0, 0}
	frame = binary.BigEndian.AppendUint32(frame, uint32(len(payload)))
	frame = append(frame, payload...)

	return append(frame, 0xce)
}

func longString(value string) []byte {
	field := binary.BigEndian.AppendUint32(nil, uint32(len(value)))
	return append(field, value...)
}

func TestBoundsReadsForTruncatedHandshakeFields(t *testing.T) {
	const declaredSize = 2 * 1024 * 1024

	locales := append([]byte{0, 9, 0, 0, 0, 0}, longString("PLAIN")...)
	locales = binary.BigEndian.AppendUint32(locales, declaredSize)
	tests := []struct {
		name   string
		fields []byte
	}{
		{name: "server properties", fields: binary.BigEndian.AppendUint32([]byte{0, 9}, declaredSize)},
		{name: "mechanisms", fields: binary.BigEndian.AppendUint32([]byte{0, 9, 0, 0, 0, 0}, declaredSize)},
		{name: "locales", fields: locales},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transport := newTransport(methodFrame(10, test.fields), false)
			t.Cleanup(func() { _ = transport.Close() })

			_, err := amqp.Open(transport, amqp.Config{})
			if err == nil {
				t.Fatal("accepted a truncated connection-start field")
			}
			if requested := transport.largestRead.Load(); requested > 128*1024 {
				t.Fatalf("requested a %d-byte read for a truncated handshake field", requested)
			}
		})
	}
}

func TestAcceptsNormalConnectionHandshake(t *testing.T) {
	tests := []struct {
		name          string
		localePadding int
	}{
		{name: "normal"},
		{name: "initial frame limit", localePadding: 4096 - 36},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			start := []byte{0, 9, 0, 0, 0, 0}
			start = append(start, longString("PLAIN")...)
			start = append(start, longString("en_US"+strings.Repeat(" ", test.localePadding))...)
			startFrame := methodFrame(10, start)
			if test.localePadding > 0 && len(startFrame) != 4096 {
				t.Fatalf("boundary fixture has %d bytes, want 4096", len(startFrame))
			}

			tune := binary.BigEndian.AppendUint16(nil, 128)
			tune = binary.BigEndian.AppendUint32(tune, 131072)
			tune = binary.BigEndian.AppendUint16(tune, 0)
			input := append(startFrame, methodFrame(30, tune)...)
			input = append(input, methodFrame(41, []byte{0})...)
			transport := newTransport(input, true)
			t.Cleanup(func() { _ = transport.Close() })

			connection, err := amqp.Open(transport, amqp.Config{
				SASL:   []amqp.Authentication{&amqp.PlainAuth{Username: "tipi", Password: "test-password"}},
				Locale: "en_US",
				Vhost:  "/",
			})
			if err != nil {
				t.Fatalf("normal handshake failed: %v", err)
			}
			if connection.IsClosed() {
				t.Fatal("connection closed during a normal handshake")
			}
			if frameSize := connection.Config.FrameSize; frameSize != 131072 {
				t.Fatalf("negotiated frame size = %d, want 131072", frameSize)
			}
			if vhost := connection.Config.Vhost; vhost != "/" {
				t.Fatalf("negotiated vhost = %q, want /", vhost)
			}

			closed := connection.NotifyClose(make(chan *amqp.Error, 1))
			_ = transport.Close()

			select {
			case <-closed:
			case <-time.After(5 * time.Second):
				t.Fatal("connection did not close after transport shutdown")
			}
		})
	}
}
