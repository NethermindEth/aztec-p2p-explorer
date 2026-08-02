import React from 'react';
import PanelHeader from '../../molecules/PanelHeader/PanelHeader';
import PanelListBody from '../../molecules/PanelListBody/PanelListBody';

import './PanelMain.css';

const PanelMain: React.FC = () => {
  return (
    <div className="panel-main-container">
      <div className="header">
        <PanelHeader />
      </div>
      <div className="body">
        <PanelListBody view="MapView" />
      </div>
    </div>
  );
};

export default PanelMain;
