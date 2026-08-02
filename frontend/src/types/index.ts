export type HistoryData = {
  date: string;
  peer_count: number;
};

export type HistoryResponse = {
  history: HistoryData[];
};

type ClientVersionData = Record<'synced' | 'unsynced', number>;

type ClientVersions = Record<string, ClientVersionData>;
export type ClientVersionInfo = {
  total: number;
  versions: ClientVersions;
};
// example: "alpha-node": {"1.0.0": {"synced": 1, "unsynced": 0}, "1.0.1": {"synced": 0, "unsynced": 0}}
export type Clients = Record<string, ClientVersionInfo>;
// TODO the SyncStatus should be handled better, this is a bit odd
export type SyncStatus = {
  not_synced?: number;
  synced?: number;
  unknown?: number;
};

export type SyncStatusResponse = {
  sync_status: SyncStatus;
};

// export type Continents = Record<
//   'Africa' | 'Antarctica' | 'Asia' | 'Europe' | 'North America' | 'Australia' | 'South America',
//   number
// >;
// TODO: Would be better to use an object like above but will require a few more changes on data handling side
export type Continent = {
  continent_name: string;
  continent_code: string;
  count: number;
};

export type ContinentsResponse = {
  continents: Continent[];
};

// TODO both of these types can be clearer with an example
export type NetworkItems = {
  [network: string]: ClientInfo[];
};
export type Countries = {
  [country: string]: ClientInfo[];
};

type IpInfo = {
  ip_address: string;
  port: number;
  as_name: string;
  as_number: number;
  city_name: string;
  country_name: string;
  country_iso: string;
  continent_name: string;
  continent_code: string;
  latitude: number;
  longitude: number;
};

type MultiAddress = {
  maddr: string;
  ip_info: IpInfo[];
};

export type Peer = {
  id: string;
  created_at: string;
  last_seen: string;
  client: string;
  multi_addresses: MultiAddress[];
  protocols: string[];
  block_height: number;
  spec_version: string;
  // null when the backend could not determine sync state (e.g. reference RPC
  // unreachable). The UI renders this as "Unknown" rather than "Syncing".
  is_synced: boolean | null;
};

export type PeersResponse = {
  peers: Peer[];
  next_pagination_token: string;
  total_peers: number;
};

export type HoveredMarker = {
  data: Peer[];
  x: number;
  y: number;
};

export type ClientInfo = {
  clientName: string;
  clientVersion: string;
  isSynced: boolean;
};

export interface FilterValues {
  client: string[];
  sync: string[];
}

type ClientVersion = {
  version: string;
  count: number;
};

export type ClientAnalytics = {
  client_name: string;
  total_count: number;
  versions: ClientVersion[];
};

type ProtocolAnalytics = {
  protocol: string;
  count: number;
};

type LocationAnalytics = {
  continent_name?: string;
  continent_code?: string;
  country_name?: string;
  country_code?: string;
  city_name?: string;
  count: number;
};

type NetworkAnalytics = {
  as_name: string;
  as_number: number;
  count: number;
};

export type AnalyticsResponse = {
  total_peers: number;
  peers_latest: number;
  peers_churn_in: number;
  peers_churn_out: number;
  clients: ClientAnalytics[];
  protocols: ProtocolAnalytics[];
  continents: LocationAnalytics[];
  countries: LocationAnalytics[];
  cities: LocationAnalytics[];
  networks: NetworkAnalytics[];
  sync_status: {
    synced: number;
    not_synced: number;
    unknown: number;
  };
};

export type AnalyticsPeers = {
  peers_churn_in: 0;
  peers_churn_out: 0;
  peers_latest: 0;
  total_peers: 0;
};

export type AnalyticsCountry = {
  count: number;
  country_code: string;
  country_name: string;
};
export type AnalyticsCountries = {
  countries: AnalyticsCountry[];
};

type AnalyticsContinent = {
  count: number;
  continent_code: string;
  continent_name: string;
};
export type AnalyticsContinents = {
  continents: AnalyticsContinent[];
};

export type PeersMapResponse = {
  clusters: {
    city: string;
    continent: string;
    count: number;
    country: string;
    latitude: number;
    longitude: number;
    peer_ids: string[];
  }[];
  active_peers: number;
  total_peers: number;
};

export type NodeModal = {
  isOpen: boolean;
  nodeId: string;
  node: Peer | null;
};
