import React from 'react';
import './NodeExplorer.css';
import RouterLink from '../../RouterLink/RouterLink';

interface NodeExplorerProps {
  onShowNetworkStats: () => void;
}

const NodeExplorer: React.FC<NodeExplorerProps> = ({ onShowNetworkStats }) => {
  const handleClick = () => {
    onShowNetworkStats();
  };

  return (
    <RouterLink to={'/explore'} className="node-explorer-button" onClick={handleClick}>
      Node Explorer
    </RouterLink>
  );
};

export default NodeExplorer;
