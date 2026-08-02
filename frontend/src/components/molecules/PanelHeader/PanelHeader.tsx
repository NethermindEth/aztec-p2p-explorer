import './PanelHeader.css';
import { useAtom } from 'jotai';
import { Tooltip } from '@radix-ui/themes';
import { InfoCircledIcon } from '@radix-ui/react-icons';
import { totalActivePeers } from '../../../hooks/atoms';
import nodeCountIcon from '../../../assets/icons/nodeCount.svg';
import Skeleton from '../Skeleton/Skeleton';
import PanelSection from '../PanelSection/PanelSection';
import { useAnalytics } from '../../../api/analytics';

const syncStatusTooltip = (
  <span>
    <strong>Sync Status</strong> shows how many peers in the latest crawl are caught up with the network&apos;s chain
    head.
    <br />
    <br />
    For each peer we compare its reported latest block to a reference Aztec node and classify it as:
    <br />• <strong>Synced</strong> — peer matches the chain head
    <br />• <strong>Unsynced</strong> — peer is behind, has a mismatched block hash, or its sync state couldn&apos;t be
    confirmed
    <br />
    <br />
    <em>Example:</em> if 90 of 100 peers are synced, the synced bar shows 90.0%.
  </span>
);

const disableSyncStatus = import.meta.env.VITE_DISABLE_SYNC_STATUS === 'true';

const HeaderSection: React.FC = () => {
  const [nodeCount] = useAtom(totalActivePeers);
  const { data: analyticsData } = useAnalytics();

  const syncStatusData = analyticsData?.sync_status
    ? (() => {
        const synced = analyticsData.sync_status.synced ?? 0;
        const unknown = analyticsData.sync_status.not_synced ?? 0;
        const total = synced + unknown;

        return [
          {
            label: 'synced',
            value: synced.toString(),
            percentage: total > 0 ? `${((synced / total) * 100).toFixed(1)}%` : '0%',
          },
          {
            label: 'unsynced',
            value: unknown.toString(),
            percentage: total > 0 ? `${((unknown / total) * 100).toFixed(1)}%` : '0%',
          },
        ];
      })()
    : [];

  return (
    <div className={'panel-header-container'}>
      <div className="top-section">
        <div className="title">
          Node Count
          <img src={nodeCountIcon} alt="node count" />
        </div>
        <div className="node-count">
          {nodeCount === undefined ? <Skeleton height="20px" width="70px" /> : nodeCount.toLocaleString()}
        </div>
      </div>

      {!disableSyncStatus && (
        <div className={'sync-container'}>
          <div className="sync-status-heading">
            <p>Sync Status</p>
            <Tooltip content={syncStatusTooltip}>
              <InfoCircledIcon className="sync-status-info-icon" aria-label="How sync status is calculated" />
            </Tooltip>
          </div>
          <div>
            <PanelSection items={syncStatusData} isSync={true} />
          </div>
        </div>
      )}
    </div>
  );
};

export default HeaderSection;
