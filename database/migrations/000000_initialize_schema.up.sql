BEGIN;

-- Holds the state and aggregated results of one particular crawl.
CREATE TABLE crawls
(
    -- A unique id that identifies a crawl instance.
    id            INT GENERATED ALWAYS AS IDENTITY,
    -- Timestamp of when this crawl was started.
    started_at    TIMESTAMPTZ NOT NULL,
    -- Timestamp of when this crawl was finished.
    finished_at  TIMESTAMPTZ NOT NULL,
    -- The number of peers that were crawled.
    crawled_peers INTEGER,
    -- The number of peers that were successfully dialed.
    dialable_peers INTEGER,
    -- The number of peers that could not be reached.
    undialable_peers INTEGER,
    -- The remaining peers that have yet to be dialed.
    remaining_peers INTEGER,
    -- The version of nebula that was used to perform the crawl.
    version TEXT,
    -- The state of the crawl.
    state TEXT,

    PRIMARY KEY (id)
);

COMMENT ON TABLE crawls IS 'Holds the state and aggregated results of one particular crawl.';
COMMENT ON COLUMN crawls.id IS 'A unique id that identifies a crawl instance.';
COMMENT ON COLUMN crawls.started_at IS 'Timestamp of when this crawl was started.';
COMMENT ON COLUMN crawls.finished_at IS 'Timestamp of when this crawl was finished.';
COMMENT ON COLUMN crawls.crawled_peers IS 'The number of peers that were crawled.';
COMMENT ON COLUMN crawls.dialable_peers IS 'The number of peers that were successfully dialed.';
COMMENT ON COLUMN crawls.undialable_peers IS 'The number of peers that could not be reached.';
COMMENT ON COLUMN crawls.remaining_peers IS 'The remaining peers that have yet to be dialed.';
COMMENT ON COLUMN crawls.version IS 'The version of nebula that was used to perform the crawl.';
COMMENT ON COLUMN crawls.state IS 'The state of the crawl.';

CREATE INDEX idx_crawls_started_at ON crawls(started_at);

-- Holds all discovered agent_versions
CREATE TABLE agent_versions
(
    -- A unique id that identifies a agent version.
    id            INT GENERATED ALWAYS AS IDENTITY,
    -- Timestamp of when this agent version was seen the last time.
    created_at    TIMESTAMPTZ NOT NULL,
    -- Agent version string as reported from the remote peer.
    agent_version TEXT        NOT NULL CHECK ( TRIM(agent_version) != '' ),
    -- The count of the agent version.
    count INT NOT NULL,
    -- Timestamp of when this agent version was updated.
    updated_at TIMESTAMPTZ NOT NULL,

    -- There should only be unique agent version strings in this table.
    CONSTRAINT uq_agent_versions_agent_version UNIQUE (agent_version),

    PRIMARY KEY (id)
);

COMMENT ON TABLE agent_versions IS 'Holds all discovered agent_versions';
COMMENT ON COLUMN agent_versions.id IS 'A unique id that identifies a agent version.';
COMMENT ON COLUMN agent_versions.created_at IS 'Timestamp of when this agent version was seen the last time.';
COMMENT ON COLUMN agent_versions.agent_version IS 'Agent version string as reported from the remote peer.';
COMMENT ON COLUMN agent_versions.count IS 'The count of the agent version.';
COMMENT ON COLUMN agent_versions.updated_at IS 'Timestamp of when this agent version was updated.';

-- Holds all the different autonomous systems
CREATE TABLE autonomous_systems
(
    -- A unique id that identifies an autonomous system.
    id         INT GENERATED ALWAYS AS IDENTITY,
    -- The autonomous system number.
    as_number     INT NOT NULL,
    -- The autonomous system organisation name.
    as_name   TEXT        NOT NULL CHECK ( TRIM(as_name) != '' ),

    -- There should only be unique protocol strings in this table
    CONSTRAINT uq_autonomous_system_number UNIQUE (as_number),

    PRIMARY KEY (id)
);

COMMENT ON TABLE autonomous_systems IS 'Holds all the different autonomous systems.';
COMMENT ON COLUMN autonomous_systems.id IS 'A unique id that identifies an autonomous system.';
COMMENT ON COLUMN autonomous_systems.as_number IS 'The autonomous system number.';
COMMENT ON COLUMN autonomous_systems.as_name IS 'The autonomous system organisation name.';

