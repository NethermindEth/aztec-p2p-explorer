import { atom } from 'jotai';
import {
  Clients,
  SyncStatus,
  NetworkItems,
  Peer,
  Continent,
  Countries,
  HistoryData,
  PeersMapResponse,
  AnalyticsCountry,
  NodeModal,
  ClientVersionInfo,
} from '../types/';

export const activeClientsAtom = atom<Clients>({});
export const clientsStatsAtom = atom<Array<ClientVersionInfo & { percent: string; name: string }>>();
export const syncStatusAtom = atom<SyncStatus>({ not_synced: 0, synced: 0 });
export const networkItemsAtom = atom<NetworkItems>({});
export const activeNetworkItemsAtom = atom<NetworkItems>({});
export const continentsAtom = atom<Continent[]>([{ continent_name: '', continent_code: '', count: 0 }]);
export const activeContinentsAtom = atom<Continent[]>([{ continent_name: '', continent_code: '', count: 0 }]);
export const countriesStatsAtom = atom<Array<AnalyticsCountry & { percent: string }>>();
export const countriesAtom = atom<Countries>({});
export const activeCountriesAtom = atom<Countries>({});

export const showListViewAtom = atom(false);

export const imageDataAtom = atom<ImageData | null>(null);
export const positionsAtom = atom<Float32Array | null>(null);

export const globeRotationAtom = atom<boolean>(true);

export const isMobileAtom = atom<boolean>(false); //useIsMobile handles this

export const isTabletAtom = atom<boolean>(false); //useIsMobile handles this

export const selectedNodeAtom = atom<Peer | null>(null);

export const isFilterOpenAtom = atom<boolean>(false);

export const searchQueryAtom = atom<string>('');

export const networkTotalsAtom = atom<{ [key: string]: number }>({});
export const networkPercentagesAtom = atom<{ [key: string]: string }>({});
export const sortedNetworksAtom = atom<string[]>([]);

export const sortedNodeListDataAtom = atom<Peer[]>([]);

export const selectedRowAtom = atom<Peer | null>(null);

export const historyAtom = atom<HistoryData[]>([]);

// Network is now determined by deployment, not user selection
export const currentNetworkAtom = atom<string>('Aztec Testnet');

export const clusterPeersAtom = atom<PeersMapResponse['clusters']>([]);

export const totalActivePeers = atom<number | undefined>(undefined);

export const nodeModalAtom = atom<NodeModal>({
  isOpen: false,
  nodeId: '',
  node: null,
});

// Cloudflare Turnstile token management
export const turnstileTokenAtom = atom<string | null>(null);
export const turnstileTokenExpiryAtom = atom<number | null>(null); // timestamp in ms
export const turnstileReadyAtom = atom<boolean>(false); // true when token is available or Turnstile is disabled
export const turnstileShowModalAtom = atom<boolean>(false); // controls modal visibility
export const turnstileNeedsBackgroundRefreshAtom = atom<boolean>(false); // triggers background re-verification
