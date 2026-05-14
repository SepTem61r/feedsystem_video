import client from './client'
import type {
  ListLatestRequest, ListLatestResponse,
  ListLikesCountRequest, ListLikesCountResponse,
  ListByFollowingRequest, ListByFollowingResponse,
  ListByPopularityRequest, ListByPopularityResponse,
} from '../types'

export const feedApi = {
  listLatest: (data: ListLatestRequest) =>
    client.post<ListLatestResponse>('/feed/listLatest', data),

  listLikesCount: (data: ListLikesCountRequest) =>
    client.post<ListLikesCountResponse>('/feed/listLikesCount', data),

  listByFollowing: (data: ListByFollowingRequest) =>
    client.post<ListByFollowingResponse>('/feed/listByFollowing', data),

  listByPopularity: (data: ListByPopularityRequest) =>
    client.post<ListByPopularityResponse>('/feed/listByPopularity', data),
}
