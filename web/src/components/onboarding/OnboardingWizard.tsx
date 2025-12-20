import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Progress } from "@/components/ui/progress";
import { useAuth } from "@/contexts/AuthContext";
import { useToast } from "@/hooks/use-toast";
import { organizationService } from "@/services/organizationService";
import { userService } from "@/services/userService";
import { CheckCircle2, ArrowRight, ArrowLeft, Sparkles } from "lucide-react";

interface OnboardingWizardProps {
  onComplete: () => void;
  onSkip: () => void;
}

const OnboardingWizard = ({ onComplete, onSkip }: OnboardingWizardProps) => {
  const { user } = useAuth();
  const { toast } = useToast();
  const [currentStep, setCurrentStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [loadingData, setLoadingData] = useState(true);

  // Step 1: Welcome (no form data)
  
  // Step 2: Organization Setup
  const [orgName, setOrgName] = useState("");
  const [orgType, setOrgType] = useState("clinic");

  // Step 3: Profile
  const [fullName, setFullName] = useState(user?.full_name || "");

  // Load organization data
  useEffect(() => {
    const loadData = async () => {
      try {
        const org = await organizationService.getMyOrganization();
        setOrgName(org.name || "");
        setOrgType(org.type || "clinic");
      } catch (error) {
        console.error("Failed to load organization:", error);
      } finally {
        setLoadingData(false);
      }
    };
    loadData();
  }, []);

  const totalSteps = 4;
  const progress = ((currentStep + 1) / totalSteps) * 100;

  if (loadingData) {
    return (
      <Card className="w-full max-w-2xl mx-auto">
        <CardContent className="p-12 text-center">
          <div className="animate-pulse">Loading...</div>
        </CardContent>
      </Card>
    );
  }

  const handleNext = async () => {
    if (currentStep === totalSteps - 1) {
      // Last step - complete onboarding
      await handleComplete();
      return;
    }
    setCurrentStep(currentStep + 1);
  };

  const handleBack = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  const handleSkip = () => {
    onSkip();
  };

  const handleComplete = async () => {
    setLoading(true);
    try {
      // Update organization if name or type changed
      const org = await organizationService.getMyOrganization();
      if (orgName && (orgName !== org.name || orgType !== org.type)) {
        await organizationService.updateOrganization({
          name: orgName,
          type: orgType,
        });
      }

      // Update profile if name changed
      if (fullName && fullName !== user?.full_name) {
        await userService.updateProfile({ full_name: fullName });
      }

      toast({
        title: "Setup Complete!",
        description: "Your account has been configured successfully.",
      });

      onComplete();
    } catch (error: unknown) {
      const err = error as { response?: { data?: { message?: string } }; message?: string };
      toast({
        title: "Error",
        description: err.response?.data?.message || err.message || "Failed to save settings",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const renderStep = () => {
    switch (currentStep) {
      case 0:
        return (
          <div className="space-y-6 text-center">
            <div className="flex justify-center">
              <div className="rounded-full bg-primary/10 p-4">
                <Sparkles className="w-12 h-12 text-primary" />
              </div>
            </div>
            <div>
              <h2 className="text-2xl font-bold mb-2">Welcome to OpenMind Practice!</h2>
              <p className="text-muted-foreground">
                Let's get your practice set up in just a few steps. This will only take a minute.
              </p>
            </div>
            <div className="space-y-3 text-left bg-muted/50 p-4 rounded-lg">
              <div className="flex items-center gap-3">
                <CheckCircle2 className="w-5 h-5 text-primary flex-shrink-0" />
                <span className="text-sm">Set up your organization details</span>
              </div>
              <div className="flex items-center gap-3">
                <CheckCircle2 className="w-5 h-5 text-primary flex-shrink-0" />
                <span className="text-sm">Complete your profile</span>
              </div>
              <div className="flex items-center gap-3">
                <CheckCircle2 className="w-5 h-5 text-primary flex-shrink-0" />
                <span className="text-sm">Learn about key features</span>
              </div>
            </div>
          </div>
        );

      case 1:
        return (
          <div className="space-y-6">
            <div>
              <h2 className="text-2xl font-bold mb-2">Organization Setup</h2>
              <p className="text-muted-foreground">
                Tell us about your practice or organization.
              </p>
            </div>
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="orgName">Organization Name</Label>
                <Input
                  id="orgName"
                  value={orgName}
                  onChange={(e) => setOrgName(e.target.value)}
                  placeholder="Enter your practice name"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="orgType">Organization Type</Label>
                <Select value={orgType} onValueChange={setOrgType}>
                  <SelectTrigger id="orgType">
                    <SelectValue placeholder="Select type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="clinic">Clinic</SelectItem>
                    <SelectItem value="private_practice">Private Practice</SelectItem>
                    <SelectItem value="group_practice">Group Practice</SelectItem>
                    <SelectItem value="hospital">Hospital</SelectItem>
                    <SelectItem value="other">Other</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>
        );

      case 2:
        return (
          <div className="space-y-6">
            <div>
              <h2 className="text-2xl font-bold mb-2">Complete Your Profile</h2>
              <p className="text-muted-foreground">
                Make sure your profile information is correct.
              </p>
            </div>
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="fullName">Full Name</Label>
                <Input
                  id="fullName"
                  value={fullName}
                  onChange={(e) => setFullName(e.target.value)}
                  placeholder="Enter your full name"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  value={user?.email || ""}
                  disabled
                  className="bg-muted"
                />
                <p className="text-xs text-muted-foreground">
                  Your email address cannot be changed here.
                </p>
              </div>
            </div>
          </div>
        );

      case 3:
        return (
          <div className="space-y-6 text-center">
            <div className="flex justify-center">
              <div className="rounded-full bg-primary/10 p-4">
                <CheckCircle2 className="w-12 h-12 text-primary" />
              </div>
            </div>
            <div>
              <h2 className="text-2xl font-bold mb-2">You're All Set!</h2>
              <p className="text-muted-foreground mb-6">
                Ready to start managing your practice? Here are some quick tips to get you started:
              </p>
            </div>
            <div className="space-y-3 text-left bg-muted/50 p-4 rounded-lg">
              <div className="flex items-start gap-3">
                <CheckCircle2 className="w-5 h-5 text-primary flex-shrink-0 mt-0.5" />
                <div>
                  <p className="font-medium text-sm">Add Your First Patient</p>
                  <p className="text-xs text-muted-foreground">
                    Start by adding patients to your practice from the Patients page.
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <CheckCircle2 className="w-5 h-5 text-primary flex-shrink-0 mt-0.5" />
                <div>
                  <p className="font-medium text-sm">Schedule Appointments</p>
                  <p className="text-xs text-muted-foreground">
                    Manage your calendar and schedule appointments with patients.
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <CheckCircle2 className="w-5 h-5 text-primary flex-shrink-0 mt-0.5" />
                <div>
                  <p className="font-medium text-sm">Create Clinical Notes</p>
                  <p className="text-xs text-muted-foreground">
                    Document patient sessions with encrypted, HIPAA-compliant notes.
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <CheckCircle2 className="w-5 h-5 text-primary flex-shrink-0 mt-0.5" />
                <div>
                  <p className="font-medium text-sm">Invite Team Members</p>
                  <p className="text-xs text-muted-foreground">
                    Collaborate with your team by inviting them to your organization.
                  </p>
                </div>
              </div>
            </div>
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <Card className="w-full max-w-2xl mx-auto">
      <CardHeader>
        <div className="space-y-2">
          <CardTitle>Get Started with OpenMind</CardTitle>
          <CardDescription>
            Step {currentStep + 1} of {totalSteps}
          </CardDescription>
          <Progress value={progress} className="mt-4" />
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {renderStep()}
        <div className="flex items-center justify-between pt-4 border-t">
          <div>
            {currentStep > 0 && (
              <Button variant="ghost" onClick={handleBack} disabled={loading}>
                <ArrowLeft className="w-4 h-4 mr-2" />
                Back
              </Button>
            )}
          </div>
          <div className="flex gap-2">
            <Button variant="outline" onClick={handleSkip} disabled={loading}>
              Skip for now
            </Button>
            <Button onClick={handleNext} disabled={loading}>
              {currentStep === totalSteps - 1 ? "Finish Setup" : "Next"}
              {currentStep < totalSteps - 1 && <ArrowRight className="w-4 h-4 ml-2" />}
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default OnboardingWizard;

