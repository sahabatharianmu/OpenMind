import api from "@/api/client";
import { SubscriptionPlan } from "./adminPlanService"; // Reuse type

export const publicPlanService = {
  listActivePlans: async () => {
    const response = await api.get<{ data: SubscriptionPlan[] }>("/plans");
    return response.data.data;
  }
};