-- Holds all the different continents
CREATE TABLE continents
(
    -- A unique id that identifies a continent.
    id         INT GENERATED ALWAYS AS IDENTITY,
    -- The code of the continent, consisting of two characters.
    code          CHAR(2) NOT NULL CHECK ( TRIM(code) != '' ),
    -- The continent name.
    continent_name   TEXT NOT NULL CHECK ( TRIM(continent_name) != '' ),

    PRIMARY KEY (id),

    -- Each code should be unique
    CONSTRAINT uq_continents_code UNIQUE (code)
);

COMMENT ON TABLE continents IS 'Holds all the different continents.';
COMMENT ON COLUMN continents.id IS 'A unique id that identifies a continent.';
COMMENT ON COLUMN continents.code IS 'The code of the continent, consisting of two characters.';
COMMENT ON COLUMN continents.continent_name IS 'The continent name.';

-- Holds all the different countries
CREATE TABLE countries
(
    -- A unique id that identifies a country.
    id         INT GENERATED ALWAYS AS IDENTITY,
    -- The ISO code of the country, consisting of two or three characters.
    iso_code          CHAR(3) NOT NULL CHECK ( TRIM(iso_code) != '' ),
    -- The country name.
    country_name   TEXT        NOT NULL CHECK ( TRIM(country_name) != '' ),
    -- The continent id of the country.
    continent_id INT,

    PRIMARY KEY (id),
    FOREIGN KEY (continent_id) REFERENCES continents (id),

    -- Each code should be unique
    CONSTRAINT uq_countries_iso_code UNIQUE (iso_code)
);

COMMENT ON TABLE countries IS 'Holds all the different countries.';
COMMENT ON COLUMN countries.id IS 'A unique id that identifies an autonomous system.';
COMMENT ON COLUMN countries.iso_code IS 'The ISO code of the country, consisting of two or three characters.';
COMMENT ON COLUMN countries.country_name IS 'The country name.';
COMMENT ON COLUMN countries.continent_id IS 'The continent id of the country.';

-- Holds all the different cities
CREATE TABLE cities
(
    -- A unique id that identifies a city.
    id         INT GENERATED ALWAYS AS IDENTITY,
    -- The city name.
    city_name   TEXT        NOT NULL CHECK ( TRIM(city_name) != '' ),
    -- The country id of the city.
    country_id  INT,
    -- The continent id of the city.
    continent_id INT,

    PRIMARY KEY (id),
    FOREIGN KEY (country_id) REFERENCES countries (id),
    FOREIGN KEY (continent_id) REFERENCES continents (id),

    -- Each city should be unique
    CONSTRAINT uq_cities_city_name_country_id_continent_id UNIQUE (city_name, country_id, continent_id)
);

COMMENT ON TABLE cities IS 'Holds all the different cities.';
COMMENT ON COLUMN cities.id IS 'A unique id that identifies a city.';
COMMENT ON COLUMN cities.city_name IS 'The city name.';
COMMENT ON COLUMN cities.country_id IS 'The country id of the city.';
COMMENT ON COLUMN cities.continent_id IS 'The continent id of the city.';

-- Holds all the different IP addresses that were derived from a multi address
CREATE TABLE ip_addresses
(
    -- A unique id that identifies this ip address.
    id               INT GENERATED ALWAYS AS IDENTITY,
    -- The country that the IP address belongs to.
    country_id      INT,
    -- The continent that the IP address belongs to.
    continent_id    INT,
    -- The city that the IP address belongs to.
    city_id         INT,
    -- The autonomous system that the IP address belongs to.
    as_id          INT,
    -- When was this IP address created.
    created_at     TIMESTAMPTZ NOT NULL,
    -- When was this IP address last updated.
    updated_at     TIMESTAMPTZ NOT NULL,
    -- The latitude of the IP address.
    latitude       DOUBLE PRECISION,
    -- The longitude of the IP address.
    longitude      DOUBLE PRECISION,
    -- The IP address derived from a multi address.
    ip_address     INET NOT NULL,
    -- The port used to connect to the IP address.
    port           INT NOT NULL,

    -- There should only be unique IP address and port combinations in this table
    CONSTRAINT uq_ip_addresses_address UNIQUE (ip_address, port),

    PRIMARY KEY (id),
    FOREIGN KEY (country_id) REFERENCES countries (id),
    FOREIGN KEY (continent_id) REFERENCES continents (id),
    FOREIGN KEY (city_id) REFERENCES cities (id),
    FOREIGN KEY (as_id) REFERENCES autonomous_systems (id)
);

