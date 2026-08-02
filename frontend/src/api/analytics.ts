import {
  AnalyticsContinents,
  AnalyticsCountries,
  AnalyticsPeers,
  AnalyticsResponse,
  ClientAnalytics,
  HistoryResponse,
} from '../types';
import { getBaseUrl, useSWR } from './shared';

function getApiBaseURL() {
  return getBaseUrl() + '/analytics';
}

//All Analytics related endpoints

export const useAnalyticsPeers = () => {
  return useSWR<AnalyticsPeers>(getApiBaseURL() + '/peers');
};

export const useAnalyticsCountries = () => {
  return useSWR<AnalyticsCountries>(getApiBaseURL() + '/countries');
};

export const useAnalyticsContinents = () => {
  return useSWR<AnalyticsContinents>(getApiBaseURL() + '/continents');
};

export const useAnalyticsClients = () => {
  return useSWR<{ agents: ClientAnalytics[] }>(getApiBaseURL() + '/clients');
};

export const useAnalyticsPeerHistory = (params: { start: Date; end: Date }) => {
  const startDate = params.start.toISOString();
  const endDate = params.end.toISOString();

  return useSWR<HistoryResponse>(getApiBaseURL() + '/peers/history', { start: startDate, end: endDate });
};

export const useAnalytics = () => {
  return useSWR<AnalyticsResponse>(getApiBaseURL());
};
