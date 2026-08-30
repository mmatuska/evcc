package modbus

import (
	"encoding/binary"
	"io"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/andig/mbserver"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/modbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrentRead(t *testing.T) {
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer l.Close()

	srv, _ := mbserver.New(&echoHandler{
		id:             0,
		RequestHandler: new(mbserver.DummyHandler),
	})
	require.NoError(t, srv.Start(l))
	defer func() { _ = srv.Stop() }()

	var wg sync.WaitGroup

	for id := 1; id <= 10; id++ {
		wg.Go(func() {
			// client
			conn, err := modbus.NewConnection(t.Context(), l.Addr().String(), "", "", 0, modbus.Tcp, uint8(id))
			require.NoError(t, err)

			for range 50 {
				addr := uint16(rand.Int31n(200) + 1)
				qty := uint16(rand.Int31n(32) + 1)

				b, err := conn.ReadInputRegisters(addr, qty)
				require.NoError(t, err)

				if err == nil {
					for u := range qty {
						assert.Equal(t, addr^uint16(id)^u, binary.BigEndian.Uint16(b[2*u:]))
					}
				}

				time.Sleep(time.Duration(rand.Int31n(1000)) * time.Microsecond)
			}
		})
	}

	wg.Wait()
}

func TestReadCoils(t *testing.T) {
	// downstream server
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer l.Close()

	srv, _ := mbserver.New(&echoHandler{
		id:             0,
		RequestHandler: new(mbserver.DummyHandler),
	})
	require.NoError(t, srv.Start(l))
	defer func() { _ = srv.Stop() }()

	// proxy server
	pl, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer pl.Close()

	downstreamConn, err := modbus.NewConnection(t.Context(), l.Addr().String(), "", "", 0, modbus.Tcp, 1)
	require.NoError(t, err)

	proxy, _ := mbserver.New(&handler{
		log:  util.NewLogger("foo"),
		conn: downstreamConn,
	})
	require.NoError(t, proxy.Start(pl))
	defer func() { _ = proxy.Stop() }()

	// test client
	{
		conn, err := modbus.NewConnection(t.Context(), pl.Addr().String(), "", "", 0, modbus.Tcp, 1)
		require.NoError(t, err)

		{ // read
			b, err := conn.ReadCoils(1, 1)
			require.NoError(t, err)
			assert.Equal(t, []byte{0x01}, b)

			b, err = conn.ReadCoils(1, 2)
			require.NoError(t, err)
			assert.Equal(t, []byte{0x03}, b)

			b, err = conn.ReadCoils(1, 9)
			require.NoError(t, err)
			assert.Equal(t, []byte{0xFF, 0x01}, b)
		}
		{ // write
			b, err := conn.WriteSingleCoil(1, 0xFF00)
			require.NoError(t, err)
			assert.Equal(t, []byte{0xFF, 0x00}, b)

			b, err = conn.WriteMultipleCoils(1, 9, []byte{0xFF, 0x01})
			require.NoError(t, err)
			assert.Equal(t, []byte{0x00, 0x09}, b)
		}
	}
}

func TestWriteMultipleRegistersMalformedDownstreamResponse(t *testing.T) {
	downstreamListener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer downstreamListener.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)

		conn, err := downstreamListener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		header := make([]byte, 7)
		if _, err := io.ReadFull(conn, header); err != nil {
			return
		}

		payload := make([]byte, int(binary.BigEndian.Uint16(header[4:6]))-1)
		if _, err := io.ReadFull(conn, payload); err != nil {
			return
		}

		resp := []byte{
			header[0], header[1], // transaction id
			header[2], header[3], // protocol id
			0x00, 0x03, // malformed length
			header[6],  // unit id
			payload[0], // function code
			0x00,       // malformed short payload
		}
		_, _ = conn.Write(resp)
	}()

	// proxy server
	proxyListener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer proxyListener.Close()

	downstreamConn, err := modbus.NewConnection(t.Context(), downstreamListener.Addr().String(), "", "", 0, modbus.Tcp, 1)
	require.NoError(t, err)

	proxy, _ := mbserver.New(&handler{
		log:  util.NewLogger("foo"),
		conn: downstreamConn,
	})
	require.NoError(t, proxy.Start(proxyListener))
	defer func() { _ = proxy.Stop() }()

	clientConn, err := modbus.NewConnection(t.Context(), proxyListener.Addr().String(), "", "", 0, modbus.Tcp, 1)
	require.NoError(t, err)

	b, err := clientConn.WriteMultipleRegisters(0x007f, 1, []byte{0x00, 0x01})
	require.NoError(t, err)
	assert.Equal(t, []byte{0x00, 0x01}, b)

	<-done
}

type echoHandler struct {
	id int
	mbserver.RequestHandler
}

func (h *echoHandler) HandleInputRegisters(req *mbserver.InputRegistersRequest) (res []uint16, err error) {
	for u := uint16(0); u < req.Quantity; u++ {
		res = append(res, req.Addr^uint16(req.UnitId)^u)
	}

	return res, err
}

func (h *echoHandler) HandleCoils(req *mbserver.CoilsRequest) (res []bool, err error) {
	if req.IsWrite {
		return nil, nil
	}

	for u := uint16(0); u < req.Quantity; u++ {
		res = append(res, true)
	}

	return res, err
}
