import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Loader2, CreditCard, QrCode, Building2, CheckCircle2, Sparkles, Users, FileText, Zap } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { useToast } from "@/hooks/use-toast";
import CreditCardPaymentForm from "./CreditCardPaymentForm";
import QRISPaymentForm from "./QRISPaymentForm";
import VirtualAccountPaymentForm from "./VirtualAccountPaymentForm";

export type PaymentMethodType = "credit_card" | "qris" | "virtual_account";

interface UpgradeModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
  planPrice: number; // Monthly price in USD
}

const UpgradeModal = ({ open, onOpenChange, onSuccess, planPrice }: UpgradeModalProps) => {
  const { toast } = useToast();
  const navigate = useNavigate();
  const [selectedMethod, setSelectedMethod] = useState<PaymentMethodType>("credit_card");
  const [isProcessing, setIsProcessing] = useState(false);
  const [step, setStep] = useState<"select" | "payment" | "success">("select");

  const handleMethodSelect = (method: PaymentMethodType) => {
    setSelectedMethod(method);
    setStep("payment");
  };

  const handlePaymentSuccess = () => {
    setIsProcessing(false);
    setStep("success");
    // Set flag to show upgrade banner on dashboard
    localStorage.setItem("upgrade_success", "true");
    if (onSuccess) {
      onSuccess();
    }
  };

  const handleContinueToDashboard = () => {
    onOpenChange(false);
    // Reset state
    setStep("select");
    setSelectedMethod("credit_card");
    navigate("/dashboard");
  };

  const handlePaymentError = (error: string) => {
    setIsProcessing(false);
    toast({
      title: "Payment Failed",
      description: error,
      variant: "destructive",
    });
  };

  const handleBack = () => {
    setStep("select");
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Upgrade to Paid Plan</DialogTitle>
          <DialogDescription>
            Choose your preferred payment method to upgrade to the paid plan.
          </DialogDescription>
        </DialogHeader>

        {step === "select" && (
          <div className="space-y-4 py-4">
            <div className="mb-4">
              <div className="text-2xl font-bold mb-1">${planPrice.toFixed(2)}</div>
              <p className="text-sm text-muted-foreground">per month</p>
            </div>

            {/* What You Get Section */}
            <div className="mb-6 p-4 bg-primary/5 border border-primary/20 rounded-lg">
              <h3 className="font-semibold mb-3 flex items-center gap-2">
                <Sparkles className="w-4 h-4 text-primary" />
                What You Get
              </h3>
              <div className="grid grid-cols-1 gap-2 text-sm">
                <div className="flex items-center gap-2">
                  <Users className="w-4 h-4 text-primary" />
                  <span>Unlimited patients</span>
                </div>
                <div className="flex items-center gap-2">
                  <Users className="w-4 h-4 text-primary" />
                  <span>Unlimited team members</span>
                </div>
                <div className="flex items-center gap-2">
                  <FileText className="w-4 h-4 text-primary" />
                  <span>All core features</span>
                </div>
                <div className="flex items-center gap-2">
                  <Zap className="w-4 h-4 text-primary" />
                  <span>Priority support</span>
                </div>
              </div>
            </div>

            <Label className="text-base font-semibold">Select Payment Method</Label>
            <RadioGroup value={selectedMethod} onValueChange={(value) => setSelectedMethod(value as PaymentMethodType)}>
              <div className="space-y-3">
                {/* Credit Card Option */}
                <Card
                  className={`cursor-pointer transition-all ${
                    selectedMethod === "credit_card"
                      ? "border-primary bg-primary/5"
                      : "hover:border-primary/50"
                  }`}
                  onClick={() => handleMethodSelect("credit_card")}
                >
                  <CardContent className="p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <div className={`p-2 rounded-md ${
                          selectedMethod === "credit_card" ? "bg-primary text-primary-foreground" : "bg-muted"
                        }`}>
                          <CreditCard className="w-5 h-5" />
                        </div>
                        <div className="flex-1">
                          <div className="font-semibold">Credit Card</div>
                          <div className="text-sm text-muted-foreground">
                            Visa, Mastercard, Amex
                          </div>
                          <div className="text-xs text-muted-foreground mt-1">
                            Instant activation
                          </div>
                        </div>
                      </div>
                      <RadioGroupItem value="credit_card" id="credit_card" />
                    </div>
                  </CardContent>
                </Card>

                {/* QRIS Option */}
                <Card
                  className={`cursor-pointer transition-all ${
                    selectedMethod === "qris"
                      ? "border-primary bg-primary/5"
                      : "hover:border-primary/50"
                  }`}
                  onClick={() => handleMethodSelect("qris")}
                >
                  <CardContent className="p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <div className={`p-2 rounded-md ${
                          selectedMethod === "qris" ? "bg-primary text-primary-foreground" : "bg-muted"
                        }`}>
                          <QrCode className="w-5 h-5" />
                        </div>
                        <div className="flex-1">
                          <div className="font-semibold">QRIS</div>
                          <div className="text-sm text-muted-foreground">
                            Scan QR code to pay
                          </div>
                          <div className="text-xs text-muted-foreground mt-1">
                            Usually 1-2 minutes
                          </div>
                        </div>
                      </div>
                      <RadioGroupItem value="qris" id="qris" />
                    </div>
                  </CardContent>
                </Card>

                {/* Virtual Account Option */}
                <Card
                  className={`cursor-pointer transition-all ${
                    selectedMethod === "virtual_account"
                      ? "border-primary bg-primary/5"
                      : "hover:border-primary/50"
                  }`}
                  onClick={() => handleMethodSelect("virtual_account")}
                >
                  <CardContent className="p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <div className={`p-2 rounded-md ${
                          selectedMethod === "virtual_account" ? "bg-primary text-primary-foreground" : "bg-muted"
                        }`}>
                          <Building2 className="w-5 h-5" />
                        </div>
                        <div className="flex-1">
                          <div className="font-semibold">Virtual Account</div>
                          <div className="text-sm text-muted-foreground">
                            Bank transfer via virtual account
                          </div>
                          <div className="text-xs text-muted-foreground mt-1">
                            Usually 1-2 hours
                          </div>
                        </div>
                      </div>
                      <RadioGroupItem value="virtual_account" id="virtual_account" />
                    </div>
                  </CardContent>
                </Card>
              </div>
            </RadioGroup>

            <div className="flex gap-2 pt-4">
              <Button
                variant="outline"
                onClick={() => onOpenChange(false)}
                className="flex-1"
              >
                Cancel
              </Button>
              <Button
                onClick={() => handleMethodSelect(selectedMethod)}
                className="flex-1"
                disabled={!selectedMethod}
              >
                Continue
              </Button>
            </div>
          </div>
        )}

        {step === "payment" && (
          <div className="space-y-4 py-4">
            <Button
              variant="ghost"
              size="sm"
              onClick={handleBack}
              className="mb-2"
            >
              ← Back to payment methods
            </Button>

            {selectedMethod === "credit_card" && (
              <CreditCardPaymentForm
                amount={planPrice}
                onSuccess={handlePaymentSuccess}
                onError={handlePaymentError}
                onCancel={() => onOpenChange(false)}
              />
            )}

            {selectedMethod === "qris" && (
              <QRISPaymentForm
                amount={planPrice}
                onSuccess={handlePaymentSuccess}
                onError={handlePaymentError}
                onCancel={() => onOpenChange(false)}
              />
            )}

            {selectedMethod === "virtual_account" && (
              <VirtualAccountPaymentForm
                amount={planPrice}
                onSuccess={handlePaymentSuccess}
                onError={handlePaymentError}
                onCancel={() => onOpenChange(false)}
              />
            )}
          </div>
        )}

        {step === "success" && (
          <div className="space-y-6 py-4 text-center">
            <div className="flex justify-center">
              <div className="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center">
                <CheckCircle2 className="w-10 h-10 text-primary" />
              </div>
            </div>
            <div>
              <h3 className="text-2xl font-bold mb-2">Upgrade Successful! 🎉</h3>
              <p className="text-muted-foreground mb-6">
                Your subscription has been upgraded to the paid plan.
              </p>
            </div>
            <div className="p-4 bg-primary/5 border border-primary/20 rounded-lg text-left">
              <h4 className="font-semibold mb-3">You now have access to:</h4>
              <ul className="space-y-2 text-sm">
                <li className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-primary" />
                  <span>Unlimited patients</span>
                </li>
                <li className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-primary" />
                  <span>Unlimited team members</span>
                </li>
                <li className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-primary" />
                  <span>Priority support</span>
                </li>
                <li className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-primary" />
                  <span>All premium features</span>
                </li>
              </ul>
            </div>
            <Button onClick={handleContinueToDashboard} className="w-full" size="lg">
              Continue to Dashboard
            </Button>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
};

export default UpgradeModal;

