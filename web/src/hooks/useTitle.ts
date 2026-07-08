import { useEffect } from 'react';

export function useTitle(title: string) {
  useEffect(() => {
    const prevTitle = document.title;
    document.title = `${title} | AITriage Enterprise`;
    return () => {
      document.title = prevTitle;
    };
  }, [title]);
}