COMMENT ON TABLE ip_addresses IS 'Holds all the different IP addresses that were derived from a multi address.';
COMMENT ON COLUMN ip_addresses.id IS 'A unique id that identifies this ip address.';
COMMENT ON COLUMN ip_addresses.country_id IS 'The country that the IP address belongs to.';
COMMENT ON COLUMN ip_addresses.continent_id IS 'The continent that the IP address belongs to.';
COMMENT ON COLUMN ip_addresses.city_id IS 'The city that the IP address belongs to.';
COMMENT ON COLUMN ip_addresses.as_id IS 'The autonomous system that the IP address belongs to.';
COMMENT ON COLUMN ip_addresses.created_at IS 'When was this IP address created.';
COMMENT ON COLUMN ip_addresses.updated_at IS 'When was this IP address last updated.';
COMMENT ON COLUMN ip_addresses.latitude IS 'The latitude of the IP address.';
COMMENT ON COLUMN ip_addresses.longitude IS 'The longitude of the IP address.';
COMMENT ON COLUMN ip_addresses.ip_address IS 'The IP address derived from a multi address.';
COMMENT ON COLUMN ip_addresses.port IS 'The port used to connect to the IP address.';

-- The `multi_addresses` table keeps track of all ever encountered multi addresses.
CREATE TABLE multi_addresses
(
    -- An internal unique id that identifies this multi address.
    id             INT GENERATED ALWAYS AS IDENTITY,
    -- The multi address in the form of `/ip4/123.456.789.123/tcp/4001`.
    maddr          TEXT        NOT NULL CHECK ( TRIM(maddr) != '' ),
    -- When was this multi address created
    created_at     TIMESTAMPTZ NOT NULL,

    -- There should only ever be distinct multi addresses here
    CONSTRAINT uq_multi_addresses_address UNIQUE (maddr),

    PRIMARY KEY (id)
);

COMMENT ON TABLE multi_addresses IS ''
    'The `multi_addresses` table keeps track of all ever encountered multi addresses.'
    'Some of these multi addresses can be associated with additional information.';
COMMENT ON COLUMN multi_addresses.id IS 'An internal unique id that identifies this multi address.';
COMMENT ON COLUMN multi_addresses.maddr IS 'The multi address in the form of `/ip4/123.456.789.123/tcp/4001`.';
COMMENT ON COLUMN multi_addresses.created_at IS 'Timestamp of when this multi address was created.';

-- Keeps track of the association of multi addresses to ip_addresses.
-- A multi address can be associated with multiple IP addresses, could happen due to dnsaddr multi addresses.
-- An IP address can be associated with multiple multi addresses as well, could happen due to NAT.
CREATE TABLE multi_addresses_x_ip_addresses
(
    -- The ID of the multi address that has been seen for the above IP address
    multi_address_id INT NOT NULL,
    -- The ID of the IP address that has been seen for the above multi address
    ip_address_id    INT NOT NULL,

    -- The maddr ID should always point to an existing multi address in the DB
    CONSTRAINT fk_multi_addresses_x_ip_addresses_multi_address_id FOREIGN KEY (multi_address_id) REFERENCES multi_addresses (id) ON DELETE CASCADE,
    -- The IP address ID should always point to an existing IP address in the DB
    CONSTRAINT fk_multi_addresses_x_ip_addresses_ip_address_id FOREIGN KEY (ip_address_id) REFERENCES ip_addresses (id) ON DELETE CASCADE,

    PRIMARY KEY (multi_address_id, ip_address_id)
);

CREATE INDEX idx_multi_addresses_x_ip_addresses_multi_address_id ON multi_addresses_x_ip_addresses (multi_address_id);

