import React from 'react';
import TableComponent from '../TableComponent/TableComponent';
import { Peer } from '../../../../types';
import './NodeList.css';
import { useEffect } from 'react';
import { useAtom } from 'jotai';
import { sortedNodeListDataAtom } from '../../../../hooks/atoms';

interface NodeListProps {
  data: Peer[];
  onRowClick: (node: Peer) => void;
  onSort: (column: string, order: 'asc' | 'desc') => void;
  sortColumn: string;
  sortOrder: 'asc' | 'desc';
  isLoading?: boolean;
}

const NodeList: React.FC<NodeListProps> = ({ data, onRowClick, onSort, sortColumn, sortOrder, isLoading }) => {
  const [sortedData, setSortedData] = useAtom(sortedNodeListDataAtom);

  useEffect(() => {
    const sorted = [...data].sort((a, b) => {
      const aValue = sortColumn.split('.').reduce((o, i) => (o ? (o as any)[i] : undefined), a) ?? '';
      const bValue = sortColumn.split('.').reduce((o, i) => (o ? (o as any)[i] : undefined), b) ?? '';

      if (aValue < bValue) return sortOrder === 'asc' ? -1 : 1;
      if (aValue > bValue) return sortOrder === 'asc' ? 1 : -1;
      return 0;
    });

    setSortedData(sorted);
  }, [data, sortColumn, sortOrder, setSortedData]);

  return (
    <div className="network-stats-list">
      <TableComponent
        data={sortedData}
        onRowClick={onRowClick}
        onSort={onSort}
        sortColumn={sortColumn}
        sortOrder={sortOrder}
        isLoading={isLoading}
      />
    </div>
  );
};

export default NodeList;
