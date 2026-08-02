import React, { useEffect, useMemo, useState } from 'react';
import { Table } from '@radix-ui/themes';
import { Peer } from '../../../../types';
import { CaretSortIcon, CaretDownIcon, CaretUpIcon } from '@radix-ui/react-icons';
import { useAtom } from 'jotai';
import { isMobileAtom, selectedNodeAtom, selectedRowAtom } from '../../../../hooks/atoms';
import { truncateMiddleWithBudget } from '../../../../utils';

import './TableComponent.css';

interface TableComponentProps {
  data: Peer[];
  onRowClick: (node: Peer) => void;
  onSort: (column: string, order: 'asc' | 'desc') => void;
  sortColumn: string;
  sortOrder: 'asc' | 'desc'; // TODO: use enum instead but it will require changes in other places as well
  isLoading?: boolean;
}

const getSortIcon = (column: string, sortColumn: string, sortOrder: 'asc' | 'desc') => {
  if (column !== sortColumn) return <CaretSortIcon />;
  return sortOrder === 'asc' ? <CaretUpIcon /> : <CaretDownIcon />;
};

enum ColumnNames {
  NodeID = 'Peer ID',
  Client = 'Client',
  InSync = 'In Sync',
}

// @note each column name is mapped to the key of the peer object
const columnMapping: Record<ColumnNames, string> = {
  [ColumnNames.NodeID]: 'id',
  [ColumnNames.Client]: 'client',
  [ColumnNames.InSync]: 'is_synced',
};

const TableComponent: React.FC<TableComponentProps> = ({
  data,
  onRowClick,
  onSort,
  sortColumn,
  sortOrder,
  isLoading,
}) => {
  const [selectedRow, setSelectedRow] = useAtom(selectedRowAtom);
  const [selectedNode] = useAtom(selectedNodeAtom);
  const [isMobile] = useAtom(isMobileAtom);

  const [viewportWidth, setViewportWidth] = useState<number>(typeof window !== 'undefined' ? window.innerWidth : 0);

  useEffect(() => {
    if (!isMobile) return;
    const onResize = () => setViewportWidth(window.innerWidth);
    window.addEventListener('resize', onResize);
    return () => window.removeEventListener('resize', onResize);
  }, [isMobile]);

  const maxCharsMobile = useMemo(() => {
    if (!isMobile) return Infinity;
    const w = viewportWidth || (typeof window !== 'undefined' ? window.innerWidth : 0);
    if (w <= 360) return 18;
    if (w <= 420) return 20;
    if (w <= 480) return 22;
    if (w <= 560) return 24;
    if (w <= 640) return 26;
    return 28;
  }, [isMobile, viewportWidth]);

  const handleRowClick = (row: Peer) => {
    setSelectedRow(row);
    onRowClick(row);
  };

  const handleSort = (column: string) => {
    const order = sortColumn === column && sortOrder === 'asc' ? 'desc' : 'asc';
    onSort(column, order);
  };

  // Renders list of peers in the Node List tab
  const renderCells = (row: Peer) => {
    type CellData = {
      key: string;
      value: string | number;
      style?: React.CSSProperties;
    };

    const cellData: CellData[] = [
      {
        key: 'id',
        value: isMobile ? truncateMiddleWithBudget(row.id, maxCharsMobile) : row.id,
        style: { textTransform: 'lowercase' },
      },
      {
        key: 'client',
        value: row.client || 'Unknown',
        style: isMobile
          ? ({
              whiteSpace: 'nowrap',
              width: '1%',
              maxWidth: '120px',
            } as React.CSSProperties)
          : undefined,
      },
      // { key: 'is_synced', value: row.is_synced ? 'Yes' : 'No' },
    ];

    return cellData.map((cell, index) => (
      <Table.Cell key={index} className="network-stats__table-cell" style={cell.style || {}}>
        {cell.value}
      </Table.Cell>
    ));
  };

  const columns = ['Peer ID', 'Client'] as const;

  return (
    <div className="network-stats__table-container fade-in">
      <Table.Root className="network-stats__table" layout="auto" size="2">
        <Table.Header className="network-stats__table-header">
          <Table.Row className="network-stats__table-row">
            {columns.map((column) => (
              <Table.ColumnHeaderCell
                key={column}
                className="network-stats__table-header-cell"
                onClick={() => handleSort(columnMapping[column])}
              >
                <div>
                  <span>{column}</span>
                  <div>{getSortIcon(columnMapping[column], sortColumn, sortOrder)}</div>
                </div>
              </Table.ColumnHeaderCell>
            ))}
          </Table.Row>
        </Table.Header>

        <Table.Body className="network-stats__table-body">
          {isLoading ? (
            <Table.Row>
              <Table.Cell colSpan={columns.length} className="no_results">
                Loading...
              </Table.Cell>
            </Table.Row>
          ) : !data || data.length < 1 ? (
            <Table.Row>
              <Table.Cell colSpan={columns.length} className="no_results">
                No results found
              </Table.Cell>
            </Table.Row>
          ) : (
            data.map((row, index) => (
              <Table.Row
                key={index}
                className={`network-stats__table-row ${selectedRow === row && selectedNode ? 'selected' : ''}`}
                onClick={() => handleRowClick(row)}
              >
                {renderCells(row)}
              </Table.Row>
            ))
          )}
        </Table.Body>
      </Table.Root>
    </div>
  );
};

export default TableComponent;
