import { Check, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import DashboardLayout from "@/components/dashboard/DashboardLayout";
import { useAuth } from "@/contexts/AuthContext";
import { useNavigate } from "react-router-dom";

const Pricing = () => {
  const { user } = useAuth();
  const navigate = useNavigate();

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
    // TODO: Implement upgrade flow (payment integration)
    // For now, just show a message
    alert("Upgrade flow coming soon! Contact support to upgrade your plan.");
  };

  return (
    <DashboardLayout>
      <div className="p-6 lg:p-8 max-w-6xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold mb-2">Choose Your Plan</h1>
          <p className="text-muted-foreground">
            Compare features and choose the plan that's right for your practice.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
          {/* Free Plan */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>Free</CardTitle>
                <Badge variant="secondary">Current</Badge>
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
          <Card className="border-primary shadow-lg">
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>Paid</CardTitle>
                <Badge variant="default">Recommended</Badge>
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
              <Button onClick={handleUpgrade} className="w-full">
                Upgrade Now
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
      </div>
    </DashboardLayout>
  );
};

export default Pricing;

