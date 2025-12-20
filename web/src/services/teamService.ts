import api from "@/api/client";
import type { PaginatedResponse } from "@/types/api";

export interface TeamInvitation {
  id: string;
  organization_id: string;
  email: string;
  role: string;
  status: "pending" | "accepted" | "expired" | "cancelled";
  expires_at: string;
  accepted_at?: string;
  created_at: string;
}

export interface SendInvitationRequest {
  email: string;
  role: string;
}

export interface AcceptInvitationRequest {
  token: string;
}

export interface RegisterWithInvitationRequest {
  token: string;
  email: string;
  password: string;
  full_name: string;
}

export interface ListInvitationsResponse {
  invitations: TeamInvitation[];
  total: number;
  page: number;
  page_size: number;
}

export const teamService = {
  sendInvitation: async (data: SendInvitationRequest) => {
    const response = await api.post<{ data: TeamInvitation }>("/team/invitations", data);
    return response.data.data;
  },

  listInvitations: async (page: number = 1, pageSize: number = 20) => {
    const response = await api.get<{ data: ListInvitationsResponse }>("/team/invitations", {
      params: { page, page_size: pageSize },
    });
    return response.data.data;
  },

  getInvitationByToken: async (token: string) => {
    const response = await api.get<{ data: TeamInvitation }>(`/team/invitations/${token}`);
    return response.data.data;
  },

  acceptInvitation: async (data: AcceptInvitationRequest) => {
    const response = await api.post<{ data: null }>("/team/invitations/accept", data);
    return response.data.data;
  },

  registerAndAcceptInvitation: async (data: RegisterWithInvitationRequest) => {
    const response = await api.post<{ data: { user_id: string } }>("/team/invitations/register", data);
    return response.data.data;
  },

  cancelInvitation: async (invitationId: string) => {
    const response = await api.delete(`/team/invitations/${invitationId}`);
    return response.data;
  },

  resendInvitation: async (invitationId: string) => {
    const response = await api.post(`/team/invitations/${invitationId}/resend`);
    return response.data;
  },
};

