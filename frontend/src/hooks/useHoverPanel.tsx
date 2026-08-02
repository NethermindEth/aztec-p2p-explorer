import { useState, useRef, useCallback } from 'react';

type UseHoverPanelOptions = {
  disableAutoClose?: boolean;
};

export const useHoverPanel = (
  setHoveredMarker: React.Dispatch<React.SetStateAction<any | null>>,
  setIsGlobeRotating?: React.Dispatch<React.SetStateAction<boolean>>,
  options: UseHoverPanelOptions = {}
) => {
  const [isInfoPanelHovered, setIsInfoPanelHovered] = useState<boolean>(false);
  const hoverTimeout = useRef<ReturnType<typeof setTimeout>>();
  const isInfoPanelHoveredRef = useRef(isInfoPanelHovered);
  const disableAutoClose = Boolean(options.disableAutoClose);

  const handlePointerOut = useCallback(() => {
    if (disableAutoClose) return;
    hoverTimeout.current = setTimeout(() => {
      if (!isInfoPanelHoveredRef.current) {
        setHoveredMarker(null);
        if (setIsGlobeRotating) {
          setIsGlobeRotating(true);
        }
      }
    }, 300);
  }, [setHoveredMarker, setIsGlobeRotating, disableAutoClose]);

  const handlePointerOver = useCallback(
    (label: any) => {
      if (hoverTimeout.current) {
        clearTimeout(hoverTimeout.current);
      }
      setHoveredMarker(label);
      if (setIsGlobeRotating) {
        setIsGlobeRotating(false);
      }
    },
    [setHoveredMarker, setIsGlobeRotating]
  );

  const handleInfoPanelHover = useCallback(
    (isHovered: boolean) => {
      if (hoverTimeout.current) {
        clearTimeout(hoverTimeout.current);
      }
      setIsInfoPanelHovered(isHovered);
      isInfoPanelHoveredRef.current = isHovered;

      if (!isHovered) {
        if (disableAutoClose) return;
        hoverTimeout.current = setTimeout(() => {
          setHoveredMarker(null);
          if (setIsGlobeRotating) {
            setIsGlobeRotating(true);
          }
        }, 300);
      }
    },
    [setHoveredMarker, setIsGlobeRotating, disableAutoClose]
  );

  const forceCloseInfoPanel = useCallback(() => {
    if (hoverTimeout.current) {
      clearTimeout(hoverTimeout.current);
    }
    setIsInfoPanelHovered(false);
    isInfoPanelHoveredRef.current = false;
    setHoveredMarker(null);
    if (setIsGlobeRotating) {
      setIsGlobeRotating(true);
    }
  }, [setHoveredMarker, setIsGlobeRotating]);

  return {
    isInfoPanelHovered,
    handlePointerOut,
    handlePointerOver,
    handleInfoPanelHover,
    forceCloseInfoPanel,
  };
};
