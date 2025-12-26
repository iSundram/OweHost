import { useState, useEffect } from 'react';
import { Settings, CheckCircle, XCircle, ToggleLeft, ToggleRight } from 'lucide-react';
import { Button, Card, CardContent, CardHeader } from '../../components/ui';
import { featureService, type FeatureCategory, type Feature } from '../../api/services';

export function AdminFeaturesPage() {
  const [categories, setCategories] = useState<FeatureCategory[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [updatingFeatures, setUpdatingFeatures] = useState<Set<string>>(new Set());

  useEffect(() => {
    loadFeatures();
  }, []);

  const loadFeatures = async () => {
    setIsLoading(true);
    try {
      const data = await featureService.list();
      setCategories(data);
    } catch (error) {
      console.error('Failed to load features:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleToggleFeature = async (featureName: string, currentEnabled: boolean) => {
    setUpdatingFeatures((prev) => new Set(prev).add(featureName));
    try {
      await featureService.update(featureName, !currentEnabled);
      // Reload features to get updated state
      await loadFeatures();
    } catch (error) {
      console.error('Failed to update feature:', error);
      alert('Failed to update feature');
    } finally {
      setUpdatingFeatures((prev) => {
        const next = new Set(prev);
        next.delete(featureName);
        return next;
      });
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-[var(--color-text-secondary)]">Loading features...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">Feature Manager</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">
            Enable or disable system features and capabilities
          </p>
        </div>
      </div>

      {/* Feature Categories */}
      <div className="space-y-4">
        {categories.map((category) => (
          <Card key={category.name}>
            <CardHeader>
              <div className="flex items-center gap-2">
                <div className="p-2 rounded-lg bg-gradient-to-br from-[#7BA4D0]/30 to-[#E7F0FA]/20">
                  <Settings size={20} className="text-[#E7F0FA]" />
                </div>
                <div>
                  <h3 className="font-semibold text-[var(--color-text-primary)]">
                    {category.display_name}
                  </h3>
                  <p className="text-xs text-[var(--color-text-secondary)]">{category.name}</p>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                {category.features.map((feature) => {
                  const isUpdating = updatingFeatures.has(feature.name);
                  return (
                    <div
                      key={feature.name}
                      className="flex items-center justify-between p-3 rounded-lg bg-[var(--color-primary-dark)] border border-[var(--color-border)]"
                    >
                      <div className="flex items-center gap-3 flex-1 min-w-0">
                        {feature.enabled ? (
                          <CheckCircle size={18} className="text-[var(--color-success)] flex-shrink-0" />
                        ) : (
                          <XCircle size={18} className="text-[var(--color-text-muted)] flex-shrink-0" />
                        )}
                        <div className="min-w-0 flex-1">
                          <div className="font-medium text-sm text-[var(--color-text-primary)] truncate">
                            {feature.display_name}
                          </div>
                          <div className="text-xs text-[var(--color-text-secondary)] truncate">
                            {feature.description}
                          </div>
                        </div>
                      </div>
                      <button
                        onClick={() => handleToggleFeature(feature.name, feature.enabled)}
                        disabled={isUpdating}
                        className="ml-3 flex-shrink-0 disabled:opacity-50 disabled:cursor-not-allowed"
                        title={feature.enabled ? 'Disable feature' : 'Enable feature'}
                      >
                        {feature.enabled ? (
                          <ToggleRight size={24} className="text-[var(--color-success)]" />
                        ) : (
                          <ToggleLeft size={24} className="text-[var(--color-text-muted)]" />
                        )}
                      </button>
                    </div>
                  );
                })}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Summary Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Card>
          <CardContent className="flex items-center gap-4 py-4">
            <div className="p-3 rounded-xl bg-[var(--color-success)]/10">
              <CheckCircle size={24} className="text-[var(--color-success)]" />
            </div>
            <div>
              <p className="text-2xl font-bold text-[var(--color-text-primary)]">
                {categories.reduce(
                  (sum, cat) => sum + cat.features.filter((f) => f.enabled).length,
                  0
                )}
              </p>
              <p className="text-sm text-[var(--color-text-secondary)]">Enabled Features</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-4 py-4">
            <div className="p-3 rounded-xl bg-[var(--color-text-muted)]/10">
              <XCircle size={24} className="text-[var(--color-text-muted)]" />
            </div>
            <div>
              <p className="text-2xl font-bold text-[var(--color-text-primary)]">
                {categories.reduce(
                  (sum, cat) => sum + cat.features.filter((f) => !f.enabled).length,
                  0
                )}
              </p>
              <p className="text-sm text-[var(--color-text-secondary)]">Disabled Features</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-4 py-4">
            <div className="p-3 rounded-xl bg-gradient-to-br from-[#7BA4D0]/30 to-[#E7F0FA]/20">
              <Settings size={24} className="text-[#E7F0FA]" />
            </div>
            <div>
              <p className="text-2xl font-bold text-[var(--color-text-primary)]">
                {categories.reduce((sum, cat) => sum + cat.features.length, 0)}
              </p>
              <p className="text-sm text-[var(--color-text-secondary)]">Total Features</p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
