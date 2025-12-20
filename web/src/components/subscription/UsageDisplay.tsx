import { useEffect, useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Badge } from "@/components/ui/badge";
import { Users, UserCheck, AlertTriangle } from "lucide-react";
import { subscriptionService, UsageStats } from "@/services/subscriptionService";
import { Button } from "@/components/ui/button";
import { useNavigate } from "react-router-dom";

interface UsageDisplayProps {
  showUpgradeButton?: boolean;
  compact?: boolean;
}

const UsageDisplay = ({ showUpgradeButton = false, compact = false }: UsageDisplayProps) => {
  const [usageStats, setUsageStats] = useState<UsageStats | null>(null);
  const [tier, setTier] = useState<string>("free");
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    const loadUsageStats = async () => {
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

    loadUsageStats();
  }, []);

  if (loading || !usageStats) {
    return (
      <Card>
        <CardContent className="p-4">
          <div className="animate-pulse">Loading usage statistics...</div>
        </CardContent>
      </Card>
    );
  }

  const isUnlimited = (limit: number) => limit === -1;
  const getProgress = (current: number, limit: number) => {
    if (isUnlimited(limit)) return 0;
    return Math.min((current / limit) * 100, 100);
  };

  const isAtLimit = (current: number, limit: number) => {
    if (isUnlimited(limit)) return false;
    return current >= limit;
  };

  const isNearLimit = (current: number, limit: number) => {
    if (isUnlimited(limit)) return false;
    return current >= limit * 0.8;
  };

  const patientProgress = getProgress(usageStats.patient_count, usageStats.patient_limit);
  const clinicianProgress = getProgress(usageStats.clinician_count, usageStats.clinician_limit);

  const patientAtLimit = isAtLimit(usageStats.patient_count, usageStats.patient_limit);
  const patientNearLimit = isNearLimit(usageStats.patient_count, usageStats.patient_limit);
  const clinicianAtLimit = isAtLimit(usageStats.clinician_count, usageStats.clinician_limit);
  const clinicianNearLimit = isNearLimit(usageStats.clinician_count, usageStats.clinician_limit);

  if (compact) {
    return (
      <div className="space-y-2">
        <div className="flex items-center justify-between text-sm">
          <span className="flex items-center gap-2">
            <Users className="h-4 w-4" />
            Patients
          </span>
          <span className={patientAtLimit ? "text-destructive font-semibold" : patientNearLimit ? "text-yellow-600" : ""}>
            {usageStats.patient_count} / {isUnlimited(usageStats.patient_limit) ? "∞" : usageStats.patient_limit}
          </span>
        </div>
        <Progress value={patientProgress} className={patientAtLimit ? "h-2" : "h-1"} />
        <div className="flex items-center justify-between text-sm">
          <span className="flex items-center gap-2">
            <UserCheck className="h-4 w-4" />
            Clinicians
          </span>
          <span className={clinicianAtLimit ? "text-destructive font-semibold" : clinicianNearLimit ? "text-yellow-600" : ""}>
            {usageStats.clinician_count} / {isUnlimited(usageStats.clinician_limit) ? "∞" : usageStats.clinician_limit}
          </span>
        </div>
        <Progress value={clinicianProgress} className={clinicianAtLimit ? "h-2" : "h-1"} />
      </div>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Usage Statistics</CardTitle>
            <CardDescription>
              Current usage for your {tier === "free" ? "Free" : "Paid"} plan
            </CardDescription>
          </div>
          <Badge variant={tier === "paid" ? "default" : "secondary"}>{tier.toUpperCase()}</Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Patients Usage */}
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Users className="h-5 w-5 text-muted-foreground" />
              <span className="font-medium">Patients</span>
            </div>
            <div className="flex items-center gap-2">
              {patientAtLimit && <AlertTriangle className="h-4 w-4 text-destructive" />}
              <span className={patientAtLimit ? "text-destructive font-semibold" : patientNearLimit ? "text-yellow-600" : ""}>
                {usageStats.patient_count} / {isUnlimited(usageStats.patient_limit) ? "Unlimited" : usageStats.patient_limit}
              </span>
            </div>
          </div>
          <Progress
            value={patientProgress}
            className={patientAtLimit ? "h-3" : patientNearLimit ? "h-2" : "h-2"}
          />
          {patientAtLimit && (
            <p className="text-sm text-destructive">
              Patient limit reached. Upgrade to add more patients.
            </p>
          )}
          {patientNearLimit && !patientAtLimit && (
            <p className="text-sm text-yellow-600">
              You're approaching your patient limit.
            </p>
          )}
        </div>

        {/* Clinicians Usage */}
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <UserCheck className="h-5 w-5 text-muted-foreground" />
              <span className="font-medium">Clinicians</span>
            </div>
            <div className="flex items-center gap-2">
              {clinicianAtLimit && <AlertTriangle className="h-4 w-4 text-destructive" />}
              <span className={clinicianAtLimit ? "text-destructive font-semibold" : clinicianNearLimit ? "text-yellow-600" : ""}>
                {usageStats.clinician_count} / {isUnlimited(usageStats.clinician_limit) ? "Unlimited" : usageStats.clinician_limit}
              </span>
            </div>
          </div>
          <Progress
            value={clinicianProgress}
            className={clinicianAtLimit ? "h-3" : clinicianNearLimit ? "h-2" : "h-2"}
          />
          {clinicianAtLimit && (
            <p className="text-sm text-destructive">
              Clinician limit reached. Upgrade to add more team members.
            </p>
          )}
          {clinicianNearLimit && !clinicianAtLimit && (
            <p className="text-sm text-yellow-600">
              You're approaching your clinician limit.
            </p>
          )}
        </div>

        {showUpgradeButton && tier === "free" && (
          <Button
            onClick={() => navigate("/pricing")}
            className="w-full"
            variant={patientAtLimit || clinicianAtLimit ? "default" : "outline"}
          >
            Upgrade to Paid Plan
          </Button>
        )}
      </CardContent>
    </Card>
  );
};

export default UsageDisplay;

