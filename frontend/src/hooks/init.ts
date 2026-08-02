//API Hooks to initialize data at root.

import { useAtom } from 'jotai';
import { useEffect } from 'react';
import { API } from '../api';
import {
  activeContinentsAtom,
  countriesStatsAtom,
  nodeModalAtom,
  historyAtom,
  activeClientsAtom,
  clientsStatsAtom,
  totalActivePeers,
} from './atoms';
import { Clients } from '../types';

function isResultDiff<T>(next: T, prev: T): boolean {
  try {
    return JSON.stringify(prev) !== JSON.stringify(next);
  } catch {
    return true;
  }
}

export const useInitCountriesStats = () => {
  const { data } = API.useAnalyticsCountries();
  const [, setCountriesStats] = useAtom(countriesStatsAtom);

  useEffect(() => {
    if (!data || !data.countries) return;

    const sorted = data.countries.sort((a, b) => b.count - a.count);
    const sum = sorted.reduce((acc, curr) => acc + curr.count, 0);
    const parsed = sorted
      .filter((item) => item.country_name.toLowerCase() !== 'unknown')
      .map((item) => ({
        ...item,
        percent: `${parseFloat(((item.count / sum) * 100).toFixed(2))}%`,
      }));

    setCountriesStats((prev) => {
      return isResultDiff(prev, parsed) ? parsed : prev;
    });
  }, [data, setCountriesStats]);
};

export const useInitContinents = () => {
  const { data } = API.useAnalyticsContinents();
  const [, setActiveContinents] = useAtom(activeContinentsAtom);

  useEffect(() => {
    if (!data || !data.continents) return;
    setActiveContinents(() => {
      return data.continents;
    });
  }, [data, setActiveContinents]);
};

// Group raw version strings into major.minor buckets so the Top Versions panel
// stays readable when a long tail of pre-release/nightly versions is reported.
// e.g. "4.1.3" -> "4.1.x", "4.1.0-rc.2" -> "4.1.x". "Unknown" / "Other" are
// passed through unchanged.
const versionBucketRegexp = /^(\d+)\.(\d+)\./;
const bucketVersion = (name: string): string => {
  const match = versionBucketRegexp.exec(name);
  if (!match) return name;
  return `${match[1]}.${match[2]}.x`;
};

export const useInitClients = () => {
  const { data } = API.useAnalyticsClients();
  const [, setActiveClients] = useAtom(activeClientsAtom);
  const [, setClientStats] = useAtom(clientsStatsAtom);

  useEffect(() => {
    if (!data || !data.agents) return;

    const clientsData = data.agents.reduce((acc, client) => {
      const versions = client.versions.reduce(
        (versionAcc, version) => {
          versionAcc[version.version] = {
            synced: version.count,
            unsynced: 0,
          };
          return versionAcc;
        },
        {} as Record<string, { synced: number; unsynced: number }>
      );

      acc[client.client_name] = { total: client.total_count, versions };
      return acc;
    }, {} as Clients);

    setActiveClients(() => clientsData);

    // Bucket by major.minor for the Top Versions panel
    const bucketedTotals = Object.entries(clientsData).reduce(
      (acc, [name, { total }]) => {
        const bucket = bucketVersion(name);
        acc[bucket] = (acc[bucket] ?? 0) + total;
        return acc;
      },
      {} as Record<string, number>
    );

    const sortedBuckets = Object.entries(bucketedTotals).sort(([, a], [, b]) => b - a);

    setClientStats(() => {
      const total = sortedBuckets.reduce((sum, [, count]) => sum + count, 0);
      return sortedBuckets.map(([bucketName, bucketTotal]) => {
        return {
          name: bucketName,
          total: bucketTotal,
          versions: {},
          percent: `${((bucketTotal / total) * 100).toFixed(2)}%`,
        };
      });
    });
  }, [data, setActiveClients]);
};

export const useInitPeerHistory = () => {
  const { data } = API.useAnalyticsPeerHistory({
    start: new Date('2024-01-01'),
    end: new Date(),
  });
  const [, setHistory] = useAtom(historyAtom);

  useEffect(() => {
    if (!data || !data.history) return;
    setHistory(data.history);
  }, [data, setHistory]);
};

export const useInitNodeModal = () => {
  const [nodeModal, setNodeModal] = useAtom(nodeModalAtom);
  const { data } = API.usePeerId(nodeModal.nodeId);

  useEffect(() => {
    if (!data) return;
    setNodeModal((prev) => {
      return {
        ...prev,
        node: isResultDiff(prev.node, data) && prev.nodeId === data.id ? data : prev.node,
      };
    });
  }, [data, nodeModal.nodeId, setNodeModal]);
};

export const useInitAnalytics = () => {
  const { data } = API.useAnalytics();
  const [, setTotalActivePeers] = useAtom(totalActivePeers);

  useEffect(() => {
    if (!data) return;
    // Use peers_latest from analytics endpoint for total active peers
    setTotalActivePeers(data.peers_latest);
  }, [data, setTotalActivePeers]);
};
