import { useAuth } from './useAuth';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:5000';

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}

export interface Song {
  id: string;
  title: string;
  artist: string;
  duration: number;
  source_id: string;
  uploaded_by?: string;
}

export interface UserItem {
  id: string;
  email: string;
  roles: string[];
  created_at: string;
}

async function authFetch<T>(
  path: string,
  options: RequestInit = {}
): Promise<{ data: T | null; error: string | null; status: number }> {
  const { getAccessToken } = useAuth();
  const token = getAccessToken();

  try {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      ...options,
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...(options.headers ?? {}),
      },
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

export function useApi() {
  function getMySongs(page = 1, limit = 20) {
    return authFetch<PaginatedResponse<Song>>(`/api/user/songs?page=${page}&limit=${limit}`);
  }

  function getUsers(page = 1, limit = 20) {
    return authFetch<PaginatedResponse<UserItem>>(`/api/users?page=${page}&limit=${limit}`);
  }

  function getUserByID(id: string) {
    return authFetch<UserItem>(`/api/users/${id}`);
  }

  function getUserSongs(id: string, page = 1, limit = 20) {
    return authFetch<PaginatedResponse<Song>>(`/api/users/${id}/songs?page=${page}&limit=${limit}`);
  }

  function getAllSongs(page = 1, limit = 20) {
    return authFetch<PaginatedResponse<Song>>(`/api/songs?page=${page}&limit=${limit}`);
  }

  function deleteSong(id: string) {
    return authFetch(`/api/songs/${id}`, {
      method: 'DELETE',
    });
  }

  function updateUserRoles(id: string, roles: string[]) {
    return authFetch(`/api/users/${id}/roles`, {
      method: 'POST',
      body: JSON.stringify({ roles }),
    });
  }

  return { getMySongs, getUsers, getUserByID, getUserSongs, getAllSongs, deleteSong, updateUserRoles };
}
