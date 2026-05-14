import client from './client'
import type {
  LoginRequest, LoginResponse,
  RegisterRequest,
  ChangePasswordRequest,
  FindByIDRequest, FindByUsernameRequest, AccountInfo,
  RenameRequest, RenameResponse,
  MessageResponse,
} from '../types'

export const accountApi = {
  login: (data: LoginRequest) =>
    client.post<LoginResponse>('/account/login', data),

  register: (data: RegisterRequest) =>
    client.post<MessageResponse>('/account/register', data),

  changePassword: (data: ChangePasswordRequest) =>
    client.post<MessageResponse>('/account/changePassword', data),

  findByID: (data: FindByIDRequest) =>
    client.post<AccountInfo>('/account/findByID', data),

  findByUsername: (data: FindByUsernameRequest) =>
    client.post<AccountInfo>('/account/findByUsername', data),

  logout: () =>
    client.post<MessageResponse>('/account/logout'),

  rename: (data: RenameRequest) =>
    client.post<RenameResponse>('/account/rename', data),
}
