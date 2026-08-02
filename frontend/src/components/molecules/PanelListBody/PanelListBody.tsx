import React from 'react';
import './PanelListBody.css';
import PanelSection from '../PanelSection/PanelSection';
import { clientsStatsAtom, countriesStatsAtom } from '../../../hooks/atoms';
import { useAtom } from 'jotai';
import { useAnalytics } from '../../../api/analytics';

interface Section {
  view: string;
  containerClass: string;
  bodyClass?: string;
  items: {
    title: string;
    items?: { label: string; value: string; percentage?: string }[];
    isSync?: boolean;
  }[];
}

const disableSyncStatus = import.meta.env.VITE_DISABLE_SYNC_STATUS === 'true';

const PanelListBody: React.FC<{ view: string }> = ({ view }) => {
  const [countriesStats] = useAtom(countriesStatsAtom);
  const [clientsStats] = useAtom(clientsStatsAtom);
  const { data: analyticsData } = useAnalytics();

  // Calculate total number of network items

  // const activeClientsInCountry = Object.fromEntries(
  //   Object.keys(countries).map((country) => {
  //     return [
  //       country,
  //       activePeers.filter((peer) =>
  //         peer.multi_addresses.some((address) => address.ip_info.some((info) => info.country_name === country))
  //       ),
  //     ];
  //   })
  // );

  // const totalCountries = Object.values(activeClientsInCountry || {}).reduce(
  //   (sum, clients) => sum + (clients?.length || 0),
  //   0
  // );
  // Sort network items by the number of clients in descending order and take the top 5
  // const countryItemsSorted = Object.entries(activeClientsInCountry || {})
  //   .sort((a, b) => (b[1]?.length || 0) - (a[1]?.length || 0))
  //   .slice(0, 6)
  //   .map(([country, clients]) => ({
  //     label: country,
  //     value: (clients?.length || 0).toString(),
  //     percentage: `${calculatePercentage(clients?.length || 0, totalCountries || 1, 2)}%`,
  //   }));

  const top6Contries = countriesStats?.slice(0, 6).map((item) => ({
    label: item.country_name,
    value: item.count.toString(),
    percentage: item.percent,
  }));

  const top6Clients = clientsStats?.slice(0, 6).map((item) => ({
    label: item.name,
    value: item.total.toString(),
    percentage: item.percent,
  }));

  // Create sync status data for the PanelSection.
  // Buckets come from the backend aggregator: `synced`, `not_synced` (peers we
  // checked and concluded are behind/forked), and `unknown` (peers we couldn't
  // judge — e.g. reference RPC was unreachable). Show all three so a silent
  // classifier outage doesn't masquerade as the whole network syncing.
  const syncStatusData = analyticsData?.sync_status
    ? (() => {
        const synced = analyticsData.sync_status.synced ?? 0;
        const notSynced = analyticsData.sync_status.not_synced ?? 0;
        const unknown = analyticsData.sync_status.unknown ?? 0;
        const total = synced + notSynced + unknown;
        const pct = (n: number) => (total > 0 ? `${((n / total) * 100).toFixed(1)}%` : '0%');

        return [
          { label: 'synced', value: synced.toString(), percentage: pct(synced) },
          { label: 'unsynced', value: notSynced.toString(), percentage: pct(notSynced) },
          { label: 'unknown', value: unknown.toString(), percentage: pct(unknown) },
        ];
      })()
    : [];

  // Define sections based on view
  const sections: Section[] = [
    {
      view: 'ListView',
      containerClass: 'panel-list-body-container',
      bodyClass: 'panel-list-body panel-row',
      items: [
        // { title: 'Sync Status', items: syncStatusData, isSync: true },
        { title: 'Top Countries', items: top6Contries },
        { title: 'Top Versions', items: top6Clients },
      ],
    },
    {
      view: 'MapView',
      containerClass: 'panel-body-container',
      bodyClass: 'panel-body',
      items: [
        // { title: 'Sync Status', items: syncStatusData, isSync: true },
        { title: 'Top Countries', items: top6Contries?.slice(0, 4) },
        { title: 'Top Versions', items: top6Clients?.slice(0, 3) },
      ],
    },
    {
      view: 'TopNetworks',
      containerClass: 'panel-body-container',
      items: [{ title: 'Top Countries', items: top6Contries }],
    },
    {
      view: 'TopClients',
      containerClass: 'panel-body',
      items: [{ title: 'Top Versions', items: top6Clients }],
    },
    {
      view: 'SyncStatus',
      containerClass: 'panel-body-container',
      bodyClass: 'panel-body',
      items: [{ title: 'Sync Status', items: syncStatusData, isSync: true }],
    },
  ];

  if (!disableSyncStatus) {
    // If sync status is not disabled, include the SyncStatus section
    sections.push({
      view: 'SyncStatus',
      containerClass: 'panel-body-container',
      bodyClass: 'panel-body',
      items: [{ title: 'Sync Status', items: syncStatusData, isSync: true }],
    });
  }

  return (
    <>
      {sections.map(
        (section) =>
          view === section.view && (
            <div className={`${section.containerClass}`} key={section.view}>
              <div className={`${section.bodyClass}`}>
                {section.items.map((item) => (
                  <PanelSection key={item.title} title={item.title} items={item.items} isSync={item.isSync} />
                ))}
              </div>
            </div>
          )
      )}
    </>
  );
};

export default PanelListBody;
