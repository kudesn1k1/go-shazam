export const MAX_AVATAR_BYTES = 2 * 1024 * 1024; // keep in sync with FILE_MAX_UPLOAD_BYTES

export type SortField = 'title' | 'artist' | 'duration' | 'created_at';
export type SortOrder = 'asc' | 'desc';

export interface SongFilterQuery {
  q?: string;
  artist?: string;
  uploaded_by?: string;
  created_after?: string;
  created_before?: string;
  sort?: SortField;
  order?: SortOrder;
  page?: number;
  limit?: number;
}

export interface SongListItem {
  id: string;
  title: string;
  artist: string;
  duration: number;
  source_id: string;
  uploaded_by?: string;
  created_at: string; // RFC3339 UTC
}

export interface PublicSong {
  id: string;
  title: string;
  artist: string;
  duration: number;
  source_id: string;
  created_at: string;
}

export interface UserWithAvatar {
  id: string;
  email: string;
  roles: string[];
  created_at: string;
  avatar_url: string | null;
}

export interface FileUploadResponse {
  hash: string;
  content_type: string;
  size_bytes: number;
}

export interface AvatarResponse {
  avatar_url: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}
