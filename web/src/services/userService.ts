import api from "@/api/client";

export interface UpdateProfileRequest {
  full_name: string;
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

export interface UserProfile {
  id: string;
  email: string;
  full_name: string;
  role: string;
}

export const userService = {
  getProfile: async () => {
    const response = await api.get<{ data: UserProfile }>("/users/me");
    return response.data.data;
  },

  updateProfile: async (data: UpdateProfileRequest) => {
    const response = await api.put<{ data: UserProfile }>("/users/me", data);
    return response.data.data;
  },

  changePassword: async (data: ChangePasswordRequest) => {
    const response = await api.put("/auth/password", data);
    return response.data;
  },
};
