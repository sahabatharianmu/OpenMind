import api from "@/api/client";
import type { PaginatedResponse } from "@/types/api";

export interface AuditLog {
  id: string;
  organization_id: string;
  user_id: string;
  user_name: string;
  action: string;
  resource_type: string;
  resource_id?: string;
  details: any;
  ip_address?: string;
  created_at: string;
}

export interface AuditLogFilters {
  resource_type?: string;
  user_id?: string;
  action?: string;
  start_date?: string;
  end_date?: string;
}

export const auditLogService = {
  list: async (page: number = 1, pageSize: number = 20, filters?: AuditLogFilters) => {
    const params: any = {
      page,
      page_size: pageSize,
    };

    if (filters?.resource_type) params.resource_type = filters.resource_type;
    if (filters?.user_id) params.user_id = filters.user_id;
    if (filters?.action) params.action = filters.action;
    if (filters?.start_date) params.start_date = filters.start_date;
    if (filters?.end_date) params.end_date = filters.end_date;

    const response = await api.get<{ data: PaginatedResponse<AuditLog> }>("/audit-logs", { params });
    return response.data;
  },
};
