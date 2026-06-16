import { useState, useEffect } from 'react';

export const useCountUp = (end: number | undefined | null, duration: number = 1000) => {
  const [count, setCount] = useState(0);

  useEffect(() => {
    if (end === undefined || end === null || isNaN(end)) {
      return;
    }

    let startTimestamp: number | null = null;
    const startVal = 0;

    let animationFrameId: number;

    const step = (timestamp: number) => {
      if (!startTimestamp) startTimestamp = timestamp;
      const progress = Math.min((timestamp - startTimestamp) / duration, 1);

      // Cubic ease-out
      const easedProgress = 1 - Math.pow(1 - progress, 3);

      setCount(Math.floor(easedProgress * (end - startVal) + startVal));

      if (progress < 1) {
        animationFrameId = window.requestAnimationFrame(step);
      } else {
        setCount(end);
      }
    };

    animationFrameId = window.requestAnimationFrame(step);

    return () => {
      window.cancelAnimationFrame(animationFrameId);
    };
  }, [end, duration]);

  return count;
};
