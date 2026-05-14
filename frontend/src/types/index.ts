// ========== Video ==========
export interface Video {
  id: number
  author_id: number
  username: string
  title: string
  create_time: string
  description: string
  play_url: string
  cover_url: string
  likes_count: number
  popularity: number
}

// ========== Feed ==========
export interface FeedAuthor {
  id: number
  username: string
}

export interface FeedVideoItem {
  id: number
  author: FeedAuthor
  title: string
  description?: string
  play_url: string
  cover_url: string
  create_time: number // unix seconds
  likes_count: number
  is_liked: boolean
}

// ========== Account ==========
export interface AccountInfo {
  id: number
  username: string
}

// ========== Comment ==========
export interface Comment {
  id: number
  username: string
  video_id: number
  content: string
  created_at: string
}

// ========== Social ==========
export interface FollowerInfo {
  id: number
  username: string
}

// ========== Request types ==========
export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  password: string
}

export interface ChangePasswordRequest {
  username: string
  old_password: string
  new_password: string
}

export interface FindByIDRequest {
  id: number
}

export interface FindByUsernameRequest {
  username: string
}

export interface RenameRequest {
  new_username: string
}

export interface PublishVideoRequest {
  title: string
  description: string
  play_url: string
  cover_url: string
}

export interface DelVideoRequest {
  id: number
}

export interface ListByAuthorIDRequest {
  author_id: number
}

export interface GetDetailRequest {
  id: number
}

export interface LikeRequest {
  video_id: number
}

export interface CommentRequest {
  video_id: number
  content: string
}

export interface CommentDelRequest {
  comment_id: number
}

export interface CommentGetAllRequest {
  video_id: number
}

export interface FollowRequest {
  vlogger_id: number
}

export interface GetAllFollowersRequest {
  vlogger_id: number
}

export interface GetAllVloggersRequest {
  follower_id: number
}

export interface ListLatestRequest {
  limit: number
  latest_time: number // unix milliseconds, 0 = first page
}

export interface ListLikesCountRequest {
  limit: number
  likes_count_before?: number | null
  id_before?: number | null
}

export interface ListByFollowingRequest {
  limit: number
  latest_time: number // unix seconds, 0 = first page
}

export interface ListByPopularityRequest {
  limit: number
  as_of: number
  offset: number
  latest_id_before?: number | null
  latest_popularity: number
  latest_before: string // ISO time string
}

// ========== Response types ==========
export interface LoginResponse {
  token: string
}

export interface RenameResponse {
  token: string
}

export interface UploadResponse {
  url: string
  cover_url: string
}

export interface IsLikedResponse {
  is_liked: boolean
}

export interface ListMyLikedVideosResponse {
  videos: Video[]
}

export interface ListLatestResponse {
  video_list: FeedVideoItem[]
  next_time: number
  has_more: boolean
}

export interface LikesCountCursor {
  likes_count: number
  id: number
}

export interface ListLikesCountResponse {
  video_list: FeedVideoItem[]
  next_likes_count_before?: number | null
  next_id_before?: number | null
  has_more: boolean
}

export interface ListByFollowingResponse {
  video_list: FeedVideoItem[]
  next_time: number
  has_more: boolean
}

export interface ListByPopularityResponse {
  video_list: FeedVideoItem[]
  as_of: number
  next_offset: number
  has_more: boolean
  next_latest_popularity?: number | null
  next_latest_before?: string | null
  next_latest_id_before?: number | null
}

export interface GetAllFollowersResponse {
  followers: FollowerInfo[]
}

export interface GetAllVloggersResponse {
  vloggers: FollowerInfo[]
}

export interface MessageResponse {
  message: string
}

export interface ErrorResponse {
  error: string
}