COMMENT ON TABLE multi_addresses_x_ip_addresses IS 'Keeps track of the association of multi addresses to ip_addresses.';
COMMENT ON COLUMN multi_addresses_x_ip_addresses.multi_address_id IS 'The ID of the multi address that has been seen for the above IP address.';
COMMENT ON COLUMN multi_addresses_x_ip_addresses.ip_address_id IS 'The ID of the IP address that has been seen for the above multi address.';

-- Holds all the different protocols that the crawler came across
CREATE TABLE protocols
(
    -- A unique id that identifies a protocol.
    -- Note: there's an invariant regarding the INT type. Don't increase it to BIGINT without changing protocolsSetHash.
    id         INT GENERATED ALWAYS AS IDENTITY,
    -- Timestamp of when this protocol was seen the last time.
    created_at TIMESTAMPTZ NOT NULL,
    -- The full protocol string.
    protocol   TEXT        NOT NULL CHECK ( TRIM(protocol) != '' ),
    -- The count of the protocol.
    count      INT NOT NULL,
    -- Timestamp of when this protocol was updated.
    updated_at TIMESTAMPTZ NOT NULL,

    -- There should only be unique protocol strings in this table
    CONSTRAINT uq_protocols_protocol UNIQUE (protocol),

    PRIMARY KEY (id)
);

COMMENT ON TABLE protocols IS 'Holds all the different protocols that the crawler came across.';
COMMENT ON COLUMN protocols.id IS 'A unique id that identifies a agent version.';
COMMENT ON COLUMN protocols.created_at IS 'Timestamp of when this protocol was seen the last time.';
COMMENT ON COLUMN protocols.protocol IS 'The full protocol string.';
COMMENT ON COLUMN protocols.count IS 'The count of the protocol.';
COMMENT ON COLUMN protocols.updated_at IS 'Timestamp of when this protocol was updated.';

-- Activate intarray extension for efficient array operations
CREATE EXTENSION IF NOT EXISTS intarray;

-- Since the set of protocols for a particular peer does not change very often in between crawls,
-- this table holds particular sets of protocols which other tables can reference and save space.
CREATE TABLE protocols_sets
(
    -- An internal unique id that identifies a unique set of protocols.
    -- We could also just use the hash below but since protocol sets are
    -- referenced many times having just a 4 byte instead of 32 byte ID
    -- can make huge storage difference.
    id           INT GENERATED ALWAYS AS IDENTITY,
    -- The protocol IDs of this protocol set. The IDs reference the protocols table (no foreign key checks).
    -- Note: there's an invariant regarding the INT type. Don't increase it to BIGINT without changing protocolsSetHash.
    protocol_ids INT[] NOT NULL CHECK ( array_length(protocol_ids, 1) IS NOT NULL ),
    -- The hash digest of the sorted protocol ids to allow a unique constraint
    hash         BYTEA NOT NULL,

    CONSTRAINT uq_protocols_sets_hash UNIQUE (hash),

    PRIMARY KEY (id)
);

CREATE INDEX idx_protocols_sets_protocol_ids on protocols_sets USING GIST (protocol_ids);

COMMENT ON TABLE protocols_sets IS ''
    'Since the set of protocols for a particular peer does not change very often in between crawls,'
    'this table holds particular sets of protocols which other tables can reference and save space.';
COMMENT ON COLUMN protocols_sets.id IS 'An internal unique id that identifies a unique set of protocols.';
COMMENT ON COLUMN protocols_sets.protocol_ids IS 'The protocol IDs of this protocol set. The IDs reference the protocols table (no foreign key checks).';
COMMENT ON COLUMN protocols_sets.hash IS 'The hash digest of the sorted protocol ids to allow a unique constraint.';

