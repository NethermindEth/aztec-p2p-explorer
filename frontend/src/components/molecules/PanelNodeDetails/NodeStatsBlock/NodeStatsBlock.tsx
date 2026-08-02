import React from 'react';
import { useAtom } from 'jotai';
import { selectedNodeAtom } from '../../../../hooks/atoms';
import './NodeStatsBlock.css';

const NodeStatsBlock: React.FC = () => {
  const [, setSelectedNode] = useAtom(selectedNodeAtom);
  const handleBackClick = () => {
    setSelectedNode(null);
  };

  return (
    <div className="node-stats-block">
      <div className="title">Node Stats</div>
      <div className="back-button" onClick={handleBackClick} style={{ border: '1px solid red' }}></div>
    </div>
  );
};

export default NodeStatsBlock;
