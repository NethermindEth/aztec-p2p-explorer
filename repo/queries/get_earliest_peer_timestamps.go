package queries

// GetEarliestPeerTimestamps retrieves the earliest timestamps for a batch of peers,
// checking visits, peer_visits_index, and falling back to peers.created_at.
// Parameters:
//
//	$1 - batchSize (int): Number of peers to retrieve
//	$2 - offset (int): Offset for pagination
const GetEarliestPeerTimestamps = `
	WITH peer_list AS (
		SELECT id, multi_hash, created_at
		FROM peers
		ORDER BY id
		LIMIT $1 OFFSET $2
	),
	earliest_visits AS (
		SELECT
			p.multi_hash,
			MIN(v.visit_started_at) as first_seen
		FROM peer_list p
		INNER JOIN visits v ON p.id = v.peer_id
		GROUP BY p.multi_hash
	),
	earliest_index AS (
		SELECT
			pvi.peer_multi_hash as multi_hash,
			MIN(pvi.created_at) as first_seen
		FROM peer_list p
		INNER JOIN peer_visits_index pvi ON p.multi_hash = pvi.peer_multi_hash
		WHERE p.multi_hash NOT IN (SELECT multi_hash FROM earliest_visits)
		GROUP BY pvi.peer_multi_hash
	)
	SELECT
		p.multi_hash,
		COALESCE(
			ev.first_seen,
			ei.first_seen,
			p.created_at
		) as first_seen,
		CASE
			WHEN ev.first_seen IS NOT NULL THEN 'visits'
			WHEN ei.first_seen IS NOT NULL THEN 'peer_visits_index'
			ELSE 'peers'
		END as source
	FROM peer_list p
	LEFT JOIN earliest_visits ev ON p.multi_hash = ev.multi_hash
	LEFT JOIN earliest_index ei ON p.multi_hash = ei.multi_hash
	ORDER BY p.id
`
