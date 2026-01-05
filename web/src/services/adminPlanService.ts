import api from "@/api/client";

export interface PlanPrice {
  id?: string;
  currency: string;
  price: number;
}

export interface SubscriptionPlan {
  id: string;
  name: string;
  prices: PlanPrice[];
  is_active: boolean;
  limits: any;
  created_at: string;
  updated_at: string;
}

export interface CreatePlanRequest {
  name: string;
  prices: PlanPrice[];
  is_active: boolean;
  limits?: any;
}

export const adminPlanService = {
  listPlans: async () => {
    const response = await api.get<{ data: SubscriptionPlan[] }>("/admin/plans");
    return response.data.data;
  },

  createPlan: async (data: CreatePlanRequest) => {
    const response = await api.post<{ data: SubscriptionPlan }>("/admin/plans", data);
    return response.data.data;
  },

  updatePlan: async (id: string, data: Partial<CreatePlanRequest>) => {
    const response = await api.put<{ data: SubscriptionPlan }>(`/admin/plans/${id}`, data);
    return response.data.data;
  },
  
  getPlan: async (id: string) => {
    const response = await api.get<{ data: SubscriptionPlan }>(`/admin/plans/${id}`);
    return response.data.data;
  }
};
