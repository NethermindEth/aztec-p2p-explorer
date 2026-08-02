import React, { Dispatch, SetStateAction, useEffect, useState } from 'react';
import Dropdown from '../Dropdown/Dropdown';
import { Continent } from '../../../types';
import RouterLink from '../../atoms/RouterLink/RouterLink';
import './FilterBar.css';
import { IFilters } from '../../../views/MapView/MapView';
import { useAtom } from 'jotai';
import { activeContinentsAtom, activeClientsAtom } from '../../../hooks/atoms';

type FilterDropdownsProps = {
  onFilterChange: Dispatch<SetStateAction<IFilters>>;
};

const FilterBar: React.FC<FilterDropdownsProps> = ({ onFilterChange }) => {
  const [region, setRegion] = useState<string[]>(
    localStorage.getItem('region') ? JSON.parse(localStorage.getItem('region')!) : ['All']
  );
  const [clients, setClients] = useState<string[]>(
    localStorage.getItem('client') ? JSON.parse(localStorage.getItem('client')!) : ['All']
  );
  // const [network, setNetwork] = useState<string[]>(
  //   localStorage.getItem('network') ? JSON.parse(localStorage.getItem('network')!) : ['All']
  // );
  // const [networkOptions, setNetworkOptions] = useState<string[]>([]);
  const [openDropdown, setOpenDropdown] = useState<string | null>(null);
  const [continents] = useAtom(activeContinentsAtom);
  const [activeClients] = useAtom(activeClientsAtom);

  // const { networkItems, continents /*clients*/ } = data;
  const clientsOptions = ['All', ...Object.keys(activeClients)];

  const clientCounts = Object.entries(activeClients).reduce(
    (acc, [clientType, versions]) => {
      return { ...acc, [clientType.toLowerCase()]: versions.total };
    },
    {
      all: Object.values(activeClients).reduce((total, versions) => total + versions.total, 0),
    }
  );

  useEffect(() => {
    localStorage.setItem('region', JSON.stringify(region));
    localStorage.setItem('client', JSON.stringify(clients));
    // localStorage.setItem('network', JSON.stringify(network));
    onFilterChange({ region, clients });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [region, clients]);

  // function createRegionsOptions() {
  //   const regionsArray = continents.map(({ continent_name }) => continent_name || 'Unknown');
  //   regionsArray.sort();
  //   regionsArray.unshift('All');
  //   return regionsArray;
  // }

  // function createNetworkOptions() {
  //   const networkArray = Object.entries(networkItems).map(([asos]) => `${asos === '' ? 'Unknown' : asos}`);
  //   networkArray.sort();
  //   networkArray.unshift('All');
  //   return networkArray;
  // }

  // function createClientsOptions() {
  //   return clientsArray.map((client) => client.charAt(0).toUpperCase() + client.slice(1));
  // }

  // const getRegionCounts = (continents: Continent[]) => {
  //   const counts: { [key: string]: number } = { all: 0, unknown: 0 };

  //   continents.forEach(({ continent_name, count }) => {
  //     const normalizedKey = (continent_name || 'unknown').toLowerCase();
  //     counts[normalizedKey] = (counts[normalizedKey] || 0) + count;
  //     counts['all'] += count;
  //   });

  //   return counts;
  // };

  // const getClientCounts = (clients: Clients) => {
  // const counts: { [key: string]: number } = { all: 0 };

  //   for (const clientType in clients) {
  //     let clientTotal = 0;
  //     for (const version in clients[clientType]) {
  //       const { synced, unsynced } = clients[clientType][version];
  //       clientTotal += synced + unsynced;
  //     }
  //     counts[clientType.toLowerCase()] = clientTotal;
  //     counts['all'] += clientTotal;
  //   }

  //   return counts;
  // };

  // const getNetworkCounts = (networkItems: NetworkItems) => {
  //   const counts: { [key: string]: number } = {};
  //   for (const key in networkItems) {
  //     counts[key.toLowerCase()] = networkItems[key].length;
  //   }
  //   counts['all'] = Object.values(networkItems).reduce((sum, clients) => sum + clients.length, 0);
  //   return counts;
  // };

  const allData: Continent = {
    count: continents.reduce((sum, curr) => sum + curr.count, 0),
    continent_code: '',
    continent_name: 'All',
  };

  // const region = [allData.continent_code, ...continents.map((country) => country.continent_code)];
  const regionOptions = [
    allData.continent_name,
    ...continents
      .sort((a, b) => a.continent_name.localeCompare(b.continent_name, undefined, { sensitivity: 'base' }))
      .map((country) => country.continent_name),
  ];
  const regionCounts = continents.reduce(
    (acc, curr) => {
      return {
        ...acc,
        [curr.continent_name.toLowerCase()]: curr.count,
      };
    },
    { [allData.continent_name.toLowerCase()]: allData.count }
  );

  const handleFilterChange = (newRegion: string[], newClients: string[]) => {
    setRegion(newRegion);
    setClients(newClients);
    onFilterChange((prev) => ({ ...prev, region: newRegion, clients: newClients }));
  };

  const handleClearFilters = () => {
    setRegion(['All']);
    setClients(['All']);
    onFilterChange({ region: ['All'], clients: ['All'] });
  };

  const handleDropdownToggle = (dropdownKey: string) => {
    if (openDropdown === dropdownKey) {
      setOpenDropdown(null);
    } else {
      setOpenDropdown(dropdownKey);
    }
  };

  const dropdownSelector = [
    {
      key: 'region',
      label: 'Region',
      value: region,
      options: regionOptions,
      counts: regionCounts,
      onFilterChange: (value: string[]) => handleFilterChange(value, clients),
    },
    {
      key: 'client',
      label: 'Client Version',
      value: clients,
      options: clientsOptions,
      counts: clientCounts,
      onFilterChange: (value: string[]) => handleFilterChange(region, value),
    },
    // {
    //   key: 'network',
    //   label: 'Network',
    //   value: network,
    //   options: networkOptions,
    //   counts: networkCounts,
    //   onFilterChange: (value: string[]) => handleFilterChange(region, client, value),
    // },
  ];
  // using this for mouse hover away in dropdown
  const filterBarRef = React.useRef<HTMLDivElement>(null);
  return (
    <div className="filter-bar">
      <div className="dropdown-group" ref={filterBarRef}>
        {dropdownSelector.map(({ key, label, value, options, counts, onFilterChange }) => (
          <Dropdown
            key={`${key}-dropdown-${value.join(',')}`}
            label={label}
            value={value}
            options={options}
            counts={counts}
            onFilterChange={onFilterChange}
            isOpen={openDropdown === key}
            onClose={() => handleDropdownToggle(key)}
            filterBarRef={filterBarRef}
          />
        ))}
      </div>

      <svg xmlns="http://www.w3.org/2000/svg" width="100%" height="2" viewBox="0 0 400 2" fill="none">
        <path opacity="0.07" d="M0 1H400" stroke="#EBEBEB" />
      </svg>

      <div className="actions-group">
        <button className="clear-filters-button" onClick={handleClearFilters}>
          Clear Filters
        </button>
        <RouterLink to={'/explore'} className="node-explorer-button">
          Node Explorer
        </RouterLink>
      </div>
    </div>
  );
};

export default FilterBar;
