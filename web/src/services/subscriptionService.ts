import api from "@/api/client";
import { organizationService } from "./organizationService";

export interface UpgradePrompt {
  feature: string; // "patients" or "clinicians"
  current: number;
  limit: number;
  upgrade_url: string;
}

export interface UsageStats {
  patient_count: number;
  clinician_count: number;
  patient_limit: number;
  clinician_limit: number;
}

export interface SubscriptionTier {
  tier: string; // "free" | "paid"
  limits: {
    patients: number; // -1 means unlimited
    clinicians: number; // -1 means unlimited
  };
}

export const subscriptionService = {
  /**
   * Get current subscription tier
   */
  getSubscriptionTier: async (): Promise<string> => {
    const org = await organizationService.getMyOrganization();
    return org.subscription_tier || "free";
  },

  /**
   * Get usage statistics
   */
  getUsageStats: async (): Promise<UsageStats | null> => {
    const org = await organizationService.getMyOrganization();
    return org.usage_stats || null;
  },

  /**
   * Get subscription tier details with limits
   */
  getTierDetails: async (): Promise<SubscriptionTier> => {
    const org = await organizationService.getMyOrganization();
    const tier = org.subscription_tier || "free";
    const usageStats = org.usage_stats;

    return {
      tier,
      limits: {
        patients: usageStats?.patient_limit ?? (tier === "free" ? 10 : -1),
        clinicians: usageStats?.clinician_limit ?? (tier === "free" ? 1 : -1),
      },
    };
  },

  /**
   * Check if a feature is at or near limit
   */
  isAtLimit: (current: number, limit: number): boolean => {
    if (limit === -1) return false; // Unlimited
    return current >= limit;
  },

  /**
   * Check if a feature is near limit (80% or more)
   */
  isNearLimit: (current: number, limit: number): boolean => {
    if (limit === -1) return false; // Unlimited
    return current >= limit * 0.8;
  },
};

