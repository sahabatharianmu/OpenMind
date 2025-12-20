import api from "@/api/client";

export interface Notification {
  id: string;
  user_id: string;
  type: 'handoff_request' | 'handoff_approved' | 'handoff_rejected' | 'handoff_cancelled' | 'team_invitation';
  title: string;
  message: string;
  related_entity_type?: 'patient_handoff' | 'team_invitation';
  related_entity_id?: string;
  is_read: boolean;
  read_at?: string;
  created_at: string;
  updated_at: string;
}

export interface NotificationListResponse {
  notifications: Notification[];
  total: number;
  unread_count: number;
}

const notificationService = {
  getNotifications: async (limit: number = 50, offset: number = 0) => {
    const response = await api.get<{ data: NotificationListResponse }>(
      `/notifications?limit=${limit}&offset=${offset}`
    );
    return response.data.data;
  },

  markAsRead: async (notificationId: string) => {
    const response = await api.put<{ data: null }>(`/notifications/${notificationId}/read`, {});
    return response.data.data;
  },

  markAllAsRead: async () => {
    const response = await api.put<{ data: null }>("/notifications/read-all", {});
    return response.data.data;
  },

  getUnreadCount: async () => {
    const response = await api.get<{ data: NotificationListResponse }>("/notifications?limit=1&offset=0");
    return response.data.data.unread_count;
  },
};

export default notificationService;