-- Keep track of all the peers that we have seen in the network
CREATE TABLE peers
(
    -- A unique id that identifies this peer.
    id               INT GENERATED ALWAYS AS IDENTITY,
    -- The agent version of the peer at the time of the last update.
    agent_version_id INT,
    -- The set of protocols that the peer supports at the time of the last update.
    protocols_set_id INT,
    -- The peer ID of the peer defined by the libP2P spec, which is the hash of the peer's public key.
    multi_hash       TEXT NOT NULL CHECK ( TRIM(multi_hash) != '' ),
    -- When was the peer created.
    created_at      TIMESTAMPTZ NOT NULL,
    -- When was the agent version or protocol set of this peer updated.
    updated_at      TIMESTAMPTZ NOT NULL,
    -- When was the peer last seen.
    last_seen       TIMESTAMPTZ NOT NULL,

    CONSTRAINT fk_peers_agent_version_id FOREIGN KEY (agent_version_id) REFERENCES agent_versions (id) ON DELETE SET NULL,
    CONSTRAINT fk_peers_protocols_set_id FOREIGN KEY (protocols_set_id) REFERENCES protocols_sets (id) ON DELETE SET NULL,

    -- There should only be unique multi hashes in this table.
    CONSTRAINT uq_peers_multi_hash UNIQUE (multi_hash),

    PRIMARY KEY (id)
);

COMMENT ON TABLE peers IS 'Keep track of all the peers that we have seen in the network.';
COMMENT ON COLUMN peers.id IS 'A unique id that identifies this peer.';
COMMENT ON COLUMN peers.agent_version_id IS 'The agent version of the peer at the time of the last update.';
COMMENT ON COLUMN peers.protocols_set_id IS 'The set of protocols that the peer supports at the time of the last update.';
COMMENT ON COLUMN peers.multi_hash IS 'The peer ID of the peer defined by the libP2P spec, which is the hash of the peer''s public key.';
COMMENT ON COLUMN peers.created_at IS 'When was the peer created.';
COMMENT ON COLUMN peers.updated_at IS 'When was the agent version or protocol set of this peer updated.';
COMMENT ON COLUMN peers.last_seen IS 'When was the peer last seen.';

-- Keep track of the blockchain states of the peers
CREATE TABLE peers_states
(
    -- A unique id that identifies this peer.
    id               INT GENERATED ALWAYS AS IDENTITY,
    -- The peer id of the peer.
    peer_id         INT NOT NULL,
    -- The block height of the peer.
    block_height   BIGINT,
    -- The JSON-RPC specification version of the peer.
    spec_version  TEXT CHECK ( TRIM(spec_version) != '' ),
    -- Whether the peer is synced when the record is made.
    is_synced     BOOLEAN,
    -- The time when the record was created.
    created_at      TIMESTAMPTZ NOT NULL,

    CONSTRAINT fk_peers_states_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,

    PRIMARY KEY (id)
);

COMMENT ON TABLE peers_states IS 'Keep track of the blockchain states of the peers.';
COMMENT ON COLUMN peers_states.id IS 'A unique id that identifies this peer.';
COMMENT ON COLUMN peers_states.peer_id IS 'The peer id of the peer.';
COMMENT ON COLUMN peers_states.block_height IS 'The block height of the peer.';
COMMENT ON COLUMN peers_states.spec_version IS 'The JSON-RPC specification version of the peer.';
COMMENT ON COLUMN peers_states.is_synced IS 'Whether the peer is synced when the record is made.';
COMMENT ON COLUMN peers_states.created_at IS 'The time when the record was created.';

-- Activate intarray extension for efficient array operations
CREATE EXTENSION IF NOT EXISTS intarray;

-- Keep track of the neighbors of the peers
CREATE TABLE neighbors
(
    -- The peer and their neighbor (entry in its k-buckets)
    peer_id      INT,
    -- The peers to which the peer above is connected
    neighbor_ids INT[],
    -- The time when the neighbors were last updated
    updated_at  TIMESTAMPTZ NOT NULL,

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_neighbors_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,

    PRIMARY KEY (peer_id)
);

COMMENT ON TABLE neighbors IS 'Keep track of the neighbors of the peers.';
COMMENT ON COLUMN neighbors.peer_id IS 'The peer and their neighbor (entry in its k-buckets).';
COMMENT ON COLUMN neighbors.neighbor_ids IS 'The peers to which the peer above is connected.';

