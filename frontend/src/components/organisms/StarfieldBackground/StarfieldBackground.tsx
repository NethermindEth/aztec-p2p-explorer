import { useEffect, useRef } from 'react';
import './StarfieldBackground.css';

const StarfieldBackground = () => {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    let targetX = 0;
    let targetY = 0;
    let currentX = 0;
    let currentY = 0;
    let animationFrameId: number = 0;
    let isPaused = false;

    const lerp = (start: number, end: number, factor: number) => {
      return start + (end - start) * factor;
    };

    const applyDeadZone = (value: number, threshold: number) => {
      if (Math.abs(value) < threshold) return 0;
      return value - Math.sign(value) * threshold;
    };

    const animate = () => {
      // Don't animate if paused (tab hidden)
      if (isPaused) {
        animationFrameId = 0;
        return;
      }

      currentX = lerp(currentX, targetX, 0.03);
      currentY = lerp(currentY, targetY, 0.03);

      // Check if values have converged (animation should stop)
      const deltaX = Math.abs(currentX - targetX);
      const deltaY = Math.abs(currentY - targetY);
      const threshold = 0.001;

      if (deltaX < threshold && deltaY < threshold) {
        // Animation has settled, stop the loop to save resources
        container.style.setProperty('--mouse-x', targetX.toFixed(3));
        container.style.setProperty('--mouse-y', targetY.toFixed(3));
        animationFrameId = 0;
        return;
      }

      // Update CSS variables
      if (Math.abs(currentX - parseFloat(container.style.getPropertyValue('--mouse-x') || '0')) > 0.001) {
        container.style.setProperty('--mouse-x', currentX.toFixed(3));
        container.style.setProperty('--mouse-y', currentY.toFixed(3));
      }

      animationFrameId = requestAnimationFrame(animate);
    };

    const handleMouseMove = (e: MouseEvent) => {
      let normalizedX = (e.clientX / window.innerWidth) * 2 - 1;
      let normalizedY = (e.clientY / window.innerHeight) * 2 - 1;

      normalizedX = applyDeadZone(normalizedX, 0.2);
      normalizedY = applyDeadZone(normalizedY, 0.2);

      targetX = normalizedX;
      targetY = normalizedY;

      // Restart animation loop if it was stopped and we're not paused
      if (animationFrameId === 0 && !isPaused) {
        animationFrameId = requestAnimationFrame(animate);
      }
    };

    let lastMoveTime = 0;
    const throttledMouseMove = (e: MouseEvent) => {
      const now = Date.now();
      // Increased throttle from 50ms to 100ms for better performance
      if (now - lastMoveTime > 100) {
        lastMoveTime = now;
        handleMouseMove(e);
      }
    };

    // Handle page visibility changes to pause animation in background tabs
    const handleVisibilityChange = () => {
      if (document.hidden) {
        // Tab is hidden, pause animation
        isPaused = true;
        if (animationFrameId !== 0) {
          cancelAnimationFrame(animationFrameId);
          animationFrameId = 0;
        }
      } else {
        // Tab is visible, resume animation if needed
        isPaused = false;
        const deltaX = Math.abs(currentX - targetX);
        const deltaY = Math.abs(currentY - targetY);
        if ((deltaX > 0.001 || deltaY > 0.001) && animationFrameId === 0) {
          animationFrameId = requestAnimationFrame(animate);
        }
      }
    };

    // Initialize
    container.style.setProperty('--mouse-x', '0');
    container.style.setProperty('--mouse-y', '0');

    window.addEventListener('mousemove', throttledMouseMove, { passive: true });
    document.addEventListener('visibilitychange', handleVisibilityChange);

    // Start animation loop
    animationFrameId = requestAnimationFrame(animate);

    return () => {
      window.removeEventListener('mousemove', throttledMouseMove);
      document.removeEventListener('visibilitychange', handleVisibilityChange);
      if (animationFrameId !== 0) {
        cancelAnimationFrame(animationFrameId);
      }
    };
  }, []);

  return (
    <div className="starfield-container" ref={containerRef}>
      <div className="starfield-layer starfield-layer-1" />
      <div className="starfield-layer starfield-layer-2" />
      <div className="starfield-vignette" />
    </div>
  );
};

export default StarfieldBackground;
