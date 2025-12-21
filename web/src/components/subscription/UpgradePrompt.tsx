import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { AlertTriangle, ArrowRight, Check, Sparkles, Users, UserCheck, Zap } from "lucide-react";
import { UpgradePrompt as UpgradePromptType } from "@/services/subscriptionService";
import { useNavigate } from "react-router-dom";

interface UpgradePromptProps {
  isOpen: boolean;
  onClose: () => void;
  upgradePrompt?: UpgradePromptType | null;
  message?: string;
}

const UpgradePrompt = ({ isOpen, onClose, upgradePrompt, message }: UpgradePromptProps) => {
  const navigate = useNavigate();

  const handleUpgrade = () => {
    onClose();
    navigate("/pricing");
  };

  const featureName = upgradePrompt?.feature === "patients" ? "Patients" : "Team Members";
  const limitReached = upgradePrompt?.current >= upgradePrompt?.limit;

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <div className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-destructive" />
            <DialogTitle>Limit Reached</DialogTitle>
          </div>
          <DialogDescription>
            {message || `You've reached your ${featureName.toLowerCase()} limit.`}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {upgradePrompt && (
            <div className="bg-muted p-4 rounded-lg space-y-2 border border-destructive/20">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium flex items-center gap-2">
                  {upgradePrompt.feature === "patients" ? (
                    <Users className="h-4 w-4" />
                  ) : (
                    <UserCheck className="h-4 w-4" />
                  )}
                  {featureName}
                </span>
                <span className="text-sm font-semibold text-destructive">
                  {upgradePrompt.current} / {upgradePrompt.limit}
                </span>
              </div>
              <div className="text-sm text-muted-foreground">
                {limitReached
                  ? `You've reached the maximum of ${upgradePrompt.limit} ${featureName.toLowerCase()} on the Free plan.`
                  : `You're using ${upgradePrompt.current} of ${upgradePrompt.limit} ${featureName.toLowerCase()}.`}
              </div>
            </div>
          )}

          <div className="p-4 bg-primary/5 border border-primary/20 rounded-lg space-y-3">
            <div className="flex items-center gap-2">
              <Sparkles className="h-5 w-5 text-primary" />
              <p className="text-sm font-semibold">Upgrade to Paid Plan to get:</p>
            </div>
            <ul className="space-y-2 text-sm">
              <li className="flex items-center gap-2">
                <Check className="h-4 w-4 text-primary flex-shrink-0" />
                <span>Unlimited {featureName.toLowerCase()}</span>
              </li>
              <li className="flex items-center gap-2">
                <Check className="h-4 w-4 text-primary flex-shrink-0" />
                <span>Unlimited team members</span>
              </li>
              <li className="flex items-center gap-2">
                <Check className="h-4 w-4 text-primary flex-shrink-0" />
                <span>All premium features</span>
              </li>
              <li className="flex items-center gap-2">
                <Check className="h-4 w-4 text-primary flex-shrink-0" />
                <span>Priority support</span>
              </li>
            </ul>
          </div>
        </div>

        <DialogFooter className="flex-col sm:flex-row gap-2">
          <Button variant="outline" onClick={onClose} className="w-full sm:w-auto">
            Maybe Later
          </Button>
          <Button onClick={handleUpgrade} className="w-full sm:w-auto">
            Upgrade Now <ArrowRight className="ml-2 h-4 w-4" />
          </Button>
        </DialogFooter>
        <div className="text-center">
          <Button
            variant="link"
            onClick={() => {
              onClose();
              navigate("/pricing");
            }}
            className="text-xs"
          >
            Learn more about plans →
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default UpgradePrompt;

