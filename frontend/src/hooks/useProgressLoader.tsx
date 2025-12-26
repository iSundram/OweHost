import { useState, useCallback } from 'react';
import { LoadingBar } from '../components/ui';

export function useProgressLoader() {
  const [loadingState, setLoadingState] = useState<{
    isLoading: boolean;
    message?: string;
    showOverlay?: boolean;
  }>({
    isLoading: false,
    message: 'Loading...',
    showOverlay: false,
  });

  const showLoading = useCallback((message?: string, showOverlay?: boolean) => {
    setLoadingState({
      isLoading: true,
      message: message || 'Loading...',
      showOverlay: showOverlay || false,
    });
  }, []);

  const hideLoading = useCallback(() => {
    setLoadingState(prev => ({
      ...prev,
      isLoading: false,
    }));
  }, []);

  const ProgressLoader = useCallback(() => (
    <LoadingBar 
      isLoading={loadingState.isLoading}
      message={loadingState.message}
      showOverlay={loadingState.showOverlay}
    />
  ), [loadingState]);

  return {
    showLoading,
    hideLoading,
    ProgressLoader,
    isLoading: loadingState.isLoading,
    message: loadingState.message,
  };
}

// Example usage in components:
/*
import { useProgressLoader } from '../hooks/useProgressLoader';

function MyComponent() {
  const { showLoading, hideLoading, ProgressLoader } = useProgressLoader();

  const handleAsyncOperation = async () => {
    showLoading('Processing your request...', true); // with overlay
    try {
      await someAsyncOperation();
    } finally {
      hideLoading();
    }
  };

  return (
    <div>
      <ProgressLoader />
      <button onClick={handleAsyncOperation}>Start Operation</button>
    </div>
  );
}
*/