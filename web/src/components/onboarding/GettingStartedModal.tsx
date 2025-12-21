import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { ArrowRight, ArrowLeft, Users, Calendar, FileText, UserPlus, Sparkles } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { subscriptionService } from "@/services/subscriptionService";

interface GettingStartedModalProps {
  onClose: () => void;
}

interface TourStep {
  id: number;
  title: string;
  description: string;
  icon: React.ReactNode;
  feature: string;
}

const tourSteps: TourStep[] = [
  {
    id: 1,
    title: "Manage Patients",
    description: "Add and manage your patients. Keep track of their information, appointments, and clinical history all in one place.",
    icon: <Users className="w-8 h-8" />,
    feature: "patients",
  },
  {
    id: 2,
    title: "Schedule Appointments",
    description: "Organize your calendar and schedule appointments with patients. Set reminders and manage your availability.",
    icon: <Calendar className="w-8 h-8" />,
    feature: "appointments",
  },
  {
    id: 3,
    title: "Create Clinical Notes",
    description: "Document patient sessions with encrypted, HIPAA-compliant clinical notes. Your data is secure and private.",
    icon: <FileText className="w-8 h-8" />,
    feature: "notes",
  },
  {
    id: 4,
    title: "Invite Team Members",
    description: "Collaborate with your team by inviting clinicians, case managers, and administrators to your organization.",
    icon: <UserPlus className="w-8 h-8" />,
    feature: "teams",
  },
];

const GettingStartedModal = ({ onClose }: GettingStartedModalProps) => {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState(0);
  const [dontShowAgain, setDontShowAgain] = useState(false);
  const [tier, setTier] = useState<string>("free");

  useEffect(() => {
    const checkTier = async () => {
      try {
        const currentTier = await subscriptionService.getSubscriptionTier();
        setTier(currentTier);
      } catch (error) {
        console.error("Failed to check tier", error);
      }
    };
    checkTier();
  }, []);

  const progress = ((currentStep + 1) / tourSteps.length) * 100;
  const currentTourStep = tourSteps[currentStep];

  const handleNext = () => {
    if (currentStep < tourSteps.length - 1) {
      setCurrentStep(currentStep + 1);
    } else {
      handleFinish();
    }
  };

  const handleBack = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  const handleFinish = () => {
    if (dontShowAgain) {
      localStorage.setItem("getting_started_shown", "true");
    }
    onClose();
  };

  const handleSkip = () => {
    if (dontShowAgain) {
      localStorage.setItem("getting_started_shown", "true");
    }
    onClose();
  };

  return (
    <Dialog open={true} onOpenChange={() => handleSkip()}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Getting Started with Closaf</DialogTitle>
          <DialogDescription>
            Let's take a quick tour of the key features
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          <Progress value={progress} className="w-full" />

          <div className="flex flex-col items-center space-y-4 text-center">
            <div className="rounded-full bg-primary/10 p-4">
              <div className="text-primary">{currentTourStep.icon}</div>
            </div>
            <div>
              <h3 className="text-xl font-semibold mb-2">{currentTourStep.title}</h3>
              <p className="text-sm text-muted-foreground">{currentTourStep.description}</p>
            </div>
          </div>

          {/* Free Tier Info & Upgrade CTA */}
          {tier === "free" && currentStep === tourSteps.length - 1 && (
            <div className="p-4 bg-primary/5 border border-primary/20 rounded-lg">
              <div className="flex items-start gap-3">
                <Sparkles className="w-5 h-5 text-primary mt-0.5" />
                <div className="flex-1">
                  <p className="text-sm font-medium mb-1">You're on the Free Tier</p>
                  <p className="text-xs text-muted-foreground mb-3">
                    Start with 10 patients and 1 team member. Upgrade anytime to unlock unlimited growth.
                  </p>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      onClose();
                      navigate("/pricing");
                    }}
                    className="w-full"
                  >
                    Learn About Upgrading
                  </Button>
                </div>
              </div>
            </div>
          )}

          <div className="flex items-center justify-between pt-4 border-t">
            <div>
              {currentStep > 0 && (
                <Button variant="ghost" onClick={handleBack}>
                  <ArrowLeft className="w-4 h-4 mr-2" />
                  Back
                </Button>
              )}
            </div>
            <div className="flex items-center gap-4">
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="dont-show"
                  checked={dontShowAgain}
                  onCheckedChange={(checked) => setDontShowAgain(checked === true)}
                />
                <Label
                  htmlFor="dont-show"
                  className="text-sm font-normal cursor-pointer"
                >
                  Don't show again
                </Label>
              </div>
              <div className="flex gap-2">
                <Button variant="outline" onClick={handleSkip}>
                  Skip Tour
                </Button>
                <Button onClick={handleNext}>
                  {currentStep === tourSteps.length - 1 ? "Get Started" : "Next"}
                  {currentStep < tourSteps.length - 1 && (
                    <ArrowRight className="w-4 h-4 ml-2" />
                  )}
                </Button>
              </div>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default GettingStartedModal;

