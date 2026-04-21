package transport

import (
	"context"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Config holds transport-specific configuration
type Config struct {
	// For Streamable HTTP
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`

	// For SSE
	BaseURL string `json:"base_url,omitempty"`

	// For stdio (experimental, not implemented)
	Command string   `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
}

// ClientConnection wraps an MCP client and session
type ClientConnection struct {
	Client  *mcp.Client
	Session *mcp.ClientSession
}

// NewStreamableConnection creates a connection to a Streamable HTTP server
func NewStreamableConnection(ctx context.Context, url string, headers map[string]string) (*ClientConnection, error) {
	// Create client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "mcp-gateway",
		Version: "1.0.0",
	}, nil)

	// Create HTTP client with custom headers
	httpClient := &http.Client{}
	if len(headers) > 0 {
		// Wrap the transport to add custom headers
		httpClient.Transport = &headerTransport{
			base:    http.DefaultTransport,
			headers: headers,
		}
	}

	// Create transport
	transport := &mcp.StreamableClientTransport{
		Endpoint:   url,
		HTTPClient: httpClient,
	}

	// Connect
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &ClientConnection{
		Client:  client,
		Session: session,
	}, nil
}

// NewSSEConnection creates a connection to an SSE server
func NewSSEConnection(ctx context.Context, baseURL string, headers map[string]string) (*ClientConnection, error) {
	// Create client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "mcp-gateway",
		Version: "1.0.0",
	}, nil)

	// Create HTTP client with custom headers
	httpClient := &http.Client{}
	if len(headers) > 0 {
		httpClient.Transport = &headerTransport{
			base:    http.DefaultTransport,
			headers: headers,
		}
	}

	// Create transport
	transport := &mcp.SSEClientTransport{
		Endpoint:   baseURL,
		HTTPClient: httpClient,
	}

	// Connect
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &ClientConnection{
		Client:  client,
		Session: session,
	}, nil
}

// Close closes the connection
func (c *ClientConnection) Close() error {
	if c.Session != nil {
		return c.Session.Close()
	}
	return nil
}

// headerTransport wraps http.RoundTripper to add custom headers
type headerTransport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	req = req.Clone(req.Context())

	// Add custom headers
	for key, value := range t.headers {
		req.Header.Set(key, value)
	}

	return t.base.RoundTrip(req)
}
