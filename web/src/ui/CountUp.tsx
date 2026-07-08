import React, { useEffect, useState } from 'react';

interface CountUpProps {
  end: number;
  duration?: number;
  className?: string;
  prefix?: string;
  suffix?: string;
  separator?: string;
}

export const CountUp: React.FC<CountUpProps> = ({
  end,
  duration = 1500,
  className = '',
  prefix = '',
  suffix = '',
  separator = ',',
}) => {
  const [count, setCount] = useState(0);

  useEffect(() => {
    let startTimestamp: number | null = null;
    const step = (timestamp: number) => {
      if (!startTimestamp) startTimestamp = timestamp;
      const progress = Math.min((timestamp - startTimestamp) / duration, 1);
      // easeOutExpo
      const ease = progress === 1 ? 1 : 1 - Math.pow(2, -10 * progress);
      setCount(Math.floor(ease * end));
      if (progress < 1) {
        window.requestAnimationFrame(step);
      }
    };
    window.requestAnimationFrame(step);
  }, [end, duration]);

  const formattedValue = count.toString().replace(/\B(?=(\d{3})+(?!\d))/g, separator);

  return (
    <span className={className}>
      {prefix}
      {formattedValue}
      {suffix}
    </span>
  );
};
