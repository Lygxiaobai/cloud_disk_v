import httpClient, { rawClient } from "@/api/http";
import type {
  LoginRequest,
  LoginResponse,
  MailCodeRequest,
  RefreshTokenRequest,
  RefreshTokenResponse,
  RegisterRequest,
  UserDetailResponse,
} from "@/types/api";

export async function login(payload: LoginRequest): Promise<LoginResponse> {
  const { data } = await rawClient.post<LoginResponse>("/user/login", payload);
  return data;
}

export async function sendRegisterCode(payload: MailCodeRequest): Promise<void> {
  await rawClient.post("/mail/code/send/register", payload);
}

export async function register(payload: RegisterRequest): Promise<void> {
  await rawClient.post("/user/register", payload);
}

export async function refreshToken(payload: RefreshTokenRequest): Promise<RefreshTokenResponse> {
  const { data } = await rawClient.put<RefreshTokenResponse>("/refresh/token", payload);
  return data;
}

export async function fetchCurrentUser(): Promise<UserDetailResponse> {
  const { data } = await httpClient.get<UserDetailResponse>("/user/detail");
  return data;
}
