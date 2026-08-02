import React, { useEffect, useState } from 'react';
// import NodeStatsBlock from '../../molecules/PanelNodeDetails/NodeStatsBlock/NodeStatsBlock';
import DetailsTable from '../../molecules/PanelNodeDetails/DetailsTable/DetailsTable';
import './PanelNodeDetails.css';
import { Peer } from '../../../types';
import nodeVersionIcon from '../../../assets/icons/tag.svg';
import clientTypeIcon from '../../../assets/icons/mix.svg';
import locationIcon from '../../../assets/icons/globe.svg';
import { Button } from '@radix-ui/themes';
import { useAtom } from 'jotai';
import { selectedNodeAtom } from '../../../hooks/atoms';
import { Cross2Icon } from '@radix-ui/react-icons';

interface PanelNodeDetailsProps {
  node: Peer;
}
interface NodeDetail {
  label: string;
  value: string;
  icon: string;
}

const PanelNodeDetails: React.FC<PanelNodeDetailsProps> = ({ node }) => {
  const [, setNodeDetails] = useState<NodeDetail[]>([]);
  const [, setSelectedNode] = useAtom(selectedNodeAtom);
  const closePanel = () => {
    setSelectedNode(null);
  };

  useEffect(() => {
    const details: NodeDetail[] = [
      {
        label: 'Version',
        value: node?.client?.split('/')?.[1] ?? 'Unknown',
        icon: nodeVersionIcon,
      },
      {
        label: 'Client type',
        value: node?.client?.split('/')?.[0] ?? 'Unknown',
        icon: clientTypeIcon,
      },
      {
        label: 'Location',
        value: node?.multi_addresses?.[0]?.ip_info?.[0]?.country_name ?? 'Unknown',
        icon: locationIcon,
      },
    ];

    setNodeDetails(details);
  }, [node]);

  return (
    <div className="node-details-container">
      <div className="node-details-header">
        <span>Node Stats</span>
        <Button variant="ghost" color="gray" className="modal-btn" onClick={closePanel}>
          <Cross2Icon />
        </Button>
      </div>
      <DetailsTable node={node} />
    </div>
  );
};

export default PanelNodeDetails;
