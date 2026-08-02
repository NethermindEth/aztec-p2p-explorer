package repo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

func TestCreateAgentVersions(t *testing.T) {
	t.Parallel()

	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	agentMap := map[string]int{
		"agent1": 3,
		"agent2": 2,
	}

	err := repo.CreateAgentVersions(context.Background(), agentMap)
	require.NoError(t, err)

	// Verify that all agents were inserted correctly
	agentOne, err := models.AgentVersions(models.AgentVersionWhere.AgentVersion.EQ("agent1")).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, "agent1", agentOne.AgentVersion)

	agentTwo, err := models.AgentVersions(models.AgentVersionWhere.AgentVersion.EQ("agent2")).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, "agent2", agentTwo.AgentVersion)
}
