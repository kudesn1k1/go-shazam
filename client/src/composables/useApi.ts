import { useAuth } from './useAuth';
import type {
  SongFilterQuery,
  SongListItem,
  PublicSong,
  UserWithAvatar,
  FileUploadResponse,
  AvatarResponse,
  PaginatedResponse,
} from '../types/api';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:5000';

function buildQuery(params: object): string {
  const parts: string[] = [];
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === '') continue;
    parts.push(`${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`);
  }
  return parts.length ? `?${parts.join('&')}` : '';
}

async function authFetch<T>(
  path: string,
  options: RequestInit = {}
): Promise<{ data: T | null; error: string | null; status: number }> {
  const { getAccessToken } = useAuth();
  const token = getAccessToken();
  const isFormData = options.body instanceof FormData;

  try {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      ...options,
      credentials: 'include',
      headers: {
        ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...(options.headers ?? {}),
      },
    });

    if (response.status === 204) {
      return { data: null, error: null, status: 204 };
    }

    if (!response.ok) {
      const err = await response.json().catch(() => ({}));
      return { data: null, error: err.error ?? `Request failed with status ${response.status}`, status: response.status };
    }

    const data = await response.json();
    return { data, error: null, status: response.status };
  } catch {
    return { data: null, error: 'Network error. Please try again.', status: 0 };
  }
}

export function useApi() {
  function listMySongs(filter: SongFilterQuery) {
    return authFetch<PaginatedResponse<SongListItem>>(`/api/user/songs${buildQuery(filter)}`);
  }
  function listPublicSongs(filter: SongFilterQuery) {
    return authFetch<PaginatedResponse<PublicSong>>(`/api/public/songs${buildQuery(filter)}`);
  }
  function getPublicSong(id: string) {
    return authFetch<PublicSong>(`/api/public/songs/${id}`);
  }
  function listAllSongs(filter: SongFilterQuery) {
    return authFetch<PaginatedResponse<SongListItem>>(`/api/songs${buildQuery(filter)}`);
  }
  function listUserSongs(id: string, filter: SongFilterQuery) {
    return authFetch<PaginatedResponse<SongListItem>>(`/api/users/${id}/songs${buildQuery(filter)}`);
  }
  function listUsers(page = 1, limit = 20) {
    return authFetch<PaginatedResponse<UserWithAvatar>>(`/api/users?page=${page}&limit=${limit}`);
  }
  function getUser(id: string) {
    return authFetch<UserWithAvatar>(`/api/users/${id}`);
  }
  function deleteSong(id: string) {
    return authFetch(`/api/songs/${id}`, { method: 'DELETE' });
  }
  function updateUserRoles(id: string, roles: string[]) {
    return authFetch(`/api/users/${id}/roles`, {
      method: 'POST',
      body: JSON.stringify({ roles }),
    });
  }

  // Files + avatars
  function uploadFile(file: File) {
    const form = new FormData();
    form.append('file', file);
    return authFetch<FileUploadResponse>('/api/files', { method: 'POST', body: form });
  }
  function setOwnAvatar(file_hash: string) {
    return authFetch<AvatarResponse>('/api/user/me/avatar', {
      method: 'PUT',
      body: JSON.stringify({ file_hash }),
    });
  }
  function clearOwnAvatar() {
    return authFetch('/api/user/me/avatar', { method: 'DELETE' });
  }
  function setUserAvatar(id: string, file_hash: string) {
    return authFetch<AvatarResponse>(`/api/users/${id}/avatar`, {
      method: 'PUT',
      body: JSON.stringify({ file_hash }),
    });
  }
  function clearUserAvatar(id: string) {
    return authFetch(`/api/users/${id}/avatar`, { method: 'DELETE' });
  }

  return {
    listMySongs, listAllSongs, listUserSongs, listPublicSongs, getPublicSong,
    listUsers, getUser,
    deleteSong, updateUserRoles,
    uploadFile, setOwnAvatar, clearOwnAvatar, setUserAvatar, clearUserAvatar,
  };
}

export { API_BASE_URL };
export type { SongFilterQuery, SongListItem, PublicSong, UserWithAvatar, FileUploadResponse, AvatarResponse, PaginatedResponse };
