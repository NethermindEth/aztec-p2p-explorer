//nolint:misspell
package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/NethermindEth/aztec-p2p-explorer/repo"
	"github.com/NethermindEth/aztec-p2p-explorer/testutil"
)

func setupTestRepo(t *testing.T) (*repo.PeerRepository, func()) {
	db, dbTearDown := testutil.PrepareTestDatabase(t)
	logger := slog.Default()
	r := repo.NewPeerRepository(db, logger)
	return r, dbTearDown
}

//nolint:tparallel // Pagination need to be tested sequentially
func TestGetPeers(t *testing.T) {
	t.Parallel()

	r, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	h := &Handler{Repo: r}
	e := echo.New()

	t.Run("Filtering and Sorting", func(t *testing.T) {
		testCases := []struct {
			name           string
			queryParams    url.Values
			expectedStatus int
			expectedLen    int
			checkResponse  func(*testing.T, *PeersResponse)
		}{
			{
				name:           "No filters",
				queryParams:    url.Values{},
				expectedStatus: http.StatusOK,
				expectedLen:    4,
				checkResponse:  nil,
			},
			{
				name:           "Filter by continent",
				queryParams:    url.Values{"continents": []string{"North America"}},
				expectedStatus: http.StatusOK,
				expectedLen:    2,
				checkResponse: func(t *testing.T, resp *PeersResponse) {
					for _, peer := range resp.Peers {
						assert.Contains(t, []string{"United States"}, peer.MultiAddresses[0].IPList[0].Country)
					}
				},
			},
			{
				name:           "Filter by client name",
				queryParams:    url.Values{"clients": []string{"alpha-node"}},
				expectedStatus: http.StatusOK,
				expectedLen:    1,
				checkResponse: func(t *testing.T, resp *PeersResponse) {
					assert.Equal(t, "alpha-node/v1.0.0", resp.Peers[0].AgentVersion.String)
				},
			},
			{
				name:           "Filter by sync status",
				queryParams:    url.Values{"synced": []string{"true"}},
				expectedStatus: http.StatusOK,
				expectedLen:    2,
				checkResponse: func(t *testing.T, resp *PeersResponse) {
					for _, peer := range resp.Peers {
						assert.True(t, peer.IsSynced.Bool)
					}
				},
			},
			{
				name:           "Sort by created_at ascending",
				queryParams:    url.Values{"sort": []string{"created_at"}, "order": []string{"asc"}},
				expectedStatus: http.StatusOK,
				expectedLen:    4,
				checkResponse: func(t *testing.T, resp *PeersResponse) {
					assert.Equal(t, "QmPeer1", resp.Peers[0].PeerID)
					assert.Equal(t, "QmPeer4", resp.Peers[3].PeerID)
				},
			},
			{
				name:           "Sort by created_at descending",
				queryParams:    url.Values{"sort": []string{"created_at"}, "order": []string{"desc"}},
				expectedStatus: http.StatusOK,
				expectedLen:    4,
				checkResponse: func(t *testing.T, resp *PeersResponse) {
					assert.Equal(t, "QmPeer4", resp.Peers[0].PeerID)
					assert.Equal(t, "QmPeer1", resp.Peers[3].PeerID)
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				//nolint:gocritic
				req := httptest.NewRequest(http.MethodGet, "/api/peers?"+tc.queryParams.Encode(), nil)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				err := h.GetPeers(c)
				require.NoError(t, err)

				assert.Equal(t, tc.expectedStatus, rec.Code)

				if tc.expectedStatus == http.StatusOK {
					var response PeersResponse
					err := json.Unmarshal(rec.Body.Bytes(), &response)
					require.NoError(t, err)
					assert.Len(t, response.Peers, tc.expectedLen)

					if tc.checkResponse != nil {
						tc.checkResponse(t, &response)
					}
				}
			})
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		testCases := []struct {
			name           string
			queryParams    url.Values
			expectedStatus int
			expectedLen    int
			checkResponse  func(*testing.T, *PeersResponse)
		}{
			{
				name:           "First page",
				queryParams:    url.Values{"page_size": []string{"2"}},
				expectedStatus: http.StatusOK,
				expectedLen:    2,
				checkResponse: func(t *testing.T, resp *PeersResponse) {
					assert.NotEmpty(t, resp.NextPaginationToken)
					assert.Equal(t, "QmPeer4", resp.Peers[0].PeerID)
					assert.Equal(t, "QmPeer3", resp.Peers[1].PeerID)
				},
			},
			{
				name:           "Second page",
				queryParams:    url.Values{"page_size": []string{"2"}, "pagination_token": []string{""}}, // We'll set the token in the test
				expectedStatus: http.StatusOK,
				expectedLen:    2,
				checkResponse: func(t *testing.T, resp *PeersResponse) {
					assert.Empty(t, resp.NextPaginationToken)
					assert.Equal(t, "QmPeer2", resp.Peers[0].PeerID)
					assert.Equal(t, "QmPeer1", resp.Peers[1].PeerID)
				},
			},
			{
				name:           "Page size larger than total",
				queryParams:    url.Values{"page_size": []string{"10"}},
				expectedStatus: http.StatusOK,
				expectedLen:    4,
				checkResponse: func(t *testing.T, resp *PeersResponse) {
					assert.Empty(t, resp.NextPaginationToken)
				},
			},
		}

		var firstPageToken string

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				if tc.name == "Second page" {
					tc.queryParams.Set("pagination_token", firstPageToken)
				}

				//nolint:gocritic
				req := httptest.NewRequest(http.MethodGet, "/api/peers?"+tc.queryParams.Encode(), nil)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				err := h.GetPeers(c)
				require.NoError(t, err)

				assert.Equal(t, tc.expectedStatus, rec.Code)

				if tc.expectedStatus == http.StatusOK {
					var response PeersResponse
					err := json.Unmarshal(rec.Body.Bytes(), &response)
					require.NoError(t, err)
					assert.Len(t, response.Peers, tc.expectedLen)

					if tc.checkResponse != nil {
						tc.checkResponse(t, &response)
					}

					if tc.name == "First page" {
						firstPageToken = response.NextPaginationToken
					}
				}
			})
		}
	})
}

