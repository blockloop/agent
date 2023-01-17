package opencensus_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/grafana/agent/component/otelcol/receiver/opencensus"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/util"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/opencensusreceiver"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
)

// Test ensures that otelcol.receiver.opencensus can start successfully.
func Test(t *testing.T) {
	httpAddr := getFreeAddr(t)

	ctx := componenttest.TestContext(t)
	l := util.TestLogger(t)

	ctrl, err := componenttest.NewControllerFromID(l, "otelcol.receiver.opencensus")
	require.NoError(t, err)

	cfg := fmt.Sprintf(`
		grpc {
			endpoint = "%s"
			transport = "tcp"
		}

		output { /* no-op */ }
	`, httpAddr)

	var args opencensus.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	go func() {
		err := ctrl.Run(ctx, args)
		require.NoError(t, err)
	}()

	require.NoError(t, ctrl.WaitRunning(time.Second))
}

func TestArguments_UnmarshalRiver(t *testing.T) {
	t.Run("grpc", func(t *testing.T) {
		httpAddr := getFreeAddr(t)
		in := fmt.Sprintf(`
		cors_allowed_origins = ["https://*.test.com", "https://test.com"]

		grpc {
			endpoint = "%s"
			transport = "tcp"
		}

		output { /* no-op */ }
		`, httpAddr)

		var args opencensus.Arguments
		require.NoError(t, river.Unmarshal([]byte(in), &args))
		args.Convert()
		otelArgs, err := (args.Convert()).(*opencensusreceiver.Config)

		require.True(t, err)

		// Check the gRPC arguments
		require.Equal(t, otelArgs.NetAddr.Endpoint, httpAddr)
		require.Equal(t, otelArgs.NetAddr.Transport, "tcp")

		// Check the CORS arguments
		require.Equal(t, len(otelArgs.CorsOrigins), 2)
		require.Equal(t, otelArgs.CorsOrigins[0], "https://*.test.com")
		require.Equal(t, otelArgs.CorsOrigins[1], "https://test.com")
	})
}

func getFreeAddr(t *testing.T) string {
	t.Helper()

	portNumber, err := freeport.GetFreePort()
	require.NoError(t, err)

	return fmt.Sprintf("localhost:%d", portNumber)
}