-- Keeps track of the association of multi addresses to peers.
CREATE TABLE peers_x_multi_addresses
(
    -- The peer ID of which we want to track the multi address
    peer_id          INT NOT NULL,
    -- The ID of the multi address that has been seen for the above peer
    multi_address_id INT NOT NULL,

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_peers_x_multi_addresses_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,
    -- The maddr ID should always point to an existing multi address in the DB
    CONSTRAINT fk_peers_x_multi_addresses_multi_address_id FOREIGN KEY (multi_address_id) REFERENCES multi_addresses (id) ON DELETE CASCADE,

    PRIMARY KEY (peer_id, multi_address_id)
);

CREATE INDEX idx_peers_x_multi_addresses_peer_id ON peers_x_multi_addresses (peer_id);

COMMENT ON TABLE peers_x_multi_addresses IS 'Keeps track of the association of multi addresses to peers.';
COMMENT ON COLUMN peers_x_multi_addresses.peer_id IS 'The peer ID of which we want to track the multi address.';
COMMENT ON COLUMN peers_x_multi_addresses.multi_address_id IS 'The ID of the multi address that has been seen for the above peer.';

-- Stores the records of visiting peers.
CREATE TABLE visits
(
    -- A unique id that identifies this visit.
    id                INT GENERATED ALWAYS AS IDENTITY,
    -- Which peer did we meet.
    peer_id           INT         NOT NULL,
    -- The agent version that this peer reported during this visit.
    agent_version_id  INT,
    -- The set of supported protocols that this peer reported.
    protocols_set_id  INT,
    -- The error that happened for this visit.
    connect_error     TEXT,
    -- The error that happened when crawling the peer.
    crawl_error       TEXT,
    -- The time it took to connect with the peer.
    visit_started_at  TIMESTAMPTZ NOT NULL,
    -- The time it took to connect with the peer.
    visit_ended_at    TIMESTAMPTZ NOT NULL,
    -- The time it took to connect with the peer or until an error occurred.
    connect_duration  INTERVAL,
    -- The time it took to crawl the peer also if an error occurred.
    crawl_duration    INTERVAL,
    -- An array of all multi address IDs of the remote peer.
    multi_address_ids INT[],
    -- In which crawl did we meet this peer.
    crawl_id         INT,

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_visits_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,
    -- The agent version ID should always point to an existing agent version in the DB
    CONSTRAINT fk_visits_agent_version_id FOREIGN KEY (agent_version_id) REFERENCES agent_versions (id) ON DELETE SET NULL,
    -- The protocol set ID should always point to an existing protocol set in the DB
    CONSTRAINT fk_visits_protocols_set_id FOREIGN KEY (protocols_set_id) REFERENCES protocols_sets (id) ON DELETE SET NULL,
    -- The crawl ID should always point to an existing crawl in the DB
    CONSTRAINT fk_visits_crawl_id FOREIGN KEY (crawl_id) REFERENCES crawls (id) ON DELETE SET NULL,

    PRIMARY KEY (id, visit_started_at)
);

CREATE INDEX idx_visits_visit_started_at ON visits (visit_started_at);
CREATE INDEX idx_visits_peer_id ON visits (peer_id);

COMMENT ON TABLE visits IS 'Stores the records of visiting peers.';
COMMENT ON COLUMN visits.id IS 'A unique id that identifies this visit.';
COMMENT ON COLUMN visits.peer_id IS 'Which peer did we meet.';
COMMENT ON COLUMN visits.agent_version_id IS 'The agent version that this peer reported during this visit.';
COMMENT ON COLUMN visits.protocols_set_id IS 'The set of supported protocols that this peer reported.';
COMMENT ON COLUMN visits.connect_error IS 'The error that happened for this visit.';
COMMENT ON COLUMN visits.crawl_error IS 'The error that happened when crawling the peer.';
COMMENT ON COLUMN visits.visit_started_at IS 'The time it took to connect with the peer.';
COMMENT ON COLUMN visits.visit_ended_at IS 'The time it took to connect with the peer.';
COMMENT ON COLUMN visits.connect_duration IS 'The time it took to connect with the peer or until an error occurred.';
COMMENT ON COLUMN visits.crawl_duration IS 'The time it took to crawl the peer also if an error occurred.';
COMMENT ON COLUMN visits.multi_address_ids IS 'An array of all multi address IDs of the remote peer.';

COMMIT;
