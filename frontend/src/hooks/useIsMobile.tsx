import { useAtom } from 'jotai';
import { isMobileAtom, isTabletAtom } from './atoms';
import { useEffect } from 'react';

export function useIsMobile(debounceTime = 100) {
  const [, setIsMobile] = useAtom(isMobileAtom);
  const [, setIsTablet] = useAtom(isTabletAtom);

  useEffect(() => {
    if (typeof window === 'undefined' || !('matchMedia' in window)) return;

    // Adjacent ranges: mobile < 768, tablet 768–1024
    // Also consider landscape orientation on smaller screens as mobile
    const mqMobile = window.matchMedia('(max-width: 767.99px)');
    const mqLandscape = window.matchMedia('(max-width: 932px) and (orientation: landscape)');
    const mqTablet = window.matchMedia('(min-width: 768px) and (max-width: 1024px)');

    let timeoutId: number | null = null;

    const apply = () => {
      setIsMobile(mqMobile.matches || mqLandscape.matches);
      setIsTablet(mqTablet.matches && !mqLandscape.matches);
    };

    const onChange = () => {
      if (timeoutId !== null) clearTimeout(timeoutId);
      timeoutId = window.setTimeout(apply, debounceTime);
    };

    // init
    apply();

    mqMobile.addEventListener?.('change', onChange) ?? mqMobile.addListener(onChange as any);
    mqLandscape.addEventListener?.('change', onChange) ?? mqLandscape.addListener(onChange as any);
    mqTablet.addEventListener?.('change', onChange) ?? mqTablet.addListener(onChange as any);

    return () => {
      if (timeoutId !== null) clearTimeout(timeoutId);
      mqMobile.removeEventListener?.('change', onChange) ?? mqMobile.removeListener(onChange as any);
      mqLandscape.removeEventListener?.('change', onChange) ?? mqLandscape.removeListener(onChange as any);
      mqTablet.removeEventListener?.('change', onChange) ?? mqTablet.removeListener(onChange as any);
    };
  }, [setIsMobile, setIsTablet, debounceTime]);
}
