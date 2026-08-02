import React from 'react';
import './PanelSection.css';
import { useLocation } from 'react-router-dom';
import Skeleton from '../Skeleton/Skeleton';
import SyncStatusLabel from '../../atoms/SyncStatusLabel/SyncStatusLabel';

interface PanelSectionProps {
  title?: string;
  items?: { label: string; value: string; percentage?: string }[];
  isSync?: boolean;
}

const PanelSection: React.FC<PanelSectionProps> = ({ title, items = [], isSync }) => {
  const location = useLocation();
  const isListView = location.pathname === '/explore';

  const syncStatus = {
    synced: items.find((item) => item.label === 'synced') || { value: '0', percentage: '0%' },
    notSynced: items.find((item) => item.label === 'not_synced') || { value: '0', percentage: '0%' },
    unsynced: items.find((item) => item.label === 'unsynced') || { value: '0', percentage: '0%' },
    unknown: items.find((item) => item.label === 'unknown') || { value: '0', percentage: '0%' },
  };

  const formatPercentage = (percentage: string | undefined) => {
    if (!percentage) return null;
    const value = parseFloat(percentage.replace('%', ''));
    return !isNaN(value) ? `${value}%` : null;
  };

  const calculateCombinedPercentage = (value1: string = '0', value2: string = '0') => {
    const percent1 = parseFloat(value1.replace('%', ''));
    const percent2 = parseFloat(value2.replace('%', ''));
    const sum = percent1 + percent2;
    return !isNaN(sum) ? `${sum}%` : null;
  };

  const SyncedNodes = () => {
    const unknownCount = parseInt(syncStatus.unknown.value);
    return (
      <div className="panel-section-sync">
        <SyncStatusLabel
          isSynced={true}
          count={syncStatus.synced.value}
          percentage={formatPercentage(syncStatus.synced.percentage) || undefined}
        />
        <SyncStatusLabel
          isSynced={false}
          count={parseInt(syncStatus.notSynced.value) + parseInt(syncStatus.unsynced.value)}
          percentage={
            calculateCombinedPercentage(syncStatus.notSynced.percentage, syncStatus.unsynced.percentage) || undefined
          }
        />
        {unknownCount > 0 && (
          <SyncStatusLabel
            isUnknown={true}
            count={syncStatus.unknown.value}
            percentage={formatPercentage(syncStatus.unknown.percentage) || undefined}
          />
        )}
      </div>
    );
  };

  const SyncingNodes = () => {
    return (
      <div className="panel-section-data">
        {items.slice(0, isListView ? 4 : items.length).map((item, index) => (
          <div className="panel-section-item" key={index}>
            <div className="panel-section-item-label">{item.label === '' ? 'Unknown' : item.label}</div>
            <div className="panel-section-item-separator">-</div>
            <div className="panel-section-item-value">{Number(item.value ?? 0).toLocaleString()}</div>
            {formatPercentage(item.percentage) && (
              <div className="panel-section-item-percentage">
                <span className="percentage-value">{formatPercentage(item.percentage)}</span>
              </div>
            )}
          </div>
        ))}
      </div>
    );
  };

  const Loaders = ({ count = 3, className }: { count?: number; className?: string }) => {
    return (
      <div className={className}>
        {Array(count)
          .fill(null)
          .map((_, idx) => (
            <div key={idx} className="panel-section-loaders">
              <Skeleton width={'140px'} height={'18px'} radius={'6px'} ml="1rem" />
            </div>
          ))}
      </div>
    );
  };

  if (isSync) {
    return (
      <div>
        {items && items.length > 0 && isSync ? (
          <SyncedNodes />
        ) : (
          <Loaders count={2} className="panel-section-loader-container" />
        )}
      </div>
    );
  }

  return (
    <div className={isListView ? 'panel-section-container-listview' : 'panel-section-container'}>
      {title && <div className="panel-section-title">{title}</div>}
      {items && items.length > 0 ? isSync ? <SyncedNodes /> : <SyncingNodes /> : <Loaders />}
    </div>
  );
};

export default PanelSection;
