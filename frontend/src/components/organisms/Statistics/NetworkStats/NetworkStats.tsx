import { useState, useEffect, useRef } from 'react';
import * as Tabs from '@radix-ui/react-tabs';
import { useAtom } from 'jotai';
import NodeList from '../NodeList/NodeList';
//import TopNetworks from '../TopNetworks/TopNetworks';
// import ClientType from '../ClientType/ClientType';
import Region from '../Region/Region';
import NodeHistory from '../NodeHistory/NodeHistory';
import SearchBar from '../../../molecules/SearchBar/SearchBar';
import PanelNodeDetails from '../../PanelNodeDetails/PanelNodeDetails';
import StatisticsFilter from '../../../molecules/StatisticsFilter/StatisticsFilter';
import SelectedFilters from '../../../molecules/StatisticsSelectedFilter/SelectedFilters';
import MobileStatisticsFilter from '../../../molecules/MobileStatisticsFilter/MobileStatisticsFilter';
//import mobileFilterIcon from '../../../../assets/icons/mobileFilter.svg';
import RenderTabTitles from './RenderTabTitles';

import { activeClientsAtom, isMobileAtom, searchQueryAtom, selectedNodeAtom } from '../../../../hooks/atoms';
import { usePeers } from '../../../../api/peers';
import { capitalizeFirstLetter } from '../../../../utils';
import { Peer, FilterValues } from '../../../../types';
import filterImg from '../../../../assets/icons/filter-icon.svg';

import './NetworkStats.css';
import MobileSelector from '../../../molecules/MobileSelector/MobileSelector';
import Pager from '../Pagination/Pager';
import useLocationHash from '../../../../hooks/useLocationHash';
import { useNavigate } from 'react-router-dom';

export enum TabViews {
  NodeList = 'nodes',
  Region = 'region',
  NodeHistory = 'history',
}

export const tabViewOptions = [
  {
    label: 'Node List',
    value: TabViews.NodeList,
  },
  {
    label: 'Distribution By Region',
    value: TabViews.Region,
  },
  {
    label: 'Node History',
    value: TabViews.NodeHistory,
  },
];

const disableSyncStatus = import.meta.env.VITE_DISABLE_SYNC_STATUS === 'true';

