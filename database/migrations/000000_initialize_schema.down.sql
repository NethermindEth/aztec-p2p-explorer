BEGIN;

DROP TABLE IF EXISTS visits;
DROP TABLE IF EXISTS peers_x_multi_addresses;
DROP TABLE IF EXISTS neighbors;
DROP TABLE IF EXISTS peers_states;
DROP TABLE IF EXISTS peers;
DROP TABLE IF EXISTS protocols_sets;
DROP TABLE IF EXISTS protocols;
DROP TABLE IF EXISTS multi_addresses_x_ip_addresses;
DROP TABLE IF EXISTS multi_addresses;
DROP TABLE IF EXISTS ip_addresses;
DROP TABLE IF EXISTS cities;
DROP TABLE IF EXISTS countries;
DROP TABLE IF EXISTS continents;
DROP TABLE IF EXISTS autonomous_systems;
DROP TABLE IF EXISTS agent_versions;
DROP TABLE IF EXISTS crawls;

COMMIT;