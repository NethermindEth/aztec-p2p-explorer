import { Peer } from '../types';

interface Filters {
  region?: string[];
  client?: string[];
  aso?: string[];
}

function extractRegionName(input: string | null | undefined): string {
  return input?.split(' ')[0] ?? '';
}

export function applyFilters(peers: Peer[], filters: Filters = {}): Peer[] {
  const { region, client, aso } = filters;

  // checks for 'All' or undefined filters
  const filterRegions = region && !region.includes('All') ? region.map(extractRegionName) : null;

  const filterClients = client && !client.includes('All') ? client.map((c) => c.toLowerCase()) : null;
  const filterAsos = aso && !aso.includes('All') ? aso.map((a) => a.trim()) : null;

  // single filter operation with combined conditions
  return peers.filter((peer) => {
    const regionMatch =
      !filterRegions ||
      filterRegions.includes(extractRegionName(peer?.multi_addresses?.[0]?.ip_info?.[0]?.continent_name ?? 'Unknown'));

    const clientMatch =
      !filterClients || filterClients.includes(peer?.client?.split('/')?.[0]?.toLowerCase() ?? 'unknown');

    const asoMatch = !filterAsos || filterAsos.includes(peer?.multi_addresses?.[0]?.ip_info?.[0]?.as_name ?? 'Unknown');

    return regionMatch && clientMatch && asoMatch;
  });
}
