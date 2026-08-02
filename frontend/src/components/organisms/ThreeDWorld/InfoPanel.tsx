import React, { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';
import './InfoPanel.css';
import '../../../globalStyles.css';
import { useAtom } from 'jotai';
import { truncateNodeID } from '../../../utils';
import { RawMarker } from '../../../utils/clusterMarkers';
import { isMobileAtom, nodeModalAtom } from '../../../hooks/atoms';
import { useSearchParams } from 'react-router-dom';
import { Button } from '@radix-ui/themes';
import { Cross2Icon } from '@radix-ui/react-icons';
import LoadingSpinner from '../../molecules/LoadingSpinner/LoadingSpinner';

type InfoPanelProps = {
  data: RawMarker[];
  x: number;
  y: number;
  onHover: (isHovered: boolean) => void;
  onClose: () => void;
};

const infoIcon = (
  <svg
    className="default-icon"
    xmlns="http://www.w3.org/2000/svg"
    width="20"
    height="21"
    viewBox="0 0 20 21"
    fill="none"
  >
    <path
      d="M9.99192 17.0833V16H15.2434C15.3075 16 15.3663 15.9733 15.4196 15.9198C15.4731 15.8665 15.4998 15.8077 15.4998 15.7435V5.25646C15.4998 5.19229 15.4731 5.13354 15.4196 5.08021C15.3663 5.02674 15.3075 5 15.2434 5H9.99192V3.91667H15.2434C15.6185 3.91667 15.9356 4.04618 16.1946 4.30521C16.4537 4.56424 16.5832 4.88132 16.5832 5.25646V15.7435C16.5832 16.1187 16.4537 16.4358 16.1946 16.6948C15.9356 16.9538 15.6185 17.0833 15.2434 17.0833H9.99192ZM9.0288 13.391L8.24671 12.6329L9.83796 11.0417H3.4165V9.95833H9.83796L8.24671 8.36708L9.0288 7.60896L11.9196 10.5L9.0288 13.391Z"
      fill="#8E8C99"
    />
  </svg>
);

const hoverIcon = (
  <svg className="hover-icon" xmlns="http://www.w3.org/2000/svg" width="20" height="21" viewBox="0 0 20 21" fill="none">
    <path
      d="M9.99192 17.0833V16H15.2434C15.3075 16 15.3663 15.9733 15.4196 15.9198C15.4731 15.8665 15.4998 15.8077 15.4998 15.7435V5.25646C15.4998 5.1923 15.4731 5.13355 15.4196 5.08021C15.3663 5.02674 15.3075 5.00001 15.2434 5.00001H9.99192V3.91667H15.2434C15.6185 3.91667 15.9356 4.04619 16.1946 4.30521C16.4537 4.56424 16.5832 4.88132 16.5832 5.25646V15.7435C16.5832 16.1187 16.4537 16.4358 16.1946 16.6948C15.9356 16.9538 15.6185 17.0833 15.2434 17.0833H9.99192ZM9.0288 13.391L8.24671 12.6329L9.83796 11.0417H3.4165V9.95834H9.83796L8.24671 8.36709L9.0288 7.60896L11.9196 10.5L9.0288 13.391Z"
      fill="#A3FAE0"
    />
  </svg>
);

const InfoPanel: React.FC<InfoPanelProps> = ({ data, x, y, onHover, onClose }) => {
  const [modalAtom, setModalAtom] = useAtom(nodeModalAtom);
  const [isMobile] = useAtom(isMobileAtom);
  const [searchParams, setSearchParams] = useSearchParams();

  const infoPanelRef = useRef<HTMLDivElement>(null);
  const contentRef = useRef<HTMLDivElement>(null);
  const [computedPos, setComputedPos] = useState<{ top: number; left: number } | null>(null);
  const openedAtRef = useRef<number>(0);
  const loadingMoreRef = useRef<boolean>(false);

  type NodeListItem = { id: string; city?: string | null; country?: string | null };
  const flattenedNodes: NodeListItem[] = useMemo(
    () => data.flatMap((item) => item.peer_ids.map((id) => ({ id, city: item.city, country: item.country }))),
    [data]
  );

  const INITIAL_COUNT = 60;
  const BATCH_SIZE = 120;
  const [visibleCount, setVisibleCount] = useState<number>(INITIAL_COUNT);

  useEffect(() => {
    // Reset pagination when dataset changes
    setVisibleCount(INITIAL_COUNT);
    loadingMoreRef.current = false;
  }, [flattenedNodes.length]);

  useEffect(() => {
    // Record open time to suppress the initial backdrop click from the tap that opened the panel
    openedAtRef.current = Date.now();
  }, []);

  if (!data) return null;
  const isSingleNode = data.length === 1 && data[0].peer_ids.length === 0;

  const handleScroll = (e: React.WheelEvent<HTMLDivElement>) => {
    if (infoPanelRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = infoPanelRef.current;

      const atTop = scrollTop === 0;
      const atBottom = scrollTop + clientHeight === scrollHeight;

      // Prevent scrolling the parent element when InfoPanel is at the top or bottom
      if ((atTop && e.deltaY < 0) || (atBottom && e.deltaY > 0)) {
        e.preventDefault();
      }

      // Stop event propagation to prevent parent scroll
      e.stopPropagation();
    }
  };

  // Render a single node and defaults to Unknown if there is no data for a value (e.g. city, country, etc.)
  const renderSingleNode = (nodeData: RawMarker) => {
    const infoItems = [
      {
        title: 'Node Info',
        data: `${nodeData?.peer_ids[0] || 'Unknown'}`,
        subData: `${nodeData?.city || 'Unknown'}, ${nodeData?.country || 'Unknown'}`,
      },
    ];

    return (
      <div className="single-node-container" onClick={() => handleOnClick(infoItems[0].data)}>
        {infoItems.map((item, index) => (
          <div className="info-item" key={index}>
            <div className="info-text">
              <div className="info-text-single-title">{item.title}</div>
              <div className="info-text-single-data">{truncateNodeID(item.data)}</div>
              {item.subData && <div className="info-text-single-data">{item.subData}</div>}
            </div>
          </div>
        ))}
      </div>
    );
  };

  function handleOnClick(nodeId: string) {
    const next = new URLSearchParams(searchParams);
    next.set('id', nodeId);
    setSearchParams(next, { replace: true });
    setModalAtom({
      isOpen: true,
      nodeId: nodeId,
      node: null,
    });
  }

  function renderMultipleNodes(items: NodeListItem[]) {
    const slice = items.slice(0, visibleCount);
    return slice.map(({ id, city, country }) => (
      <div key={id} className="info-item" onClick={() => handleOnClick(id)}>
        <div className="info-text">
          <div className="info-text-ip">{truncateNodeID(id || 'Unknown')}</div>
          <div className="info-text-city">
            {city || 'Unknown'}, {country || 'Unknown'}
          </div>
        </div>
        <div className="info-icon">
          {infoIcon}
          {hoverIcon}
        </div>
      </div>
    ));
  }

  const nodeCount = data.reduce((sum, curr) => sum + curr.count, 0);

  // Compute edge-aware desktop position; mobile uses CSS positioning
  useLayoutEffect(() => {
    if (isMobile) {
      return; // mobile handled by CSS
    }

    const recalc = () => {
      const panelEl = infoPanelRef.current;
      if (!panelEl) {
        setComputedPos({ top: Math.max(10, y - 35), left: Math.max(10, x - 35) });
        return;
      }

      // The panel is absolutely positioned within the scene container via the backdrop.
      // Convert viewport-relative pointer coordinates (clientX/clientY) to container-relative.
      const backdropEl = panelEl.parentElement as HTMLElement | null; // .info-panel-backdrop
      const containerRect = backdropEl?.getBoundingClientRect();
      if (!containerRect) {
        setComputedPos({ top: Math.max(10, y - 35), left: Math.max(10, x - 35) });
        return;
      }

      const { offsetWidth: panelWidth, offsetHeight: panelHeight } = panelEl;
      const GAP = 10;

      // Coordinates relative to the container hosting the panel
      const relX = x - containerRect.left;
      const relY = y - containerRect.top;

      const maxLeft = Math.max(GAP, containerRect.width - panelWidth - GAP);
      const maxTop = Math.max(GAP, containerRect.height - panelHeight - GAP);

      // Prefer right of the cursor; flip to left if overflowing the container
      let left = relX + GAP;
      if (left + panelWidth > containerRect.width - GAP) {
        left = Math.max(GAP, relX - panelWidth - GAP);
      }
      // Clamp within container safe area
      left = Math.min(Math.max(left, GAP), maxLeft);

      // Prefer below the cursor; flip upward if overflowing the container
      let top = relY + GAP;
      if (top + panelHeight > containerRect.height - GAP) {
        top = Math.max(GAP, relY - panelHeight - GAP);
      }
      // Clamp within container safe area
      top = Math.min(Math.max(top, GAP), maxTop);

      setComputedPos({ top, left });
    };

    recalc();

    // Recompute on resize and when panel size changes
    const ro = new ResizeObserver(() => recalc());
    if (infoPanelRef.current) {
      ro.observe(infoPanelRef.current);
    }
    const handleResize = () => recalc();
    window.addEventListener('resize', handleResize);

    return () => {
      ro.disconnect();
      window.removeEventListener('resize', handleResize);
    };
  }, [x, y, isMobile, data]);

  const maybeLoadMore = useCallback(() => {
    if (!contentRef.current) return;
    if (loadingMoreRef.current) return;
    const el = contentRef.current;
    const threshold = 64; // px from bottom to trigger
    if (el.scrollTop + el.clientHeight >= el.scrollHeight - threshold) {
      if (visibleCount >= flattenedNodes.length) return;
      loadingMoreRef.current = true;
      // Defer to allow frame rendering before heavy DOM work
      setTimeout(() => {
        setVisibleCount((prev) => Math.min(prev + BATCH_SIZE, flattenedNodes.length));
        loadingMoreRef.current = false;
      }, 0);
    }
  }, [visibleCount, flattenedNodes.length]);

  const desktopStyles = useMemo(() => {
    if (isMobile) return {} as React.CSSProperties;
    const top = computedPos?.top ?? y - 35;
    const left = computedPos?.left ?? x - 35;
    return {
      '--desktop-top': `${top}px`,
      '--desktop-left': `${left}px`,
    } as React.CSSProperties;
  }, [computedPos, isMobile, x, y]);

  const handleBackdropClick = (e: React.MouseEvent) => {
    // Only close on mobile when clicking the backdrop
    if (e.target === e.currentTarget) {
      // Ignore the very first click that often is the same tap used to open the panel
      if (Date.now() - openedAtRef.current < 250) return;
      onClose();
    }
  };

  const handleClose = () => {
    onClose();
  };

  return (
    <>
      {/* Mobile backdrop */}
      <div className="info-panel-backdrop" onClick={handleBackdropClick}>
        <div
          className={`info-panel ${isSingleNode ? 'single-node' : ''} custom-scrollbar`}
          style={desktopStyles}
          onMouseEnter={() => !isMobile && onHover(true)}
          onMouseLeave={() => {
            if (isMobile) return;
            // Keep list open while node modal is open
            if (modalAtom?.isOpen) return;
            onHover(false);
          }}
          onWheel={handleScroll}
          ref={infoPanelRef}
          onClick={(e) => e.stopPropagation()}
        >
          {/* Header with close button */}
          <div className="info-panel-header">
            {!isSingleNode && <div className="info-panel-node-count">{nodeCount} nodes here</div>}
            <Button variant="ghost" color="gray" className="info-panel-close-btn modal-btn" onClick={handleClose}>
              <Cross2Icon />
            </Button>
          </div>

          {/* Content */}
          <div className="info-panel-content" ref={contentRef} onScroll={maybeLoadMore}>
            {isSingleNode ? (
              renderSingleNode(data[0])
            ) : (
              <div className="info-item-container">
                {renderMultipleNodes(flattenedNodes)}
                {visibleCount < flattenedNodes.length && (
                  <div style={{ display: 'flex', justifyContent: 'center', padding: '0.75rem 0' }}>
                    <LoadingSpinner />
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </>
  );
};

export default InfoPanel;