const NetworkStats = () => {
  const [isMobile] = useAtom(isMobileAtom);
  const [searchQuery, setSearchQuery] = useAtom(searchQueryAtom);
  const [selectedNode, setSelectedNode] = useAtom(selectedNodeAtom);
  const [clients] = useAtom(activeClientsAtom);
  const [filterValues, setFilterValues] = useState<FilterValues>({
    client: ['All'],
    sync: ['All'],
  });
  const [clientsOptions, setClientsOptions] = useState<string[]>([]);

  const [currentPage, setCurrentPage] = useState(1);
  const [sortColumn, setSortColumn] = useState<string>('last_seen');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc'); // TODO better to use enum
  const [itemsPerPage, setItemsPerPage] = useState(10);
  const [paginationTokens, setPaginationTokens] = useState<Map<number, string>>(new Map());

  // Build API parameters for the current page
  const apiParams = {
    page_size: itemsPerPage,
    latest: true,
    sort: sortColumn,
    order: sortOrder,
    // Add filtering parameters
    clients: filterValues.client.includes('All') ? undefined : filterValues.client.map((c) => c.toLowerCase()),
    synced: filterValues.sync.includes('All') ? undefined : filterValues.sync.includes('Synced'),
    // Add search by peer ID if provided
    id: searchQuery.trim() || undefined,
    // Add pagination token for pages beyond the first
    pagination_token: currentPage > 1 ? paginationTokens.get(currentPage) : undefined,
  };

  // Use the paginated peers API
  const { data: peersResponse, isLoading: peersLoading } = usePeers(apiParams);

  usePeers({ ...apiParams, pagination_token: paginationTokens.get(currentPage + 1) ?? undefined });

  const activePeers = peersResponse?.peers || [];

  // Store pagination token for next page
  useEffect(() => {
    if (peersResponse?.next_pagination_token && currentPage < 1000) {
      setPaginationTokens((prev) => new Map(prev).set(currentPage + 1, peersResponse.next_pagination_token));
    }
  }, [peersResponse?.next_pagination_token, currentPage]);

  const paginationRef = useRef<HTMLDivElement>(null);
  const disableRowClickRef = useRef(false);
  const tabsRef = useRef<HTMLDivElement>(null);

  const navigate = useNavigate();
  const hash = useLocationHash();

  // State for controlled tabs
  const [activeTab, setActiveTab] = useState<TabViews>(() => {
    return !hash ? TabViews.NodeList : (hash as TabViews);
  });

  // Sync state with URL hash
  useEffect(() => {
    const tabFromHash = !hash ? TabViews.NodeList : (hash as TabViews);
    setActiveTab(tabFromHash);
  }, [hash]);

  const handleTabSelect = (val: TabViews) => {
    if (!val) return;
    // Update state immediately for responsive UI
    setActiveTab(val);
    // Then update URL
    navigate(`/explore#${val}`);
  };

  const syncOptions = ['All', 'Synced', 'Unsynced'];

  const clientsArray = ['All', ...Object.keys(clients || {})];

  const createClientsOptions = () => clientsArray.map((client) => capitalizeFirstLetter(client));

  useEffect(() => {
    setClientsOptions(createClientsOptions());
  }, [clients, clientsArray]);

  const handleFilterChange = (newClient: string[], newSync: string[]) => {
    if (newClient.length === 0) newClient = ['All'];
    if (newSync.length === 0) newSync = ['All'];
    setFilterValues({ client: newClient, sync: newSync });
    setCurrentPage(1);
  };

  // Use the peers directly from API without client-side filtering
  const currentData = activePeers;

  const statisticsFilterSelector = [
    {
      key: 'client',
      label: 'Client Version',
      value: filterValues.client,
      options: clientsOptions,
      onFilterChange: (value: string[]) => handleFilterChange(value, filterValues.sync),
    },
  ];

  if (!disableSyncStatus) {
    statisticsFilterSelector.push({
      key: 'sync',
      label: 'Sync status',
      value: filterValues.sync,
      options: syncOptions,
      onFilterChange: (value: string[]) => handleFilterChange(filterValues.client, value),
    });
  }

  // Reset to page 1 when filters or sort change
  useEffect(() => {
    setCurrentPage(1);
    setPaginationTokens(new Map()); // Clear pagination tokens when filters change
  }, [sortColumn, sortOrder, searchQuery, filterValues]);

  useEffect(() => {
    if (currentPage) {
      handleScrollToPagination();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentPage]);

  useEffect(() => {
    if (disableRowClickRef.current) {
      disableRowClickRef.current = false;
    }
  }, [activeTab]);

  const handleScrollToPagination = () => {
    if (currentPage == totalPages - 1) {
      if (paginationRef.current) {
        paginationRef.current.scrollIntoView({ behavior: 'instant', block: 'center' });
      }
    }
  };

  const handlePageChange = (val: number) => {
    setCurrentPage(val);
  };

  const handleRowClick = (node: Peer) => {
    if (disableRowClickRef.current) {
      return;
    }
    setSelectedNode(node);
  };

  const handleSort = (column: string, order: 'asc' | 'desc') => {
    // Map frontend sort columns to API sort columns
    let apiSortColumn = column;
    if (column === 'id') {
      apiSortColumn = 'id';
    } else if (column === 'client') {
      apiSortColumn = 'client';
    } else if (column === 'clientName' || column === 'clientVersion') {
      // For now, we'll handle client sorting on the frontend since the API doesn't support it directly
      apiSortColumn = 'client';
    } else if (column === 'block_height') {
      apiSortColumn = 'block_height';
    } else if (column === 'created_at') {
      apiSortColumn = 'created_at';
    } else if (column === 'last_seen') {
      apiSortColumn = 'last_seen';
    } else {
      apiSortColumn = 'last_seen'; // default
    }

    setSortColumn(apiSortColumn);
    setSortOrder(order);
  };

  // Calculate total pages using the actual total count from API
  const totalCount = peersResponse?.total_peers || 0;
  const totalPages = Math.max(1, Math.ceil(totalCount / itemsPerPage));

  const [openFilter, setOpenFilter] = useState<string | null>(null);
  const [isMobileFilterPanelOpen, setIsMobileFilterPanelOpen] = useState(false);

  const handleFilterToggle = (filterKey: string) => {
    if (openFilter === filterKey) {
      setOpenFilter(null);
    } else {
      setOpenFilter(filterKey);
    }
  };

  const handleMobileFilterClick = () => {
    setIsMobileFilterPanelOpen((prev) => !prev);
  };

  const handleApplyFilters = (newFilterValues: FilterValues) => {
    setFilterValues(newFilterValues);
    setIsMobileFilterPanelOpen(false);
    setCurrentPage(1);
  };

  // Reset selected node when changing tabs
  useEffect(() => {
    if (!isMobile && activeTab !== TabViews.NodeList) {
      setSelectedNode(null);
    }
  }, [activeTab, isMobile]);

  return (
    <div className="network-stats-wrapper">
      <h1 className="network-stats__title">Aztec Statistics</h1>
      <div className={'network-stats'}>
        <Tabs.Root
          value={activeTab}
          className="network-stats__tabs"
          onValueChange={(val) => handleTabSelect(val as TabViews)}
          ref={tabsRef}
        >
          <div className="network-stats__tabs-header">
            <Tabs.List className="network-stats__tabs-list">{RenderTabTitles()}</Tabs.List>
          </div>

          <div>
            {isMobile && (
              <>
                <span className="listview-mobile-title">Aztec Statistics</span>
                <div className="network-stats__tabs-header-mobile">
                  <div className="mobile-symbol-select-wrapper">
                    <MobileSelector
                      defaultValue={activeTab}
                      options={tabViewOptions}
                      setMobileActiveTab={(val) => {
                        setActiveTab(val);
                        navigate(`/explore#${val}`);
                      }}
                    />
                  </div>
                  {activeTab === TabViews.NodeList && (
                    <div
                      className={`mobile-filters ${isMobileFilterPanelOpen ? 'is-open' : ''}`}
                      onClick={handleMobileFilterClick}
                    >
                      <span className="mobile-filters-text">Filters </span>
                      <img src={filterImg} alt="" />
                    </div>
                  )}
                </div>
              </>
            )}
            {activeTab === TabViews.NodeList && isMobileFilterPanelOpen && (
              <MobileStatisticsFilter
                filters={statisticsFilterSelector}
                onClose={() => setIsMobileFilterPanelOpen(false)}
                onApplyFilters={handleApplyFilters}
              />
            )}

            <div className="network-stats-filter-wrapper" aria-hidden={activeTab !== TabViews.NodeList}>
              {!isMobile && activeTab === TabViews.NodeList && (
                <div className="network-stats-filter-container">
                  {statisticsFilterSelector.map(({ key, label, value, options, onFilterChange }) => (
                    <StatisticsFilter
                      key={`${key}-dropdown-${value.join(',')}`}
                      label={label}
                      value={value}
                      options={options}
                      onFilterChange={onFilterChange}
                      isOpen={openFilter === key}
                      onClose={() => handleFilterToggle(key)}
                    />
                  ))}
                </div>
              )}

              {activeTab === TabViews.NodeList && <SearchBar onChange={(e) => setSearchQuery(e.target.value)} />}
            </div>

            <section className="network-stats__tabs-container">
              {isMobile && activeTab === TabViews.NodeList && (
                <div className="network-stats__tabs-header-mobile">
                  {!selectedNode && (
                    <div className="listview-searchbar">
                      {activeTab === TabViews.NodeList && (
                        <SearchBar onChange={(e) => setSearchQuery(e.target.value)} />
                      )}
                    </div>
                  )}
                </div>
              )}

              {activeTab === TabViews.NodeList && (
                <SelectedFilters filterValues={filterValues} handleFilterChange={handleFilterChange} />
              )}
              <Tabs.Content value={TabViews.NodeList}>
                <div
                  style={{
                    display: isMobile ? 'block' : 'grid',
                    gridTemplateColumns: selectedNode && !isMobile ? '1.85fr 1fr' : '1fr',
                    gap: '24px',
                  }}
                >
                  <NodeList
                    data={currentData}
                    onRowClick={handleRowClick}
                    onSort={handleSort}
                    sortColumn={sortColumn}
                    sortOrder={sortOrder}
                    isLoading={peersLoading}
                  />

                  {selectedNode && (
                    <aside className="network-stats-panel desktop-only">
                      <PanelNodeDetails node={selectedNode} />
                    </aside>
                  )}
                </div>
              </Tabs.Content>

              <Tabs.Content value={TabViews.Region}>
                <Region />
              </Tabs.Content>

              <Tabs.Content value={TabViews.NodeHistory}>
                <NodeHistory />
              </Tabs.Content>

              {/* {activeTab === 'topASOs' && <TopNetworks data={filteredNetworkData} />} */}
              {/* {activeTab === 'clientType' && <ClientType filterValues={filterValues} />} */}
            </section>
          </div>
        </Tabs.Root>
        {/* TODO: Should be in NodeList */}
        <div className="network-stats__pagination-container" ref={paginationRef}>
          {activeTab === TabViews.NodeList && currentData.length > 0 && (
            <Pager
              currentPage={currentPage}
              totalPages={totalPages}
              onPageChange={handlePageChange}
              rowsPerPage={itemsPerPage}
              onRowsPerPageChange={setItemsPerPage}
              isLoading={peersLoading}
            />
          )}
        </div>
      </div>
    </div>
  );
};

export default NetworkStats;
