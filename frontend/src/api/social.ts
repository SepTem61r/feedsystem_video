import client from './client'
import type {
  FollowRequest,
  GetAllFollowersRequest, GetAllFollowersResponse,
  GetAllVloggersRequest, GetAllVloggersResponse,
  MessageResponse,
} from '../types'

export const socialApi = {
  follow: (data: FollowRequest) =>
    client.post<MessageResponse>('/social/follow', data),

  unfollow: (data: FollowRequest) =>
    client.post<MessageResponse>('/social/unfollow', data),

  getAllFollowers: (data: GetAllFollowersRequest) =>
    client.post<GetAllFollowersResponse>('/social/getallfollowers', data),

  getAllVloggers: (data: GetAllVloggersRequest) =>
    client.post<GetAllVloggersResponse>('/social/getallvloggers', data),
}
