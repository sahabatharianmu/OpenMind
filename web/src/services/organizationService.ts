import api from "@/api/client";

export interface UsageStats {
  patient_count: number;
  clinician_count: number;
  patient_limit: number; // -1 means unlimited
  clinician_limit: number; // -1 means unlimited
}

export interface Organization {
  id: string;
  name: string;
  type: string;
  subscription_tier: string;
  tax_id?: string;
  npi?: string;
  address?: string;
  currency: string;
  locale: string;
  member_count: number;
  usage_stats?: UsageStats;
  created_at: string;
}

export interface UpdateOrganizationRequest {
  name: string;
  type?: string;
  tax_id?: string;
  npi?: string;
  address?: string;
  currency?: string;
  locale?: string;
}

export const organizationService = {
  getMyOrganization: async () => {
    const response = await api.get<{ data: Organization }>("/organizations/me");
    return response.data.data;
  },

  updateOrganization: async (data: UpdateOrganizationRequest) => {
    const response = await api.put<{ data: Organization }>("/organizations/me", data);
    return response.data.data;
  },

  listTeamMembers: async () => {
    const response = await api.get<{ data: TeamMember[] }>("/organizations/me/members");
    return response.data.data;
  },

  updateMemberRole: async (userId: string, role: string) => {
    const response = await api.put<{ data: null }>(`/organizations/me/members/${userId}/role`, { role });
    return response.data.data;
  },

  removeMember: async (userId: string) => {
    const response = await api.delete<{ data: null }>(`/organizations/me/members/${userId}`);
    return response.data.data;
  },
};

export interface TeamMember {
  user_id: string;
  email: string;
  full_name: string;
  role: string;
  joined_at: string;
}
