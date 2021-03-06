package test

import (
	"net"
	"testing"
	"time"

	"github.com/Applifier/golang-backend-assignment/internal/client"
	"github.com/Applifier/golang-backend-assignment/internal/server"
	"github.com/Applifier/golang-backend-assignment/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const clientCount = 100
const benchmarkServerPort = 50000

func TestBenchmark(t *testing.T) {
	srv := server.New()
	serverAddr := net.TCPAddr{Port: benchmarkServerPort}
	require.NoError(t, srv.Start(&serverAddr))

	var clients []*client.Client
	var clientChs []chan protocol.MessageFromClient
	for i := 0; i < clientCount; i++ {
		cli := client.New()
		require.NoError(t, cli.Connect(&serverAddr))
		clientCh := make(chan protocol.MessageFromClient)
		go cli.HandleIncomingMessages(clientCh)
		defer func() {
			assert.NoError(t, cli.Close())
		}()
		defer close(clientCh)
		clients = append(clients, cli)
		clientChs = append(clientChs, clientCh)
	}

	defer func() {
		assert.NoError(t, srv.Stop())
	}()

	waitForClientsToConnect(t, srv)

	t.Run("short messages", func(t *testing.T) {
		payload := []byte("FOOBAR")
		result := testing.Benchmark(func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				assert.NoError(b, clients[0].SendMsg(srv.ListClientIDs(), payload))
				for j := 1; j < clientCount; j++ {
					<-clientChs[j]
				}
			}
		})
		t.Logf("Short message benchmark\n%s\n", result.String())
	})

	t.Run("long messages", func(t *testing.T) {
		payload := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Duis sed est id mi blandit fringilla vulputate nec urna. Duis non porttitor arcu. Mauris ac ullamcorper turpis, ac tincidunt risus. In rutrum efficitur porttitor. Cras scelerisque eu mi ut tristique. Phasellus enim elit, pretium ut mi vel, semper interdum nisl. Duis gravida blandit risus, a semper ipsum lacinia quis. Nam eros purus, congue in metus id, volutpat dapibus velit. Cras ut dictum libero, non placerat quam. Vivamus sem justo, varius at magna sed, blandit consequat mi. Cras viverra, orci nec feugiat ullamcorper, mauris erat tincidunt nisi, nec rutrum neque est a libero. Nullam pharetra dolor at erat elementum convallis. Phasellus dictum fermentum odio non eleifend. Etiam scelerisque, neque a fringilla molestie, purus turpis posuere erat, ut pulvinar nisl nisl nec nisl. In pellentesque risus sem, id pretium eros gravida sit amet. In vel massa justo. Fusce euismod mattis massa. Fusce at nibh in est condimentum luctus. Integer a molestie arcu. Suspendisse aliquam venenatis nisl, sit amet aliquam ante convallis quis. Praesent nec ipsum lectus. Ut elementum pretium mollis. ")
		result := testing.Benchmark(func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				assert.NoError(b, clients[0].SendMsg(srv.ListClientIDs(), payload))
				for j := 1; j < clientCount; j++ {
					<-clientChs[j]
				}
			}
		})
		t.Logf("Long message benchmark\n%s\n", result.String())
	})
}

func waitForClientsToConnect(t *testing.T, srv *server.Server) {
	for i := 0; i < 5; i++ {
		if clientCount != len(srv.ListClientIDs()) {
			time.Sleep(time.Millisecond * 200)
		} else {
			break
		}
	}
	// If even after polling for several seconds, not all clients have connected, stop the test
	require.Equal(t, clientCount, len(srv.ListClientIDs()))
}
