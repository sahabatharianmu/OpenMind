import { useState, useEffect } from "react";
import { Check, X, Sparkles, Users, UserCheck } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import DashboardLayout from "@/components/dashboard/DashboardLayout";
import { useAuth } from "@/contexts/AuthContext";
import { useNavigate } from "react-router-dom";
import UpgradeModal from "@/components/payment/UpgradeModal";
import { subscriptionService, UsageStats } from "@/services/subscriptionService";
import { publicPlanService } from "@/services/publicPlanService";
import { SubscriptionPlan } from "@/services/adminPlanService";

const Pricing = () => {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [showUpgradeModal, setShowUpgradeModal] = useState(false);
  const [usageStats, setUsageStats] = useState<UsageStats | null>(null);
  const [tier, setTier] = useState<string>("free");
  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedPlan, setSelectedPlan] = useState<SubscriptionPlan | null>(null);

  // const planPrice = 29; // Monthly price in USD - Now dynamic

  useEffect(() => {
    const loadData = async () => {
      try {
        const [stats, currentTier, activePlans] = await Promise.all([
          subscriptionService.getUsageStats(),
          subscriptionService.getSubscriptionTier(),
          publicPlanService.listActivePlans(),
        ]);
        setUsageStats(stats);
        setTier(currentTier);
        setPlans(activePlans);
      } catch (error) {
        console.error("Failed to load pricing data", error);
      } finally {
        setLoading(false);
      }
    };
    loadData();
  }, []);

  const features = [
    { name: "Patients", free: "Up to 10", paid: "Unlimited" },
    { name: "Team Members", free: "1 (Owner only)", paid: "Unlimited" },
    { name: "Clinical Notes", free: "Unlimited", paid: "Unlimited" },
    { name: "Appointments", free: "Unlimited", paid: "Unlimited" },
    { name: "Patient Assignments", free: "Yes", paid: "Yes" },
    { name: "Patient Handoffs", free: "Yes", paid: "Yes" },
    { name: "Export/Import", free: "Yes", paid: "Yes" },
    { name: "Audit Logs", free: "Yes", paid: "Yes" },
    { name: "HIPAA Compliance", free: "Yes", paid: "Yes" },
    { name: "Email Support", free: "Community", paid: "Priority" },
  ];

  const handleUpgrade = (plan: SubscriptionPlan) => {
    // Check if user can upgrade (owner/admin only)
    const canManagePaymentMethods = user?.role === "admin" || user?.role === "owner";
    if (!canManagePaymentMethods) {
      return;
    }
    setSelectedPlan(plan);
    setShowUpgradeModal(true);
  };

  const handleUpgradeSuccess = () => {
    // Refresh user data or redirect
    window.location.reload(); // Simple refresh for now
  };

  const isFreeTier = tier === "free";
  
  // Helper to get display price (default to USD or first available)
  const getDisplayPrice = (plan: SubscriptionPlan) => {
      if (!plan.prices || plan.prices.length === 0) {
          console.warn("Plan has no prices:", plan);
          return { price: 0, currency: 'USD' };
      }
      // TODO: improvements to match user locale
      const usdPrice = plan.prices.find(p => p.currency === 'USD');
      const price = usdPrice || plan.prices[0];
      
      const result = { 
          price: price?.price ?? 0, 
          currency: price?.currency || 'USD' 
      };
      // Debug log
      if (!result.currency) console.error("Computed invalid currency for plan:", plan, result);
      return result;
  }

  // Helper to get plan features from limits (simplified)
  const getPlanFeatures = (plan: SubscriptionPlan) => {
      const displayPrice = getDisplayPrice(plan);
      return [
          { name: "Patients", value: plan.limits?.patient_limit === -1 ? "Unlimited" : plan.limits?.patient_limit },
          { name: "Team Members", value: plan.limits?.clinician_limit === -1 ? "Unlimited" : plan.limits?.clinician_limit },
          { name: "Clinical Notes", value: "Unlimited" }, // Placeholder
          { name: "Support", value: displayPrice.price > 0 ? "Priority" : "Community" },
      ];
  };

  // ... (Keep Usage Indicators Logic if needed) ...
  const patientUsage = usageStats ? `${usageStats.patient_count}/${usageStats.patient_limit === -1 ? "∞" : usageStats.patient_limit}` : "0/10";
  const clinicianUsage = usageStats ? `${usageStats.clinician_count}/${usageStats.clinician_limit === -1 ? "∞" : usageStats.clinician_limit}` : "0/1";
  const patientProgress = usageStats && usageStats.patient_limit !== -1 
    ? Math.min((usageStats.patient_count / usageStats.patient_limit) * 100, 100)
    : 0;
  const clinicianProgress = usageStats && usageStats.clinician_limit !== -1
    ? Math.min((usageStats.clinician_count / usageStats.clinician_limit) * 100, 100)
    : 0;


  return (
    <DashboardLayout>
      <div className="p-6 lg:p-8 max-w-6xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold mb-2">Choose Your Plan</h1>
          <p className="text-muted-foreground">
            Compare features and choose the plan that's right for your practice.
          </p>
        </div>

        {/* Upgrade Benefits Banner, Usage Indicators - Keep existing structure but maybe hide if loading */}
        {isFreeTier && (
          <Card className="mb-6 border-primary/20 bg-primary/5">
            <CardContent className="p-4">
              <div className="flex items-start gap-3">
                <Sparkles className="w-5 h-5 text-primary mt-0.5" />
                <div className="flex-1">
                  <h3 className="font-semibold mb-1">Unlock Unlimited Growth</h3>
                  <p className="text-sm text-muted-foreground mb-3">
                    Upgrade to the paid plan to remove limits and scale your practice.
                  </p>
                  <div className="flex flex-wrap gap-4 text-sm">
                    <div className="flex items-center gap-2">
                       <Users className="w-4 h-4 text-primary" />
                      <span>Unlimited patients</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <UserCheck className="w-4 h-4 text-primary" />
                      <span>Unlimited team members</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Sparkles className="w-4 h-4 text-primary" />
                      <span>Priority support</span>
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Usage Indicators */}
        {isFreeTier && usageStats && (
          <Card className="mb-6">
            <CardHeader>
              <CardTitle className="text-lg">Current Usage</CardTitle>
              <CardDescription>Your usage on the free tier</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <div className="flex items-center justify-between text-sm">
                  <span className="flex items-center gap-2">
                    <Users className="w-4 h-4" />
                    Patients
                  </span>
                  <span className="font-medium">{patientUsage}</span>
                </div>
                <Progress value={patientProgress} />
              </div>
              <div className="space-y-2">
                <div className="flex items-center justify-between text-sm">
                  <span className="flex items-center gap-2">
                    <UserCheck className="w-4 h-4" />
                    Team Members
                  </span>
                  <span className="font-medium">{clinicianUsage}</span>
                </div>
                <Progress value={clinicianProgress} />
              </div>
            </CardContent>
          </Card>
        )}

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-12">
          {plans.map((plan) => {
             const isCurrentPlan = false; // TODO: Match with current subscription ID from org
             const displayPrice = getDisplayPrice(plan);
             const isPlanFree = displayPrice.price === 0;

             // Aesthetic decision: Highlight the first paid plan or a specific "Pro" plan
             const isHighlighted = !isPlanFree; 

             return (
              <div 
                key={plan.id} 
                className={`
                  relative rounded-2xl p-8 transition-all duration-300
                  ${isHighlighted 
                    ? "bg-white border-2 border-primary/20 shadow-xl shadow-primary/5 scale-105 z-10" 
                    : "bg-white/50 border border-border/50 hover:border-primary/20 hover:shadow-lg hover:-translate-y-1"
                  }
                `}
              >
                {isHighlighted && (
                  <div className="absolute -top-4 left-1/2 -translate-x-1/2">
                    <span className="bg-primary text-primary-foreground text-xs font-bold px-3 py-1 rounded-full uppercase tracking-wider">
                      Most Popular
                    </span>
                  </div>
                )}

                <div className="mb-6">
                  <h3 className="font-heading text-lg font-bold text-foreground">{plan.name}</h3>
                  <p className="text-sm text-muted-foreground mt-1">
                    {isPlanFree ? "Perfect for getting started" : "For growing practices"}
                  </p>
                </div>

                <div className="mb-6 flex items-baseline gap-1">
                  <span className="text-4xl font-extrabold font-heading text-foreground">
                      {(() => {
                          if (!displayPrice.currency) {
                              console.error("Missing Currency for plan:", plan.id, plan);
                              return "N/A";
                          }
                          try {
                              return (displayPrice.price / 100).toLocaleString('en-US', { style: 'currency', currency: displayPrice.currency });
                          } catch (e) {
                              console.error("Error formatting price:", e, displayPrice);
                              return `$${displayPrice.price / 100}`;
                          }
                      })()}
                  </span>
                  <span className="text-sm font-medium text-muted-foreground">/month</span>
                </div>

                <Button 
                    onClick={() => handleUpgrade(plan)} 
                    className={`
                        w-full mb-8 font-semibold transition-all
                        ${isHighlighted 
                            ? "bg-primary hover:bg-primary/90 shadow-lg shadow-primary/20" 
                            : "bg-white border-2 border-primary/10 hover:border-primary hover:bg-primary/5 text-foreground"
                        }
                    `}
                    variant={isHighlighted ? "default" : "outline"}
                    disabled={isCurrentPlan}
                >
                    {isCurrentPlan ? "Current Plan" : (isPlanFree ? "Get Started" : "Upgrade Now")}
                </Button>

                <div className="space-y-4">
                    <p className="text-xs font-bold uppercase tracking-wider text-muted-foreground">Features</p>
                    <ul className="space-y-3">
                        {getPlanFeatures(plan).map((feature, index) => (
                        <li key={index} className="flex items-start gap-3 text-sm group">
                            <div className={`
                                mt-0.5 rounded-full p-0.5 
                                ${isHighlighted ? "bg-primary/10 text-primary" : "bg-gray-100 text-gray-500 group-hover:text-primary transition-colors"}
                            `}>
                                <Check className="w-3.5 h-3.5" />
                            </div>
                            <span className="text-foreground/80">
                                <span className="font-semibold text-foreground">{feature.value}</span> {feature.name}
                            </span>
                        </li>
                        ))}
                    </ul>
                </div>
              </div>
             );
          })}
        </div>

        <div className="bg-slate-50 border border-border/50 rounded-2xl p-8 text-center max-w-2xl mx-auto">
             <h3 className="font-heading text-lg font-bold mb-2">Need a Custom Enterprise Plan?</h3>
             <p className="text-muted-foreground mb-6">
               For large organizations requiring custom limits, dedicated support, or on-premise deployment.
             </p>
             <Button variant="outline" onClick={() => navigate("/dashboard")}>
               Contact Sales
             </Button>
        </div>
        <UpgradeModal
          open={showUpgradeModal}
          onOpenChange={setShowUpgradeModal}
          onSuccess={handleUpgradeSuccess}
          planPrice={selectedPlan ? getDisplayPrice(selectedPlan).price / 100 : 0}
        />
      </div>
    </DashboardLayout>
  );
};

export default Pricing;

