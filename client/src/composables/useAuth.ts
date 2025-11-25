import { computed, readonly, ref, type Ref } from 'vue';

// =============================================================================
// Types
// =============================================================================

interface User {
  id: string;
  email: string;
}

interface TokenResponse {
  access_token: string;
  expires_in: number;
}

interface AuthState {
  user: Ref<User | null>;
  isLoading: Ref<boolean>;
  error: Ref<string | null>;
  isAuthenticated: Ref<boolean>;
}

// =============================================================================
// Configuration
// =============================================================================

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:5000';

const ENDPOINTS = {
  register: `${API_BASE_URL}/api/auth/register`,
  login: `${API_BASE_URL}/api/auth/login`,
  logout: `${API_BASE_URL}/api/auth/logout`,
  refresh: `${API_BASE_URL}/api/auth/refresh`,
  me: `${API_BASE_URL}/api/user/me`,
} as const;

// =============================================================================
// State (module-level singleton)
// =============================================================================

const accessToken = ref<string | null>(null);
const user = ref<User | null>(null);
const isLoading = ref(false);
const error = ref<string | null>(null);

let tokenExpiry: number | null = null;
let refreshTimer: ReturnType<typeof setTimeout> | null = null;

// =============================================================================
// Computed
// =============================================================================

const isAuthenticated = computed(() => {
  return accessToken.value !== null && user.value !== null;
});

// =============================================================================
// Private helpers
// =============================================================================

async function request<T>(
  url: string,
  options: RequestInit = {}
): Promise<{ data: T | null; error: string | null; status: number }> {
  try {
    const response = await fetch(url, {
      ...options,
      credentials: 'include',
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      return {
        data: null,
        error: errorData.error || `Request failed with status ${response.status}`,
        status: response.status,
      };
    }

    const data = await response.json();
    return { data, error: null, status: response.status };
  } catch {
    return { data: null, error: 'Network error. Please try again.', status: 0 };
  }
}

function scheduleTokenRefresh(expiresInSeconds: number): void {
  clearRefreshTimer();

  // Refresh 60 seconds before expiry, minimum 10 seconds
  const delayMs = Math.max((expiresInSeconds - 60) * 1000, 10_000);

  refreshTimer = setTimeout(async () => {
    await refreshTokens();
  }, delayMs);
}

function clearRefreshTimer(): void {
  if (refreshTimer) {
    clearTimeout(refreshTimer);
    refreshTimer = null;
  }
}

function setAuthTokens(tokens: TokenResponse): void {
  accessToken.value = tokens.access_token;
  tokenExpiry = Date.now() + tokens.expires_in * 1000;
  scheduleTokenRefresh(tokens.expires_in);
}

function clearAuthState(): void {
  accessToken.value = null;
  tokenExpiry = null;
  user.value = null;
  error.value = null;
  clearRefreshTimer();
}

// =============================================================================
// Auth operations
// =============================================================================

async function refreshTokens(): Promise<boolean> {
  const { data, error: err } = await request<TokenResponse>(ENDPOINTS.refresh, {
    method: 'POST',
  });

  if (err || !data) {
    clearAuthState();
    return false;
  }

  setAuthTokens(data);
  return true;
}

async function fetchUser(): Promise<User | null> {
  if (!accessToken.value) {
    return null;
  }

  const { data, error: err, status } = await request<User>(ENDPOINTS.me, {
    headers: {
      Authorization: `Bearer ${accessToken.value}`,
    },
  });

  if (status === 401) {
    // Token expired, try refresh
    const refreshed = await refreshTokens();
    if (refreshed) {
      return fetchUser();
    }
    clearAuthState();
    return null;
  }

  if (err || !data) {
    return null;
  }

  user.value = data;
  return data;
}

async function register(email: string, password: string): Promise<boolean> {
  isLoading.value = true;
  error.value = null;

  const { data, error: err } = await request<TokenResponse>(ENDPOINTS.register, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });

  if (err || !data) {
    error.value = err;
    isLoading.value = false;
    return false;
  }

  setAuthTokens(data);
  await fetchUser();
  isLoading.value = false;
  return true;
}

async function login(email: string, password: string): Promise<boolean> {
  isLoading.value = true;
  error.value = null;

  const { data, error: err } = await request<TokenResponse>(ENDPOINTS.login, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });

  if (err || !data) {
    error.value = err;
    isLoading.value = false;
    return false;
  }

  setAuthTokens(data);
  await fetchUser();
  isLoading.value = false;
  return true;
}

async function logout(): Promise<void> {
  await request(ENDPOINTS.logout, { method: 'POST' }).catch(() => {});
  clearAuthState();
}

async function initialize(): Promise<void> {
  // Try to restore session from httpOnly refresh token cookie
  const refreshed = await refreshTokens();
  if (refreshed) {
    await fetchUser();
  }
}

function getAccessToken(): string | null {
  return accessToken.value;
}

// =============================================================================
// Composable
// =============================================================================

export function useAuth() {
  return {
    // State (readonly to prevent external mutations)
    user: readonly(user),
    isLoading: readonly(isLoading),
    error: readonly(error),
    isAuthenticated,

    // Actions
    register,
    login,
    logout,
    initialize,
    getAccessToken,
    fetchUser,
  };
}

// Re-export types
export type { User, AuthState };
