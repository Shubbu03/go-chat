import axios, { AxiosInstance, AxiosResponse } from "axios";
import { notify } from "./notify";

export interface AuthResponse {
  message: string;
  status: string;
  access_token: string;
  refresh_token?: string;
  expires_in: number;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface SignupRequest {
  name: string;
  email: string;
  password: string;
}

class ApiService {
  private api: AxiosInstance;
  private baseURL: string;

  constructor() {
    this.baseURL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

    this.api = axios.create({
      baseURL: this.baseURL,
      timeout: 10000,
      withCredentials: true,
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
    });

    this.api.interceptors.request.use(
      (config) => {
        const token = this.getTokenFromStorage();
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    this.api.interceptors.response.use(
      (response) => response,
      async (error) => {
        const originalRequest = error.config;

        if (error.response?.status === 401 && !originalRequest._retry) {
          originalRequest._retry = true;

          try {
            await this.refreshToken();

            const token = this.getTokenFromStorage();
            if (token) {
              originalRequest.headers.Authorization = `Bearer ${token}`;
            }
            return this.api(originalRequest);
          } catch (refreshError) {
            this.clearAuthData();
            if (typeof window !== "undefined") {
              window.location.href = "/login";
            }
            return Promise.reject(refreshError);
          }
        }

        return Promise.reject(error);
      }
    );
  }

  private getTokenFromStorage(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem("access_token");
  }

  private getRefreshTokenFromStorage(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem("refresh_token");
  }

  private setTokenInStorage(token: string): void {
    if (typeof window === "undefined") return;
    localStorage.setItem("access_token", token);
  }

  private setRefreshTokenInStorage(refreshToken: string): void {
    if (typeof window === "undefined") return;
    localStorage.setItem("refresh_token", refreshToken);
  }

  private clearAuthData(): void {
    if (typeof window === "undefined") return;
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
  }

  async refreshToken(): Promise<AuthResponse> {
    try {
      const refreshToken = this.getRefreshTokenFromStorage();

      if (!refreshToken) {
        throw new Error("No refresh token available");
      }

      const response: AxiosResponse<AuthResponse> = await this.api.post(
        "/auth/refresh",
        {},
        {
          headers: {
            Authorization: `Bearer refresh ${refreshToken}`,
          },
        }
      );

      if (response.data.access_token) {
        this.setTokenInStorage(response.data.access_token);
      }

      if (response.data.refresh_token) {
        this.setRefreshTokenInStorage(response.data.refresh_token);
      }

      return response.data;
    } catch (error: unknown) {
      throw new Error(
        error instanceof Error ? error.message : "Token refresh failed"
      );
    }
  }

  async login(credentials: LoginRequest): Promise<AuthResponse> {
    try {
      const response: AxiosResponse<AuthResponse> = await this.api.post(
        "/api/auth/login",
        credentials
      );

      if (response.data.access_token) {
        this.setTokenInStorage(response.data.access_token);
      }

      if (response.data.refresh_token) {
        this.setRefreshTokenInStorage(response.data.refresh_token);
      }

      return response.data;
    } catch (error: unknown) {
      throw new Error(error instanceof Error ? error.message : "Login failed");
    }
  }

  async signup(userData: SignupRequest): Promise<AuthResponse> {
    try {
      const response: AxiosResponse<AuthResponse> = await this.api.post(
        "/api/auth/signup",
        userData
      );

      if (response.data.access_token) {
        this.setTokenInStorage(response.data.access_token);
      }

      if (response.data.refresh_token) {
        this.setRefreshTokenInStorage(response.data.refresh_token);
      }

      return response.data;
    } catch (error: unknown) {
      throw new Error(error instanceof Error ? error.message : "Signup failed");
    }
  }

  async logout(): Promise<void> {
    try {
      await this.api.post("/auth/logout");
    } catch (error) {
      notify(`Error occured while logging out: ${error}`, "error");
    } finally {
      this.clearAuthData();
    }
  }
}

const apiService = new ApiService();
export default apiService;
