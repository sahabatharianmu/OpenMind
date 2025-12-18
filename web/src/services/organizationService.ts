import api from "@/api/client";

export interface Organization {
  id: string;
  name: string;
  type: string;
  tax_id?: string;
  npi?: string;
  address?: string;
  currency: string;
  locale: string;
  member_count: number;
  created_at: string;
}

export interface UpdateOrganizationRequest {
  name: string;
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
};
