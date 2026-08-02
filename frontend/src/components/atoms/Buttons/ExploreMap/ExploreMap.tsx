import { useAtom } from 'jotai';
import { showListViewAtom, selectedNodeAtom, selectedRowAtom } from '../../../../hooks/atoms';
import './ExploreMap.css';
import { Link } from 'react-router-dom';

const ExploreMap: React.FC = () => {
  const [, setShowListView] = useAtom(showListViewAtom);
  const [, setSelectedNode] = useAtom(selectedNodeAtom);
  const [, setSelectedRow] = useAtom(selectedRowAtom);

  const handleClick = () => {
    // Clear state before navigation
    setSelectedNode(null);
    setSelectedRow(null);
    setShowListView(false);

    // Scroll after navigation
    setTimeout(() => {
      const scrollContainer = document.querySelector('.main-container');
      if (scrollContainer) {
        scrollContainer.scrollTo({ top: 0, behavior: 'smooth' });
      }
    }, 100);
  };

  return (
    <div className="map-view-button-container">
      <Link to="/" className="map-view-button" onClick={handleClick}>
        <span className="explore-map-text">Explore Map</span>
      </Link>
    </div>
  );
};

export default ExploreMap;