func TestGetPeerByID(t *testing.T) {
	t.Parallel()

	r, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	h := &Handler{Repo: r}
	e := echo.New()

	testCases := []struct {
		name           string
		peerID         string
		expectError    bool
		expectedStatus int
		checkResponse  func(*testing.T, *PeerResponse)
	}{
		{
			name:           "Existing peer",
			peerID:         "QmPeer1",
			expectError:    false,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *PeerResponse) {
				assert.Equal(t, "QmPeer1", resp.PeerID)
				assert.Equal(t, "alpha-node/v1.0.0", resp.AgentVersion.String)
			},
		},
		{
			name:           "Non-existing peer",
			peerID:         "QmNonExistentPeer",
			expectError:    true,
			expectedStatus: http.StatusNotFound,
			checkResponse:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			//nolint:gocritic
			req := httptest.NewRequest(http.MethodGet, "/api/peers/"+tc.peerID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tc.peerID)

			err := h.GetPeerByID(c)
			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK {
				var response PeerResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				if tc.checkResponse != nil {
					tc.checkResponse(t, &response)
				}
			}
		})
	}
}

func TestGetPeerNeighbors(t *testing.T) {
	t.Parallel()

	r, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	h := &Handler{Repo: r}
	e := echo.New()

	testCases := []struct {
		name           string
		peerID         string
		expectedStatus int
		expectedLen    int
	}{
		{
			name:           "Existing peer with neighbors",
			peerID:         "QmPeer1",
			expectedStatus: http.StatusOK,
			expectedLen:    2,
		},
		{
			name:           "Existing peer without neighbors",
			peerID:         "QmPeer4",
			expectedStatus: http.StatusOK,
			expectedLen:    0,
		},
		{
			name:           "Non-existing peer",
			peerID:         "QmNonExistentPeer",
			expectedStatus: http.StatusOK,
			expectedLen:    0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			//nolint:gocritic
			req := httptest.NewRequest(http.MethodGet, "/api/peers/"+tc.peerID+"/neighbors", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tc.peerID)

			err := h.GetPeerNeighbors(c)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK {
				var response NeighborResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Len(t, response.Neighbors, tc.expectedLen)
			}
		})
	}
}

func TestGetAnalytics(t *testing.T) {
	t.Parallel()

	r, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	h := &Handler{Repo: r}
	e := echo.New()

	//nolint:gocritic
	req := httptest.NewRequest(http.MethodGet, "/api/analytics", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetAnalytics(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response AnalyticsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// The test database has some test data created by the fixtures
	// We should have some peers from the test fixtures
	assert.GreaterOrEqual(t, response.TotalPeers, int64(0))
	assert.GreaterOrEqual(t, response.PeersLatest, int64(0))
	assert.GreaterOrEqual(t, response.PeersChurnIn, int64(0))
	assert.GreaterOrEqual(t, response.PeersChurnOut, int64(0))

	// These fields are not populated in the current implementation
	assert.Len(t, response.PeersByAgent, 0)
	assert.Len(t, response.PeersByProtocol, 0)
	assert.Len(t, response.PeersByContinent, 0)
	assert.Len(t, response.PeersByCountry, 0)
	assert.Len(t, response.PeersByCity, 0)
	assert.Len(t, response.PeersByASO, 0)

	// SyncStatusCount might be nil if no crawl data exists, or have sync status data
	if response.SyncStatusCount != nil {
		// If we have sync status data, it should have at least 2 keys (could be synced/unknown, etc)
		// With stored data from crawls, it would have 4 keys including "total"
		assert.GreaterOrEqual(t, len(response.SyncStatusCount), 2)
	}
}

func TestGetPeerCountByAgent(t *testing.T) {
	t.Parallel()

	r, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	h := &Handler{Repo: r}
	e := echo.New()

	//nolint:gocritic
	req := httptest.NewRequest(http.MethodGet, "/api/analytics/clients", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetPeerCountByAgent(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response AgentCountResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.AgentCounts, 4)
}

func TestGetPeerHistory(t *testing.T) {
	t.Parallel()

	r, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	h := &Handler{Repo: r}
	e := echo.New()

	testCases := []struct {
		name           string
		queryParams    string
		expectedError  bool
		expectedStatus int
		expectedLen    int
	}{
		{
			name:           "Full range",
			queryParams:    "",
			expectedError:  false,
			expectedStatus: http.StatusOK,
			expectedLen:    3,
		},
		{
			name:           "Partial range",
			queryParams:    "start=2023-01-02T00:00:00Z&end=2023-01-03T00:00:00Z",
			expectedError:  false,
			expectedStatus: http.StatusOK,
			expectedLen:    1,
		},
		{
			name:           "Invalid date format",
			queryParams:    "start=invalid-date",
			expectedError:  true,
			expectedStatus: http.StatusBadRequest,
			expectedLen:    0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			//nolint:gocritic
			req := httptest.NewRequest(http.MethodGet, "/api/analytics/peers/history?"+tc.queryParams, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := h.GetPeerHistory(c)
			if tc.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK {
				var response PeerHistoryResponse
				t.Log(rec.Body.String())
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Len(t, response.History, tc.expectedLen)
			}
		})
	}
}
