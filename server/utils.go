package server

import (
	"strings"

	"github.com/NethermindEth/aztec-p2p-explorer/repo"
)

// ProcessAgentCounts groups agent counts by agent name and version
func ProcessAgentCounts(counts []*repo.AgentCount) []*AgentVersionCounts {
	agentCounts := make(map[string]*AgentVersionCounts)

	for _, count := range counts {
		parts := strings.SplitN(count.Client, "/", 2)
		agentName := parts[0]
		var agentVersion string
		if len(parts) > 1 {
			agentVersion = parts[1]
		}

		if _, ok := agentCounts[agentName]; !ok {
			agentCounts[agentName] = &AgentVersionCounts{
				AgentName:  agentName,
				TotalCount: 0,
				Versions:   make([]*VersionCount, 0),
			}
		}

		agentCounts[agentName].TotalCount += count.Count

		if agentVersion != "" {
			agentCounts[agentName].Versions = append(agentCounts[agentName].Versions, &VersionCount{
				Version: agentVersion,
				Count:   count.Count,
			})
		}
	}

	// Convert map to slice
	result := make([]*AgentVersionCounts, 0, len(agentCounts))
	for _, v := range agentCounts {
		result = append(result, v)
	}

	return result
}

type AgentVersionCounts struct {
	AgentName  string          `json:"client_name"`
	TotalCount int             `json:"total_count"`
	Versions   []*VersionCount `json:"versions"`
}

type VersionCount struct {
	Version string `json:"version"`
	Count   int    `json:"count"`
}
