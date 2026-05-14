import client from './client'
import type { CommentRequest, CommentDelRequest, CommentGetAllRequest, Comment, MessageResponse } from '../types'

export const commentApi = {
  publish: (data: CommentRequest) =>
    client.post<MessageResponse>('/comment/publish', data),

  delete: (data: CommentDelRequest) =>
    client.post<MessageResponse>('/comment/delete', data),

  getAll: (data: CommentGetAllRequest) =>
    client.post<Comment[]>('/comment/getAll', data),
}
