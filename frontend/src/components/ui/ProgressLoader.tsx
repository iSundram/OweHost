import { useEffect, useState } from 'react';

interface ProgressLoaderProps {
  isLoading: boolean;
  message?: string;
  showOverlay?: boolean;
  overlayOpacity?: number;
}

export function ProgressLoader({ 
  isLoading, 
  message = 'Loading...', 
  showOverlay = true,
  overlayOpacity = 0.8
}: ProgressLoaderProps) {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted || !isLoading) return null;

  return (
    <>
      {/* Progress Bar */}
      <div className="owehost-progress-bar-container">
        <div className="owehost-progress-bar-inner" />
      </div>
      
      {/* Loading Overlay */}
      {showOverlay && (
        <div 
          className="owehost-loading-overlay"
          style={{ backgroundColor: `rgba(255, 255, 255, ${overlayOpacity})` }}
        >
          <div className="owehost-loading-content">
            <div className="owehost-spinner" />
            {message && <p className="owehost-loading-message">{message}</p>}
          </div>
        </div>
      )}
    </>
  );
}

// Context-based Progress Loader for global usage
interface ProgressContextType {
  showLoading: (message?: string, showOverlay?: boolean) => void;
  hideLoading: () => void;
}

export const useProgressLoader = (): ProgressContextType => {
  const [isLoading, setIsLoading] = useState(false);
  const [message, setMessage] = useState('Loading...');
  const [showOverlay, setShowOverlay] = useState(true);

  const showLoading = (msg?: string, overlay?: boolean) => {
    setMessage(msg || 'Loading...');
    setShowOverlay(overlay !== false);
    setIsLoading(true);
  };

  const hideLoading = () => {
    setIsLoading(false);
  };

  return { showLoading, hideLoading };
};

export function GlobalProgressLoader({ 
  isLoading, 
  message, 
  showOverlay 
}: { 
  isLoading: boolean; 
  message?: string; 
  showOverlay?: boolean;
}) {
  return <ProgressLoader isLoading={isLoading} message={message} showOverlay={showOverlay} />;
}