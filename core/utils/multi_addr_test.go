package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/NethermindEth/aztec-p2p-explorer/core/types"
)

func TestConvertMultiAddrs(t *testing.T) {
	testCases := []struct {
		name     string
		addrs    []string
		expected []*types.MultiAddr
		wantErr  bool
	}{
		{
			name: "Single IPv4 address",
			addrs: []string{
				"/ip4/192.168.1.1/tcp/8080",
			},
			expected: []*types.MultiAddr{
				{
					Address: "/ip4/192.168.1.1/tcp/8080",
					IPList: []*types.IPInfo{
						{
							IPAddress: "192.168.1.1",
							Port:      8080,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple addresses with IPv4 and IPv6",
			addrs: []string{
				"/ip4/192.168.1.1/tcp/8080",
				"/ip6/2001:db8::1/udp/1234",
				"/ip4/10.0.0.1/tcp/9090/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N",
			},
			expected: []*types.MultiAddr{
				{
					Address: "/ip4/192.168.1.1/tcp/8080",
					IPList: []*types.IPInfo{
						{
							IPAddress: "192.168.1.1",
							Port:      8080,
						},
					},
				},
				{
					Address: "/ip6/2001:db8::1/udp/1234",
					IPList: []*types.IPInfo{
						{
							IPAddress: "2001:db8::1",
							Port:      1234,
						},
					},
				},
				{
					Address: "/ip4/10.0.0.1/tcp/9090/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N",
					IPList: []*types.IPInfo{
						{
							IPAddress: "10.0.0.1",
							Port:      9090,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Address with multiple IP/Port pairs",
			addrs: []string{
				"/ip4/192.168.1.1/tcp/8080/ip6/2001:db8::1/udp/1234",
			},
			expected: []*types.MultiAddr{
				{
					Address: "/ip4/192.168.1.1/tcp/8080/ip6/2001:db8::1/udp/1234",
					IPList: []*types.IPInfo{
						{
							IPAddress: "192.168.1.1",
							Port:      8080,
						},
						{
							IPAddress: "2001:db8::1",
							Port:      1234,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "Empty input",
			addrs:    []string{},
			expected: []*types.MultiAddr{},
			wantErr:  false,
		},
		{
			name: "Invalid address",
			addrs: []string{
				"invalid_address",
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ConvertMultiAddrs(tc.addrs)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestExtractIPs(t *testing.T) {
	testCases := []struct {
		name     string
		addr     string
		expected []*types.IPInfo
		wantErr  bool
	}{
		{
			name: "Valid IPv4 address",
			addr: "/ip4/192.168.1.1/tcp/8080",
			expected: []*types.IPInfo{
				{
					IPAddress: "192.168.1.1",
					Port:      8080,
				},
			},
			wantErr: false,
		},
		{
			name: "Valid IPv6 address",
			addr: "/ip6/2001:db8::1/udp/1234",
			expected: []*types.IPInfo{
				{
					IPAddress: "2001:db8::1",
					Port:      1234,
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple IP/Port pairs",
			addr: "/ip4/192.168.1.1/tcp/8080/ip6/2001:db8::1/udp/1234",
			expected: []*types.IPInfo{
				{
					IPAddress: "192.168.1.1",
					Port:      8080,
				},
				{
					IPAddress: "2001:db8::1",
					Port:      1234,
				},
			},
			wantErr: false,
		},
		{
			name:     "Invalid address",
			addr:     "invalid_address",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := extractIPs(tc.addr)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
