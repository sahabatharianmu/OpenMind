import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { AlertTriangle, ArrowRight, Check } from "lucide-react";
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
            <div className="bg-muted p-4 rounded-lg space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">{featureName}</span>
                <span className="text-sm text-muted-foreground">
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

          <div className="space-y-2">
            <p className="text-sm font-medium">Upgrade to Paid Plan to get:</p>
            <ul className="space-y-1 text-sm text-muted-foreground">
              <li className="flex items-center gap-2">
                <Check className="h-4 w-4 text-primary" />
                Unlimited {featureName.toLowerCase()}
              </li>
              <li className="flex items-center gap-2">
                <Check className="h-4 w-4 text-primary" />
                Unlimited team members
              </li>
              <li className="flex items-center gap-2">
                <Check className="h-4 w-4 text-primary" />
                All premium features
              </li>
            </ul>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Maybe Later
          </Button>
          <Button onClick={handleUpgrade}>
            Upgrade Now <ArrowRight className="ml-2 h-4 w-4" />
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default UpgradePrompt;

