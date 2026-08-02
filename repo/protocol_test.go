package repo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

func TestCreateProtocols(t *testing.T) {
	t.Parallel()

	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	protocolMap := map[string]int{
		"testProto1": 3,
		"testProto2": 2,
	}

	err := repo.CreateProtocols(context.Background(), protocolMap)
	require.NoError(t, err)

	// Verify the protocols were inserted correctly
	dbProtocolOne, err := models.Protocols(models.ProtocolWhere.Protocol.EQ("testProto1")).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, "testProto1", dbProtocolOne.Protocol)
	assert.Equal(t, 3, dbProtocolOne.Count)

	dbProtocolTwo, err := models.Protocols(models.ProtocolWhere.Protocol.EQ("testProto2")).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, "testProto2", dbProtocolTwo.Protocol)
	assert.Equal(t, 2, dbProtocolTwo.Count)
}

func TestUpsertProtocolWithSet(t *testing.T) {
	t.Parallel()

	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	id, err := repo.upsertProtocolsWithSet(context.Background(), repo.db, map[string]int{
		"testProto1": 3,
		"testProto2": 2,
	})
	require.NoError(t, err)

	// Verify the protocols were inserted correctly
	dbProtocolOne, err := models.Protocols(models.ProtocolWhere.Protocol.EQ("testProto1")).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, "testProto1", dbProtocolOne.Protocol)

	dbProtocolTwo, err := models.Protocols(models.ProtocolWhere.Protocol.EQ("testProto2")).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, "testProto2", dbProtocolTwo.Protocol)

	// Verify the protocol set was inserted correctly
	dbProtocolSet, err := models.ProtocolsSets(models.ProtocolsSetWhere.ID.EQ(id)).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, id, dbProtocolSet.ID)
	assert.Len(t, dbProtocolSet.ProtocolIds, 2)
	assert.Contains(t, dbProtocolSet.ProtocolIds, int64(dbProtocolOne.ID))
	assert.Contains(t, dbProtocolSet.ProtocolIds, int64(dbProtocolTwo.ID))
}
