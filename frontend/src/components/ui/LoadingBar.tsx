import { useEffect, useState } from 'react';
import './LoadingBar.css';

interface LoadingBarProps {
  isLoading: boolean;
}

export function LoadingBar({ isLoading }: LoadingBarProps) {
  const [show, setShow] = useState(false);

  useEffect(() => {
    if (isLoading) {
      setShow(true);
    } else {
      // Delay hiding to allow animation to complete
      const timer = setTimeout(() => setShow(false), 300);
      return () => clearTimeout(timer);
    }
  }, [isLoading]);

  if (!show) return null;

  return (
    <>
      <div className="loading-progress-bar">
        <div className="loading-progress-bar-inner" />
      </div>
      {isLoading && <div className="loading-overlay" />}
    </>
  );
}
