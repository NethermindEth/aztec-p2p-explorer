package utils

import (
	"strconv"

	ma "github.com/multiformats/go-multiaddr"

	"github.com/NethermindEth/aztec-p2p-explorer/core/types"
)

func ConvertMultiAddrs(addrs []string) ([]*types.MultiAddr, error) {
	maddrs := []*types.MultiAddr{}
	seen := map[string]bool{}
	for _, addr := range addrs {
		if seen[addr] {
			continue
		}
		seen[addr] = true

		netInfo, err := extractIPs(addr)
		if err != nil {
			return nil, err
		}

		maddrs = append(maddrs, &types.MultiAddr{
			Address: addr,
			IPList:  netInfo,
		})
	}

	return maddrs, nil
}

func extractIPs(addr string) ([]*types.IPInfo, error) {
	maddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return nil, err
	}

	var (
		networkInfoList []*types.IPInfo
		currentIP       string
		currentPort     int
	)

	ma.ForEach(maddr, func(c ma.Component) bool {
		switch c.Protocol().Code {
		case ma.P_IP4, ma.P_IP6:
			// If we have a previous IP and port, add them to the list
			if currentIP != "" && currentPort != 0 {
				networkInfoList = append(networkInfoList, &types.IPInfo{
					IPAddress: currentIP,
					Port:      currentPort,
				})
			}
			currentIP = c.Value()
			currentPort = 0
		case ma.P_TCP, ma.P_UDP:
			p, err := strconv.Atoi(c.Value())
			if err != nil {
				return false
			}
			currentPort = p
		}
		return true
	})

	// Add the last IP and port if they exist
	if currentIP != "" && currentPort != 0 {
		networkInfoList = append(networkInfoList, &types.IPInfo{
			IPAddress: currentIP,
			Port:      currentPort,
		})
	}

	return networkInfoList, nil
}
