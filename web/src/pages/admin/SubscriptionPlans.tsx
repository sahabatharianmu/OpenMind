import { useState, useEffect, useCallback } from "react";
import AdminLayout from "@/layouts/AdminLayout";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { adminPlanService, SubscriptionPlan } from "@/services/adminPlanService";
import { CreatePlanDialog } from "@/components/subscription/CreatePlanDialog";
import { useTranslation } from "react-i18next";

export default function SubscriptionPlans() {
  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
  const [loading, setLoading] = useState(true);
  const { t } = useTranslation();
  
  const fetchPlans = useCallback(async () => {
    try {
      setLoading(true);
      const data = await adminPlanService.listPlans();
      setPlans(data);
    } catch (error) {
      console.error("Failed to list plans", error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchPlans();
  }, [fetchPlans]);

  return (
    <AdminLayout>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">{t('common.plans')}</h2>
          <p className="text-muted-foreground">{t('common.manage_pricing')}</p>
        </div>
        <CreatePlanDialog onPlanCreated={fetchPlans} />
      </div>

      {loading ? (
        <div>Loading plans...</div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {plans.map((plan) => (
            <Card key={plan.id}>
              <CardHeader>
                <CardTitle>{plan.name}</CardTitle>
                <CardDescription>{plan.is_active ? t('common.active') : t('common.inactive')}</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-2 mb-4">
                  {plan.prices && plan.prices.length > 0 ? (
                      plan.prices.map((p, idx) => (
                        <div key={idx} className="text-2xl font-bold">
                          {(p.price / 100).toLocaleString(undefined, { style: 'currency', currency: p.currency })}
                          <span className="text-sm font-normal text-muted-foreground">/mo</span>
                        </div>
                      ))
                  ) : (
                      <div className="text-muted-foreground text-sm">No pricing configured</div>
                  )}
                </div>
                <ul className="list-disc list-inside text-sm space-y-1 mb-4">
                   {/* Parse limits if possible, or show generic info */}
                  <li>{t('common.patients')}: {plan.limits?.patient_limit === -1 ? t('common.unlimited') : plan.limits?.patient_limit}</li>
                  <li>{t('common.clinicians')}: {plan.limits?.clinician_limit === -1 ? t('common.unlimited') : plan.limits?.clinician_limit}</li>
                </ul>
                <Button variant="outline" className="w-full">{t('common.edit_plan')}</Button>
              </CardContent>
            </Card>
          ))}
          
          {plans.length === 0 && (
            <div className="col-span-full text-center p-8 border rounded-lg border-dashed">
              <p className="text-muted-foreground">{t('common.no_plans')}</p>
            </div>
          )}
        </div>
      )}
    </AdminLayout>
  );
}
