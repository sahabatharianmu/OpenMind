import api from "@/api/client";

export interface AdminStats {
  total_tenants: number;
  total_users: number;
  active_subscriptions: number;
  monthly_revenue: number;
}

export interface TenantListItem {
  id: string;
  name: string;
  subscription_tier: string;
  member_count: number;
  created_at: string;
}

export interface TenantListResponse {
  tenants: TenantListItem[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface UserListItem {
  id: string;
  email: string;
  full_name: string;
  system_role: string;
  org_name: string | null;
  created_at: string;
}

export interface UserListResponse {
  users: UserListItem[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export const adminService = {
  async getStats(): Promise<AdminStats> {
    const res = await api.get<{ data: AdminStats }>("/admin/stats");
    return res.data.data;
  },

  async listTenants(page = 1, perPage = 20): Promise<TenantListResponse> {
    const res = await api.get<{ data: TenantListResponse }>("/admin/tenants", {
      params: { page, per_page: perPage },
    });
    return res.data.data;
  },

  async listUsers(page = 1, perPage = 20): Promise<UserListResponse> {
    const res = await api.get<{ data: UserListResponse }>("/admin/users", {
      params: { page, per_page: perPage },
    });
    return res.data.data;
  },

  async updateUserRole(userId: string, systemRole: string): Promise<void> {
    await api.patch(`/admin/users/${userId}`, { system_role: systemRole });
  },
};
