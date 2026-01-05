import { useState, useEffect, useCallback } from "react";
import AdminLayout from "@/layouts/AdminLayout";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { adminPlanService, SubscriptionPlan } from "@/services/adminPlanService";
import { CreatePlanDialog } from "@/components/subscription/CreatePlanDialog";

export default function SubscriptionPlans() {
  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
  const [loading, setLoading] = useState(true);
  
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
          <h2 className="text-3xl font-bold tracking-tight">Subscription Plans</h2>
          <p className="text-muted-foreground">Manage your product pricing and packages.</p>
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
                <CardDescription>{plan.is_active ? "Active" : "Inactive"}</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold mb-4">
                  {(plan.price / 100).toLocaleString('en-US', { style: 'currency', currency: plan.currency })}
                  <span className="text-sm font-normal text-muted-foreground">/mo</span>
                </div>
                <ul className="list-disc list-inside text-sm space-y-1 mb-4">
                   {/* Parse limits if possible, or show generic info */}
                  <li>Patients: {plan.limits?.patient_limit === -1 ? "Unlimited" : plan.limits?.patient_limit}</li>
                  <li>Clinicians: {plan.limits?.clinician_limit === -1 ? "Unlimited" : plan.limits?.clinician_limit}</li>
                </ul>
                <Button variant="outline" className="w-full">Edit Plan</Button>
              </CardContent>
            </Card>
          ))}
          
          {plans.length === 0 && (
            <div className="col-span-full text-center p-8 border rounded-lg border-dashed">
              <p className="text-muted-foreground">No plans found. Create one to get started.</p>
            </div>
          )}
        </div>
      )}
    </AdminLayout>
  );
}
