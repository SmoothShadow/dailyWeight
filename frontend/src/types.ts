export interface User {
  id: number
  username: string
}

export interface WeightRecord {
  id: number
  userId: number
  username: string
  date: string
  weight: number
  createdAt: string
  updatedAt: string
}

export interface AuthResponse {
  user: User
}

export interface WeightsResponse {
  records: WeightRecord[]
}
