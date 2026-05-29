import type { AxiosRequestConfig, AxiosError } from 'axios'

// Extend AxiosRequestConfig to include the _retry property
declare module 'axios' {
  export interface AxiosRequestConfig {
    _retry?: boolean
  }
}

// Utility type for typed axios errors
export type TypedAxiosError<T = any> = AxiosError<T>
