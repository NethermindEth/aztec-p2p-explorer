import React from 'react';
import './SyncStatusLabel.css';

interface SyncStatusLabelProps {
  isSynced?: boolean;
  isUnknown?: boolean;
  count?: string | number;
  percentage?: string;
}

const SyncStatusLabel: React.FC<SyncStatusLabelProps> = ({ isSynced, isUnknown, count, percentage }) => {
  const label = isUnknown ? 'Unknown' : isSynced ? 'Synced' : 'Syncing';
  const symbolClass = isUnknown ? 'unknown-symbol' : isSynced ? 'sync-symbol' : 'not-synced-symbol';
  const percentageClass = isUnknown
    ? 'sync-status-percentage-unknown'
    : isSynced
      ? 'sync-status-percentage-synced'
      : 'sync-status-percentage-syncing';

  return (
    <div className="sync-status-label-container">
      <div className="sync-status-item">
        <div className="status-text">{label}</div>
        {count !== undefined && (
          <>
            <div className="sync-status-item-separator">-</div>
            <div className="panel-section-item-value">{Number(count).toLocaleString()}</div>
          </>
        )}
        <div className={symbolClass}></div>
      </div>
      {percentage && (
        <div className={percentageClass}>
          <span className="percentage-value">{percentage}</span>
        </div>
      )}
    </div>
  );
};

export default SyncStatusLabel;
