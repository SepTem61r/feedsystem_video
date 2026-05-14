import client from './client'
import type {
  Video,
  PublishVideoRequest,
  DelVideoRequest,
  ListByAuthorIDRequest,
  GetDetailRequest,
  UploadResponse,
  MessageResponse,
} from '../types'

export const videoApi = {
  publish: (data: PublishVideoRequest) =>
    client.post<Video>('/video/publish', data),

  delete: (data: DelVideoRequest) =>
    client.post<MessageResponse>('/video/delete', data),

  listByAuthorID: (data: ListByAuthorIDRequest) =>
    client.post<Video[]>('/video/listByButhorID', data),

  getDetail: (data: GetDetailRequest) =>
    client.post<Video>('/video/getDetail', data),

  uploadVideo: (file: File) => {
    const form = new FormData()
    form.append('file', file)
    return client.post<UploadResponse>('/video/uploadVideo', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 120000,
    })
  },

  uploadCover: (file: File) => {
    const form = new FormData()
    form.append('filename', file)
    return client.post<UploadResponse>('/video/uploadCover', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 60000,
    })
  },
}
