import client from './client'
import type { LikeRequest, IsLikedResponse, ListMyLikedVideosResponse, MessageResponse } from '../types'

export const likeApi = {
  like: (data: LikeRequest) =>
    client.post<MessageResponse>('/like/like', data),

  unlike: (data: LikeRequest) =>
    client.post<MessageResponse>('/like/unlike', data),

  isLiked: (data: LikeRequest) =>
    client.post<IsLikedResponse>('/like/isLiked', data),

  listMyLikedVideos: () =>
    client.post<ListMyLikedVideosResponse>('/like/listMyLikedVideos'),
}
