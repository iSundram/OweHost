import { useState } from 'react';
import { Button, LoadingBar } from '../components/ui';
import { useProgressLoader } from '../hooks/useProgressLoader';

export function DemoProgressPage() {
  const { showLoading, hideLoading, ProgressLoader } = useProgressLoader();
  const [customLoading, setCustomLoading] = useState(false);

  const simulateAsyncOperation = async (duration: number, message: string) => {
    showLoading(message, true); // With overlay
    await new Promise(resolve => setTimeout(resolve, duration));
    hideLoading();
  };

  const simulateQuickOperation = async () => {
    showLoading('Quick task...', false); // Without overlay
    await new Promise(resolve => setTimeout(resolve, 2000));
    hideLoading();
  };

  const simulateLongOperation = async () => {
    showLoading('Processing large files...', true);
    await new Promise(resolve => setTimeout(resolve, 5000));
    hideLoading();
  };

  const testStandaloneLoadingBar = () => {
    setCustomLoading(true);
    setTimeout(() => setCustomLoading(false), 3000);
  };

  return (
    <div className="min-h-screen bg-[var(--color-background)] p-8">
      <ProgressLoader />
      
      {/* Standalone Loading Bar for demonstration */}
      <LoadingBar 
        isLoading={customLoading} 
        message="Standalone loading bar demo..." 
        showOverlay={true}
      />

      <div className="max-w-4xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-[var(--color-text-primary)] mb-2">
            Progress Loader Demo
          </h1>
          <p className="text-[var(--color-text-secondary)]">
            Test different loading scenarios with the new OweHost progress loader
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Quick Operations */}
          <div className="bg-white p-6 rounded-lg border border-[var(--color-border)]">
            <h2 className="text-xl font-semibold text-[var(--color-text-primary)] mb-4">
              Quick Operations
            </h2>
            <div className="space-y-3">
              <Button 
                onClick={simulateQuickOperation}
                variant="secondary"
                className="w-full"
              >
                Quick Task (2s)
              </Button>
              <Button 
                onClick={() => simulateAsyncOperation(1500, 'Loading data...')}
                variant="outline"
                className="w-full"
              >
                Load Data (1.5s)
              </Button>
              <Button 
                onClick={testStandaloneLoadingBar}
                variant="ghost"
                className="w-full"
              >
                Standalone Loading Bar (3s)
              </Button>
            </div>
          </div>

          {/* Medium Operations */}
          <div className="bg-white p-6 rounded-lg border border-[var(--color-border)]">
            <h2 className="text-xl font-semibold text-[var(--color-text-primary)] mb-4">
              Medium Operations
            </h2>
            <div className="space-y-3">
              <Button 
                onClick={() => simulateAsyncOperation(3000, 'Processing request...')}
                variant="primary"
                className="w-full"
              >
                Process Request (3s)
              </Button>
              <Button 
                onClick={() => simulateAsyncOperation(4000, 'Validating input...')}
                variant="secondary"
                className="w-full"
              >
                Validate Input (4s)
              </Button>
            </div>
          </div>

          {/* Long Operations */}
          <div className="bg-white p-6 rounded-lg border border-[var(--color-border)]">
            <h2 className="text-xl font-semibold text-[var(--color-text-primary)] mb-4">
              Long Operations
            </h2>
            <div className="space-y-3">
              <Button 
                onClick={simulateLongOperation}
                variant="primary"
                className="w-full"
              >
                Process Large Files (5s)
              </Button>
              <Button 
                onClick={() => simulateAsyncOperation(8000, 'Generating report...')}
                variant="secondary"
                className="w-full"
              >
                Generate Report (8s)
              </Button>
            </div>
          </div>

          {/* Error Simulation */}
          <div className="bg-white p-6 rounded-lg border border-[var(--color-border)]">
            <h2 className="text-xl font-semibold text-[var(--color-text-primary)] mb-4">
              Error Scenarios
            </h2>
            <div className="space-y-3">
              <Button 
                onClick={() => simulateAsyncOperation(2000, 'Network timeout...')}
                variant="danger"
                className="w-full"
              >
                Simulate Network Error
              </Button>
              <Button 
                onClick={() => simulateAsyncOperation(1000, 'Server maintenance...')}
                variant="outline"
                className="w-full"
              >
                Simulate Server Error
              </Button>
            </div>
          </div>
        </div>

        {/* Progress Bar Variants */}
        <div className="mt-8 bg-white p-6 rounded-lg border border-[var(--color-border)]">
          <h2 className="text-xl font-semibold text-[var(--color-text-primary)] mb-4">
            Progress Bar Variants
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="text-center">
              <h3 className="font-medium text-[var(--color-text-primary)] mb-2">Default</h3>
              <div className="h-12 bg-gray-100 rounded flex items-center justify-center">
                <span className="text-sm text-gray-500">4px height</span>
              </div>
            </div>
            <div className="text-center">
              <h3 className="font-medium text-[var(--color-text-primary)] mb-2">Mini</h3>
              <div className="h-12 bg-gray-100 rounded flex items-center justify-center">
                <span className="text-sm text-gray-500">2px height</span>
              </div>
            </div>
            <div className="text-center">
              <h3 className="font-medium text-[var(--color-text-primary)] mb-2">Animation</h3>
              <div className="h-12 bg-gray-100 rounded flex items-center justify-center">
                <span className="text-sm text-gray-500">1.5s loop</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}