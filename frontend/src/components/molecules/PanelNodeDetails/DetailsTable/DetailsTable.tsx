import React from 'react';
import './DetailsTable.css';
import RowItem from '../../../atoms/NodeDetailsRowItem/RowItem';
import SyncStatusLabel from '../../../atoms/SyncStatusLabel/SyncStatusLabel';
import { formatDateTime } from '../../../../utils';
import { Peer } from '../../../../types';

const disableSyncStatus = import.meta.env.VITE_DISABLE_SYNC_STATUS === 'true';

interface DetailsTableProps {
  node: Peer | null;
}

const DetailsTable: React.FC<DetailsTableProps> = ({ node }) => {
  return (
    <div className="details-table">
      <div className="detail-row">
        <RowItem title="Version" content={node?.client} />
        {!disableSyncStatus && (
          <RowItem
            title="Status"
            content={
              <SyncStatusLabel
                isSynced={node?.is_synced === true}
                isUnknown={node?.is_synced === null || node?.is_synced === undefined}
              />
            }
          />
        )}
      </div>
      <div className="detail-row">
        <RowItem
          title="Location"
          content={
            (node?.multi_addresses?.[0]?.ip_info?.[0]?.country_name || '') +
              (node?.multi_addresses?.[0]?.ip_info?.[0]?.city_name
                ? (node?.multi_addresses?.[0]?.ip_info?.[0]?.country_name ? ', ' : '') +
                  node.multi_addresses[0].ip_info[0].city_name
                : '') || 'Unknown'
          }
        />
        {/* <RowItem title="Client type" content={node?.client?.split('/')?.[0] ?? 'Unknown'} /> */}
        {/* <RowItem title="Network type" content={node.multi_addresses[0].ip_info[0].as_name} /> */}
      </div>
      {/* <div className="detail-row">
        <RowItem title="Network type" content={node.multi_addresses[0].ip_info[0].as_name} />
        <RowItem title="Block height" content={node.block_height} />
      </div> */}
      <div className="detail-row">
        <RowItem title="Latitude" content={node?.multi_addresses[0].ip_info[0].latitude} />
        <RowItem title="Longitude" content={node?.multi_addresses[0].ip_info[0].longitude} />
      </div>
      <div className="detail-row">
        <RowItem title="First seen" content={node ? formatDateTime(node.created_at) : null} />
        <RowItem title="Last seen" content={node ? formatDateTime(node.last_seen) : null} />
      </div>
      <div className="detail-row">
        <RowItem title="Peer ID" content={node?.id} />
      </div>
    </div>
  );
};

export default DetailsTable;
