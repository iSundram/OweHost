import { useState } from 'react';
import type { FormEvent } from 'react';
import { Navigate, Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { Button, Input } from '../components/ui';
import { Lock, User, Eye, EyeOff } from 'lucide-react';
import type { UserRole } from '../types';

function getDashboardRoute(role?: UserRole): string {
  switch (role) {
    case 'admin':
      return '/admin';
    case 'reseller':
      return '/reseller';
    case 'user':
    default:
      return '/user';
  }
}

export function LoginPage() {
  const { login, isAuthenticated, isLoading, user } = useAuth();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');

  if (isAuthenticated && user) {
    return <Navigate to={getDashboardRoute(user.role)} replace />;
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    
    if (!username || !password) {
      setError('Please enter both username and password');
      return;
    }

    try {
      await login({ username, password });
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Invalid username or password';
      setError(errorMessage);
      console.error('Login error:', err);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-[var(--color-background)] p-4">
      {/* Background Pattern */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute -top-40 -right-40 w-80 h-80 bg-[#7BA4D0] rounded-full opacity-10 blur-3xl"></div>
        <div className="absolute -bottom-40 -left-40 w-80 h-80 bg-[#E7F0FA] rounded-full opacity-10 blur-3xl"></div>
      </div>

      <div className="w-full max-w-md relative z-10">
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-[#7BA4D0] to-[#E7F0FA] shadow-2xl mb-4">
            <span className="text-3xl font-bold text-white">O</span>
          </div>
          <h1 className="text-3xl font-bold text-[var(--color-text-primary)]">OweHost</h1>
          <p className="text-[var(--color-text-secondary)] mt-2">Web Hosting Panel</p>
        </div>

        {/* Login Card */}
        <div className="bg-[var(--color-surface)] rounded-2xl p-8 shadow-2xl border border-[var(--color-border-light)]">
          <h2 className="text-xl font-semibold text-[var(--color-text-primary)] mb-6">
            Sign in to your account
          </h2>

          {error && (
            <div className="mb-4 p-3 rounded-lg bg-[var(--color-error)]/10 border border-[var(--color-error)]/20 text-[var(--color-error)] text-sm">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-5">
            <Input
              label="Username"
              type="text"
              placeholder="Enter your username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              leftIcon={<User size={18} />}
              autoComplete="username"
            />

            <Input
              label="Password"
              type={showPassword ? 'text' : 'password'}
              placeholder="Enter your password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              leftIcon={<Lock size={18} />}
              rightIcon={
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="hover:text-[var(--color-text-primary)] transition-colors"
                >
                  {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                </button>
              }
              autoComplete="current-password"
            />

            <div className="flex items-center justify-between text-sm">
              <label className="flex items-center gap-2 text-[var(--color-text-secondary)]">
                <input
                  type="checkbox"
                  className="w-4 h-4 rounded border-[var(--color-border)] bg-[var(--color-primary-dark)] text-[#E7F0FA] focus:ring-[#E7F0FA]"
                />
                Remember me
              </label>
              <Link
                to="/forgot-password"
                className="text-[#E7F0FA] hover:text-[var(--color-text-primary)]"
              >
                Forgot password?
              </Link>
            </div>

            <Button
              type="submit"
              variant="primary"
              size="lg"
              className="w-full"
              isLoading={isLoading}
            >
              Sign in
            </Button>
          </form>

          <div className="mt-6 text-center">
            <p className="text-sm text-[var(--color-text-muted)]">
              Don't have an account?{' '}
              <Link to="/register" className="text-[#E7F0FA] hover:text-[var(--color-text-primary)] font-medium">
                Contact administrator
              </Link>
            </p>
          </div>
        </div>

        {/* Footer */}
        <p className="text-center mt-8 text-sm text-[var(--color-text-muted)]">
          © 2024 OweHost. All rights reserved.
        </p>
      </div>
    </div>
  );
}
