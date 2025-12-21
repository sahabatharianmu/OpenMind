import { useState, useEffect } from "react";
import DashboardLayout from "@/components/dashboard/DashboardLayout";
import StatsCards from "@/components/dashboard/StatsCards";
import PatientList from "@/components/dashboard/PatientList";
import UpcomingAppointments from "@/components/dashboard/UpcomingAppointments";
import { useAuth } from "@/contexts/AuthContext";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { X, Sparkles, Users, UserCheck } from "lucide-react";
import { useNavigate } from "react-router-dom";

const Dashboard = () => {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [showUpgradeBanner, setShowUpgradeBanner] = useState(false);
  const [tier, setTier] = useState<string>("free");

  useEffect(() => {
    // Check if user just upgraded
    const upgradeSuccess = localStorage.getItem("upgrade_success");
    if (upgradeSuccess === "true") {
      setShowUpgradeBanner(true);
      localStorage.removeItem("upgrade_success");
    }

    // Check current tier
    const checkTier = async () => {
      try {
        const currentTier = await import("@/services/subscriptionService").then(
          (m) => m.subscriptionService.getSubscriptionTier()
        );
        setTier(currentTier);
        // If on paid tier and banner was shown, keep it for first visit
        if (currentTier === "paid" && upgradeSuccess === "true") {
          setShowUpgradeBanner(true);
        }
      } catch (error) {
        console.error("Failed to check tier", error);
      }
    };
    checkTier();
  }, []);

  const getGreeting = () => {
    const hour = new Date().getHours();
    if (hour < 12) return "Good morning";
    if (hour < 18) return "Good afternoon";
    return "Good evening";
  };

  const firstName = user?.full_name?.split(" ")[0] || "there";

  return (
    <DashboardLayout>
      <div className="p-4 sm:p-6 lg:p-8">
        {/* Upgrade Success Banner */}
        {showUpgradeBanner && tier === "paid" && (
          <Card className="mb-6 border-primary/20 bg-gradient-to-r from-primary/10 to-primary/5">
            <CardContent className="p-4">
              <div className="flex items-start gap-3">
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-2">
                    <Sparkles className="w-5 h-5 text-primary" />
                    <h3 className="font-semibold text-lg">🎉 Welcome to Paid Plan!</h3>
                  </div>
                  <p className="text-sm text-muted-foreground mb-3">
                    You now have unlimited patients and team members. Get started by inviting team members or adding more patients.
                  </p>
                  <div className="flex flex-wrap gap-4 text-sm mb-3">
                    <div className="flex items-center gap-2">
                      <Users className="w-4 h-4 text-primary" />
                      <span>Unlimited patients</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <UserCheck className="w-4 h-4 text-primary" />
                      <span>Unlimited team members</span>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => navigate("/dashboard/teams")}
                    >
                      Invite Team Members
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => navigate("/dashboard/patients")}
                    >
                      Add Patients
                    </Button>
                  </div>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  onClick={() => setShowUpgradeBanner(false)}
                >
                  <X className="w-4 h-4" />
                </Button>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Header */}
        <div className="mb-4 sm:mb-8">
          <h1 className="text-xl sm:text-2xl lg:text-3xl font-bold">
            {getGreeting()}, {firstName}
          </h1>
          <p className="text-muted-foreground mt-1 text-sm sm:text-base">
            Here's what's happening in your practice today.
          </p>
        </div>

        {/* Stats */}
        <StatsCards />

        {/* Main Content Grid */}
        <div className="grid grid-cols-1 xl:grid-cols-2 gap-6 mt-6">
          <UpcomingAppointments />
          <PatientList />
        </div>
      </div>
    </DashboardLayout>
  );
};

export default Dashboard;
