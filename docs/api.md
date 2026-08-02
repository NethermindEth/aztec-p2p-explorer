


# Aztec P2P Explorer API
This is the aztec P2P Explorer API server.
  

## Informations

### Version

1

### Contact

aztec-p2p-explorer  http://github.com/NethermindEth/aztec-p2p-explorer

## Content negotiation

### URI Schemes
  * http

### Consumes
  * application/json

### Produces
  * application/json

## All endpoints

###  analytics

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/analytics | [get analytics](#get-analytics) | Get overall analytics |
| GET | /api/analytics/cities | [get analytics cities](#get-analytics-cities) | Get peer count by city |
| GET | /api/analytics/clients | [get analytics clients](#get-analytics-clients) | Get peer count by client names and versions |
| GET | /api/analytics/continents | [get analytics continents](#get-analytics-continents) | Get peer count by continent |
| GET | /api/analytics/countries | [get analytics countries](#get-analytics-countries) | Get peer count by country |
| GET | /api/analytics/networks | [get analytics networks](#get-analytics-networks) | Get peer count by ASO |
| GET | /api/analytics/peers | [get analytics peers](#get-analytics-peers) | Get total peer count and peers in the last crawl |
| GET | /api/analytics/peers/history | [get analytics peers history](#get-analytics-peers-history) | Get peer count history |
| GET | /api/analytics/protocols | [get analytics protocols](#get-analytics-protocols) | Get peer count by protocol |
  


###  apiops

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/health | [get health](#get-health) | Check health |
  


###  peers

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/peers | [get peers](#get-peers) | Get a list of peers |
| GET | /api/peers/{id} | [get peers ID](#get-peers-id) | Get a peer by ID |
| GET | /api/peers/{id}/neighbors | [get peers ID neighbors](#get-peers-id-neighbors) | Get peer neighbors |
| GET | /api/peers/map | [get peers map](#get-peers-map) | Get peer clusters for map visualisation |
| GET | /api/peers/neighbors | [get peers neighbors](#get-peers-neighbors) | Get every peer's neighbors |
  


## Paths

### <span id="get-analytics"></span> Get overall analytics (*GetAnalytics*)

```
GET /api/analytics
```

Get comprehensive analytics data about the network

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-analytics-200) | OK | OK |  | [schema](#get-analytics-200-schema) |
| [500](#get-analytics-500) | Internal Server Error | Internal Server Error |  | [schema](#get-analytics-500-schema) |

#### Responses


##### <span id="get-analytics-200"></span> 200 - OK
Status: OK

###### <span id="get-analytics-200-schema"></span> Schema
   
  

[ServerAnalyticsResponse](#server-analytics-response)

##### <span id="get-analytics-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-analytics-500-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

### <span id="get-analytics-cities"></span> Get peer count by city (*GetAnalyticsCities*)

```
GET /api/analytics/cities
```

Get the number of peers grouped by city

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-analytics-cities-200) | OK | OK |  | [schema](#get-analytics-cities-200-schema) |
| [500](#get-analytics-cities-500) | Internal Server Error | Internal Server Error |  | [schema](#get-analytics-cities-500-schema) |

#### Responses


##### <span id="get-analytics-cities-200"></span> 200 - OK
Status: OK

###### <span id="get-analytics-cities-200-schema"></span> Schema
   
  

[ServerCityCountResponse](#server-city-count-response)

##### <span id="get-analytics-cities-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-analytics-cities-500-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

### <span id="get-analytics-clients"></span> Get peer count by client names and versions (*GetAnalyticsClients*)

```
GET /api/analytics/clients
```

Get the number of peers grouped by client names and versions

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-analytics-clients-200) | OK | OK |  | [schema](#get-analytics-clients-200-schema) |
| [500](#get-analytics-clients-500) | Internal Server Error | Internal Server Error |  | [schema](#get-analytics-clients-500-schema) |

#### Responses


##### <span id="get-analytics-clients-200"></span> 200 - OK
Status: OK

###### <span id="get-analytics-clients-200-schema"></span> Schema
   
  

[ServerAgentCountResponse](#server-agent-count-response)

##### <span id="get-analytics-clients-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-analytics-clients-500-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

### <span id="get-analytics-continents"></span> Get peer count by continent (*GetAnalyticsContinents*)

```
GET /api/analytics/continents
```

Get the number of peers grouped by continent

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-analytics-continents-200) | OK | OK |  | [schema](#get-analytics-continents-200-schema) |
| [500](#get-analytics-continents-500) | Internal Server Error | Internal Server Error |  | [schema](#get-analytics-continents-500-schema) |

#### Responses


##### <span id="get-analytics-continents-200"></span> 200 - OK
Status: OK

###### <span id="get-analytics-continents-200-schema"></span> Schema
   
  

[ServerContinentCountResponse](#server-continent-count-response)

##### <span id="get-analytics-continents-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-analytics-continents-500-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

### <span id="get-analytics-countries"></span> Get peer count by country (*GetAnalyticsCountries*)

```
GET /api/analytics/countries
```

Get the number of peers grouped by country

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-analytics-countries-200) | OK | OK |  | [schema](#get-analytics-countries-200-schema) |
| [500](#get-analytics-countries-500) | Internal Server Error | Internal Server Error |  | [schema](#get-analytics-countries-500-schema) |

#### Responses


##### <span id="get-analytics-countries-200"></span> 200 - OK
Status: OK

###### <span id="get-analytics-countries-200-schema"></span> Schema
   
  

[ServerCountryCountResponse](#server-country-count-response)

##### <span id="get-analytics-countries-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-analytics-countries-500-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

### <span id="get-analytics-networks"></span> Get peer count by ASO (*GetAnalyticsNetworks*)

```
GET /api/analytics/networks
```

Get the number of peers grouped by Autonomous System Organization

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-analytics-networks-200) | OK | OK |  | [schema](#get-analytics-networks-200-schema) |
| [500](#get-analytics-networks-500) | Internal Server Error | Internal Server Error |  | [schema](#get-analytics-networks-500-schema) |

#### Responses


##### <span id="get-analytics-networks-200"></span> 200 - OK
Status: OK

###### <span id="get-analytics-networks-200-schema"></span> Schema
   
  

[ServerASCountResponse](#server-a-s-count-response)

##### <span id="get-analytics-networks-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-analytics-networks-500-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

### <span id="get-analytics-peers"></span> Get total peer count and peers in the last crawl (*GetAnalyticsPeers*)

```
GET /api/analytics/peers
```

Get the total number of peers in the network and the number of peers in the last crawl

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-analytics-peers-200) | OK | OK |  | [schema](#get-analytics-peers-200-schema) |
| [500](#get-analytics-peers-500) | Internal Server Error | Internal Server Error |  | [schema](#get-analytics-peers-500-schema) |

#### Responses


##### <span id="get-analytics-peers-200"></span> 200 - OK
Status: OK

###### <span id="get-analytics-peers-200-schema"></span> Schema
   
  

[ServerPeerCountResponse](#server-peer-count-response)

##### <span id="get-analytics-peers-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-analytics-peers-500-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

### <span id="get-analytics-peers-history"></span> Get peer count history (*GetAnalyticsPeersHistory*)

```
GET /api/analytics/peers/history
```

Get the history of peer counts over time (daily)

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| end | `query` | string | `string` |  |  |  | End date (RFC3339 format) |
| start | `query` | string | `string` |  |  |  | Start date (RFC3339 format) |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-analytics-peers-history-200) | OK | OK |  | [schema](#get-analytics-peers-history-200-schema) |
| [400](#get-analytics-peers-history-400) | Bad Request | Bad Request |  | [schema](#get-analytics-peers-history-400-schema) |
| [500](#get-analytics-peers-history-500) | Internal Server Error | Internal Server Error |  | [schema](#get-analytics-peers-history-500-schema) |

#### Responses


##### <span id="get-analytics-peers-history-200"></span> 200 - OK
Status: OK

###### <span id="get-analytics-peers-history-200-schema"></span> Schema
   
  

[ServerPeerHistoryResponse](#server-peer-history-response)

##### <span id="get-analytics-peers-history-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-analytics-peers-history-400-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

##### <span id="get-analytics-peers-history-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-analytics-peers-history-500-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

### <span id="get-analytics-protocols"></span> Get peer count by protocol (*GetAnalyticsProtocols*)

```
GET /api/analytics/protocols
```

Get the number of peers grouped by protocol

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-analytics-protocols-200) | OK | OK |  | [schema](#get-analytics-protocols-200-schema) |
| [500](#get-analytics-protocols-500) | Internal Server Error | Internal Server Error |  | [schema](#get-analytics-protocols-500-schema) |

#### Responses


##### <span id="get-analytics-protocols-200"></span> 200 - OK
Status: OK

###### <span id="get-analytics-protocols-200-schema"></span> Schema
   
  

[ServerProtocolCountResponse](#server-protocol-count-response)

##### <span id="get-analytics-protocols-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-analytics-protocols-500-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

### <span id="get-health"></span> Check health (*GetHealth*)

```
GET /api/health
```

Chech health

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-health-200) | OK | OK |  | [schema](#get-health-200-schema) |

#### Responses


##### <span id="get-health-200"></span> 200 - OK
Status: OK

###### <span id="get-health-200-schema"></span> Schema
   
  



### <span id="get-peers"></span> Get a list of peers (*GetPeers*)

```
GET /api/peers
```

Get a list of peers based on the provided query options

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| as_names | `query` | []string | `[]string` | `csv` |  |  | Names of Autonomous System Organizations |
| as_numbers | `query` | []string | `[]string` | `csv` |  |  | Numbers of Autonomous System Organizations |
| cities | `query` | []string | `[]string` | `csv` |  |  | City names |
| clients | `query` | []string | `[]string` | `csv` |  |  | Client names (e.g. aztec-node) |
| continents | `query` | []string | `[]string` | `csv` |  |  | Continent names |
| continents_code | `query` | []string | `[]string` | `csv` |  |  | Continent codes |
| countries | `query` | []string | `[]string` | `csv` |  |  | Country names |
| countries_codes | `query` | []string | `[]string` | `csv` |  |  | Country iso codes |
| id | `query` | string | `string` |  |  |  | Peer ID (e.g. 12...) |
| latest | `query` | boolean | `bool` |  |  |  | Include peers from latest crawl only (default: false, all crawls) |
| order | `query` | string | `string` |  |  |  | Sort order (asc or desc) |
| page_size | `query` | integer | `int64` |  |  |  | Number of items per page |
| pagination_token | `query` | string | `string` |  |  |  | Pagination token |
| sort | `query` | string | `string` |  |  |  | Sort column (created_at, last_seen, block_height) |
| synced | `query` | boolean | `bool` |  |  |  | Sync status |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-peers-200) | OK | OK |  | [schema](#get-peers-200-schema) |
| [400](#get-peers-400) | Bad Request | Bad Request |  | [schema](#get-peers-400-schema) |
| [500](#get-peers-500) | Internal Server Error | Internal Server Error |  | [schema](#get-peers-500-schema) |

#### Responses


##### <span id="get-peers-200"></span> 200 - OK
Status: OK

###### <span id="get-peers-200-schema"></span> Schema
   
  

[ServerPeersResponse](#server-peers-response)

##### <span id="get-peers-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-peers-400-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

##### <span id="get-peers-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-peers-500-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

### <span id="get-peers-id"></span> Get a peer by ID (*GetPeersID*)

```
GET /api/peers/{id}
```

Get detailed information about a specific peer

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | string | `string` |  | ✓ |  | Peer ID (e.g. 12...) |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-peers-id-200) | OK | OK |  | [schema](#get-peers-id-200-schema) |
| [404](#get-peers-id-404) | Not Found | Not Found |  | [schema](#get-peers-id-404-schema) |
| [500](#get-peers-id-500) | Internal Server Error | Internal Server Error |  | [schema](#get-peers-id-500-schema) |

#### Responses


##### <span id="get-peers-id-200"></span> 200 - OK
Status: OK

###### <span id="get-peers-id-200-schema"></span> Schema
   
  

[ServerPeerResponse](#server-peer-response)

##### <span id="get-peers-id-404"></span> 404 - Not Found
Status: Not Found

###### <span id="get-peers-id-404-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

##### <span id="get-peers-id-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-peers-id-500-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

### <span id="get-peers-id-neighbors"></span> Get peer neighbors (*GetPeersIDNeighbors*)

```
GET /api/peers/{id}/neighbors
```

Get the list of neighbors for a specific peer

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| id | `path` | string | `string` |  | ✓ |  | Peer ID (e.g. 12...) |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-peers-id-neighbors-200) | OK | OK |  | [schema](#get-peers-id-neighbors-200-schema) |
| [404](#get-peers-id-neighbors-404) | Not Found | Not Found |  | [schema](#get-peers-id-neighbors-404-schema) |
| [500](#get-peers-id-neighbors-500) | Internal Server Error | Internal Server Error |  | [schema](#get-peers-id-neighbors-500-schema) |

#### Responses


##### <span id="get-peers-id-neighbors-200"></span> 200 - OK
Status: OK

###### <span id="get-peers-id-neighbors-200-schema"></span> Schema
   
  

[ServerNeighborResponse](#server-neighbor-response)

##### <span id="get-peers-id-neighbors-404"></span> 404 - Not Found
Status: Not Found

###### <span id="get-peers-id-neighbors-404-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

##### <span id="get-peers-id-neighbors-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-peers-id-neighbors-500-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

### <span id="get-peers-map"></span> Get peer clusters for map visualisation (*GetPeersMap*)

```
GET /api/peers/map
```

Get geographical clusters of peers with their locations and counts

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| as_names | `query` | []string | `[]string` | `csv` |  |  | Names of Autonomous System Organisations |
| as_numbers | `query` | []string | `[]string` | `csv` |  |  | Numbers of Autonomous System Organisations |
| cities | `query` | []string | `[]string` | `csv` |  |  | City names |
| clients | `query` | []string | `[]string` | `csv` |  |  | Client names (e.g. aztec-node) |
| continents | `query` | []string | `[]string` | `csv` |  |  | Continent names |
| continents_code | `query` | []string | `[]string` | `csv` |  |  | Continent codes |
| countries | `query` | []string | `[]string` | `csv` |  |  | Country names |
| countries_codes | `query` | []string | `[]string` | `csv` |  |  | Country iso codes |
| latest | `query` | boolean | `bool` |  |  |  | Include peers from latest crawl only (default: true) |
| synced | `query` | boolean | `bool` |  |  |  | Sync status |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-peers-map-200) | OK | OK |  | [schema](#get-peers-map-200-schema) |
| [400](#get-peers-map-400) | Bad Request | Bad Request |  | [schema](#get-peers-map-400-schema) |
| [500](#get-peers-map-500) | Internal Server Error | Internal Server Error |  | [schema](#get-peers-map-500-schema) |

#### Responses


##### <span id="get-peers-map-200"></span> 200 - OK
Status: OK

###### <span id="get-peers-map-200-schema"></span> Schema
   
  

[ServerPeersMapResponse](#server-peers-map-response)

##### <span id="get-peers-map-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-peers-map-400-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

##### <span id="get-peers-map-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-peers-map-500-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

### <span id="get-peers-neighbors"></span> Get every peer's neighbors (*GetPeersNeighbors*)

```
GET /api/peers/neighbors
```

Get the list of neighbors for all peers

#### Consumes
  * application/json

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-peers-neighbors-200) | OK | OK |  | [schema](#get-peers-neighbors-200-schema) |
| [404](#get-peers-neighbors-404) | Not Found | Not Found |  | [schema](#get-peers-neighbors-404-schema) |
| [500](#get-peers-neighbors-500) | Internal Server Error | Internal Server Error |  | [schema](#get-peers-neighbors-500-schema) |

#### Responses


##### <span id="get-peers-neighbors-200"></span> 200 - OK
Status: OK

###### <span id="get-peers-neighbors-200-schema"></span> Schema
   
  

[ServerNeighborsResponse](#server-neighbors-response)

##### <span id="get-peers-neighbors-404"></span> 404 - Not Found
Status: Not Found

###### <span id="get-peers-neighbors-404-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

##### <span id="get-peers-neighbors-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="get-peers-neighbors-500-schema"></span> Schema
   
  

[ServerAPIError](#server-api-error)

## Models

### <span id="repo-a-s-count"></span> repo.ASCount


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| as_name | string| `string` |  | |  |  |
| as_number | integer| `int64` |  | |  |  |
| count | integer| `int64` |  | |  |  |



### <span id="repo-city-count"></span> repo.CityCount


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| city_name | string| `string` |  | |  |  |
| count | integer| `int64` |  | |  |  |



### <span id="repo-continent-count"></span> repo.ContinentCount


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| continent_code | string| `string` |  | |  |  |
| continent_name | string| `string` |  | |  |  |
| count | integer| `int64` |  | |  |  |



### <span id="repo-country-count"></span> repo.CountryCount


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| count | integer| `int64` |  | |  |  |
| country_code | string| `string` |  | |  |  |
| country_name | string| `string` |  | |  |  |



### <span id="repo-peer-history-point"></span> repo.PeerHistoryPoint


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| date | string| `string` |  | |  |  |
| peer_count | integer| `int64` |  | |  |  |



### <span id="repo-peer-info"></span> repo.PeerInfo


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| block_height | integer| `int64` |  | |  |  |
| client | string| `string` |  | |  |  |
| created_at | string| `string` |  | |  |  |
| id | string| `string` |  | |  |  |
| is_synced | boolean| `bool` |  | |  |  |
| last_seen | string| `string` |  | |  |  |
| multi_addresses | [][TypesMultiAddr](#types-multi-addr)| `[]*TypesMultiAddr` |  | |  |  |
| protocols | []string| `[]string` |  | |  |  |
| spec_version | string| `string` |  | |  |  |



### <span id="repo-peer-neighbors"></span> repo.PeerNeighbors


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| id | string| `string` |  | |  |  |
| neighbors | []string| `[]string` |  | |  |  |



### <span id="repo-protocol-count"></span> repo.ProtocolCount


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| count | integer| `int64` |  | |  |  |
| protocol | string| `string` |  | |  |  |



### <span id="server-api-error"></span> server.APIError


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| message | [any](#any)| `any` |  | |  |  |
| status | integer| `int64` |  | |  |  |



### <span id="server-a-s-count-response"></span> server.ASCountResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| networks | [][RepoASCount](#repo-a-s-count)| `[]*RepoASCount` |  | |  |  |



### <span id="server-agent-count-response"></span> server.AgentCountResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| agents | [][ServerAgentVersionCounts](#server-agent-version-counts)| `[]*ServerAgentVersionCounts` |  | |  |  |



### <span id="server-agent-version-counts"></span> server.AgentVersionCounts


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| client_name | string| `string` |  | |  |  |
| total_count | integer| `int64` |  | |  |  |
| versions | [][ServerVersionCount](#server-version-count)| `[]*ServerVersionCount` |  | |  |  |



### <span id="server-analytics-response"></span> server.AnalyticsResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| cities | [][RepoCityCount](#repo-city-count)| `[]*RepoCityCount` |  | |  |  |
| clients | [][ServerAgentVersionCounts](#server-agent-version-counts)| `[]*ServerAgentVersionCounts` |  | |  |  |
| continents | [][RepoContinentCount](#repo-continent-count)| `[]*RepoContinentCount` |  | |  |  |
| countries | [][RepoCountryCount](#repo-country-count)| `[]*RepoCountryCount` |  | |  |  |
| networks | [][RepoASCount](#repo-a-s-count)| `[]*RepoASCount` |  | |  |  |
| peers_churn_in | integer| `int64` |  | | Number of new peers that were seen in the latest crawl but were not seen in the previous crawls. |  |
| peers_churn_out | integer| `int64` |  | | Number of peers that were previously seen but are no longer seen in the latest crawl. |  |
| peers_latest | integer| `int64` |  | | Number of peers that were seen in the latest crawl. |  |
| protocols | [][RepoProtocolCount](#repo-protocol-count)| `[]*RepoProtocolCount` |  | |  |  |
| sync_status | map of integer| `map[string]int64` |  | |  |  |
| total_peers | integer| `int64` |  | | Total number of peers that were seen in the network. |  |



### <span id="server-city-count-response"></span> server.CityCountResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| cities | [][RepoCityCount](#repo-city-count)| `[]*RepoCityCount` |  | |  |  |



### <span id="server-cluster"></span> server.Cluster


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| city | string| `string` |  | |  |  |
| continent | string| `string` |  | |  |  |
| count | integer| `int64` |  | |  |  |
| country | string| `string` |  | |  |  |
| latitude | number| `float64` |  | |  |  |
| longitude | number| `float64` |  | |  |  |
| peer_ids | []string| `[]string` |  | |  |  |



### <span id="server-continent-count-response"></span> server.ContinentCountResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| continents | [][RepoContinentCount](#repo-continent-count)| `[]*RepoContinentCount` |  | |  |  |



### <span id="server-country-count-response"></span> server.CountryCountResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| countries | [][RepoCountryCount](#repo-country-count)| `[]*RepoCountryCount` |  | |  |  |



### <span id="server-neighbor-response"></span> server.NeighborResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| id | string| `string` |  | |  |  |
| neighbors | []string| `[]string` |  | |  |  |



### <span id="server-neighbors-response"></span> server.NeighborsResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| neighbors | [][RepoPeerNeighbors](#repo-peer-neighbors)| `[]*RepoPeerNeighbors` |  | |  |  |



### <span id="server-peer-count-response"></span> server.PeerCountResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| peers_churn_in | integer| `int64` |  | |  |  |
| peers_churn_out | integer| `int64` |  | |  |  |
| peers_latest | integer| `int64` |  | |  |  |
| total_peers | integer| `int64` |  | |  |  |



### <span id="server-peer-history-response"></span> server.PeerHistoryResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| history | [][RepoPeerHistoryPoint](#repo-peer-history-point)| `[]*RepoPeerHistoryPoint` |  | |  |  |



### <span id="server-peer-response"></span> server.PeerResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| block_height | integer| `int64` |  | |  |  |
| client | string| `string` |  | |  |  |
| created_at | string| `string` |  | |  |  |
| id | string| `string` |  | |  |  |
| is_synced | boolean| `bool` |  | |  |  |
| last_seen | string| `string` |  | |  |  |
| multi_addresses | [][TypesMultiAddr](#types-multi-addr)| `[]*TypesMultiAddr` |  | |  |  |
| protocols | []string| `[]string` |  | |  |  |
| spec_version | string| `string` |  | |  |  |



### <span id="server-peers-map-response"></span> server.PeersMapResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| active_peers | integer| `int64` |  | |  |  |
| clusters | [][ServerCluster](#server-cluster)| `[]*ServerCluster` |  | |  |  |
| total_peers | integer| `int64` |  | |  |  |



### <span id="server-peers-response"></span> server.PeersResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| next_pagination_token | string| `string` |  | |  |  |
| peers | [][RepoPeerInfo](#repo-peer-info)| `[]*RepoPeerInfo` |  | |  |  |
| total_peers | integer| `int64` |  | |  |  |



### <span id="server-protocol-count-response"></span> server.ProtocolCountResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| protocols | [][RepoProtocolCount](#repo-protocol-count)| `[]*RepoProtocolCount` |  | |  |  |



### <span id="server-version-count"></span> server.VersionCount


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| count | integer| `int64` |  | |  |  |
| version | string| `string` |  | |  |  |



### <span id="types-ip-info"></span> types.IPInfo


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| as_name | string| `string` |  | |  |  |
| as_number | integer| `int64` |  | |  |  |
| city_name | string| `string` |  | |  |  |
| continent_code | string| `string` |  | |  |  |
| continent_name | string| `string` |  | |  |  |
| country_iso | string| `string` |  | |  |  |
| country_name | string| `string` |  | |  |  |
| ip_address | string| `string` |  | |  |  |
| latitude | number| `float64` |  | |  |  |
| longitude | number| `float64` |  | |  |  |
| port | integer| `int64` |  | | Refers to the port number of the peer when connected via p2p.</br>Not the HTTP port or any other port. |  |



### <span id="types-multi-addr"></span> types.MultiAddr


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| ip_info | [][TypesIPInfo](#types-ip-info)| `[]*TypesIPInfo` |  | |  |  |
| maddr | string| `string` |  | |  |  |


