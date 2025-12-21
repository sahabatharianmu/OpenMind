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

const Pricing = () => {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [showUpgradeModal, setShowUpgradeModal] = useState(false);
  const [usageStats, setUsageStats] = useState<UsageStats | null>(null);
  const [tier, setTier] = useState<string>("free");
  const [loading, setLoading] = useState(true);

  const planPrice = 29; // Monthly price in USD

  useEffect(() => {
    const loadData = async () => {
      try {
        const [stats, currentTier] = await Promise.all([
          subscriptionService.getUsageStats(),
          subscriptionService.getSubscriptionTier(),
        ]);
        setUsageStats(stats);
        setTier(currentTier);
      } catch (error) {
        console.error("Failed to load usage stats", error);
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

  const handleUpgrade = () => {
    // Check if user can upgrade (owner/admin only)
    const canManagePaymentMethods = user?.role === "admin" || user?.role === "owner";
    if (!canManagePaymentMethods) {
      return;
    }
    setShowUpgradeModal(true);
  };

  const handleUpgradeSuccess = () => {
    // Refresh user data or redirect
    window.location.reload(); // Simple refresh for now
  };

  const isFreeTier = tier === "free";
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

        {/* Upgrade Benefits Banner */}
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

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
          {/* Free Plan */}
          <Card className={isFreeTier ? "border-primary" : ""}>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>Free</CardTitle>
                <Badge variant={isFreeTier ? "default" : "secondary"}>
                  {isFreeTier ? "Current Plan" : "Previous"}
                </Badge>
              </div>
              <CardDescription>
                Perfect for getting started
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="mb-6">
                <span className="text-4xl font-bold">$0</span>
                <span className="text-muted-foreground">/month</span>
              </div>
              <ul className="space-y-3 mb-6">
                {features.map((feature, index) => (
                  <li key={index} className="flex items-start justify-between">
                    <span className="text-sm">{feature.name}</span>
                    <div className="flex items-center gap-2">
                      {feature.free === "Yes" || feature.free === "Unlimited" ? (
                        <Check className="h-4 w-4 text-primary" />
                      ) : (
                        <X className="h-4 w-4 text-muted-foreground" />
                      )}
                      <span className="text-sm text-muted-foreground">{feature.free}</span>
                    </div>
                  </li>
                ))}
              </ul>
              <Button variant="outline" className="w-full" disabled>
                Current Plan
              </Button>
            </CardContent>
          </Card>

          {/* Paid Plan */}
          <Card className={!isFreeTier ? "border-primary shadow-lg" : "border-primary/50"}>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>Paid</CardTitle>
                <Badge variant={!isFreeTier ? "default" : "default"}>
                  {!isFreeTier ? "Current Plan" : "Recommended"}
                </Badge>
              </div>
              <CardDescription>
                For growing practices
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="mb-6">
                <span className="text-4xl font-bold">$29</span>
                <span className="text-muted-foreground">/month</span>
              </div>
              <ul className="space-y-3 mb-6">
                {features.map((feature, index) => (
                  <li key={index} className="flex items-start justify-between">
                    <span className="text-sm">{feature.name}</span>
                    <div className="flex items-center gap-2">
                      <Check className="h-4 w-4 text-primary" />
                      <span className="text-sm text-muted-foreground">{feature.paid}</span>
                    </div>
                  </li>
                ))}
              </ul>
              <Button 
                onClick={handleUpgrade} 
                className="w-full"
                disabled={!isFreeTier}
              >
                {isFreeTier ? "Upgrade Now" : "Current Plan"}
              </Button>
            </CardContent>
          </Card>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Need Help Choosing?</CardTitle>
            <CardDescription>
              Contact our team to discuss which plan is right for you.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button variant="outline" onClick={() => navigate("/dashboard")}>
              Back to Dashboard
            </Button>
          </CardContent>
        </Card>

        <UpgradeModal
          open={showUpgradeModal}
          onOpenChange={setShowUpgradeModal}
          onSuccess={handleUpgradeSuccess}
          planPrice={planPrice}
        />
      </div>
    </DashboardLayout>
  );
};

export default Pricing;

