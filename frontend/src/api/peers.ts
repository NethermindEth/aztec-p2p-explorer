import { SWRConfiguration } from 'swr';
import { PeersResponse, PeersMapResponse } from '../types';
import { getBaseUrl, useSWR } from './shared';

function getApiBaseURL() {
  return getBaseUrl() + '/peers';
}

//All Peer related endpoints
export const usePeers = (
  params?: {
    id?: string;
    continents?: string[];
    continents_code?: string[];
    countries?: string[];
    countries_codes?: string[];
    cities?: string[];
    as_names?: string;
    as_numbers?: string[];
    clients?: string[];
    synced?: boolean;
    sort?: string;
    order?: string;
    latest?: boolean;
    page_size?: number;
    pagination_token?: string;
  },
  options?: SWRConfiguration
) => {
  return useSWR<PeersResponse>(getApiBaseURL(), params, options);
};

export const usePeerId = (id?: string, options?: SWRConfiguration) => {
  const { data, ...result } = usePeers({ id: id }, options);
  return { ...result, data: data?.peers?.[0] || null };
};

//TODO — when API is complete
export const usePeersMap = (params?: {
  continents?: string[];
  continents_code?: string[];
  countries?: string[];
  countries_codes?: string[];
  cities?: string[];
  as_names?: string;
  as_numbers?: string[];
  clients?: string[];
  synced?: boolean;
  latest?: boolean;
}) => {
  return useSWR<PeersMapResponse>(getApiBaseURL() + '/map', params);
};
