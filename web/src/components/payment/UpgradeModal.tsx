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
import { Loader2, CreditCard, QrCode, Building2, CheckCircle2 } from "lucide-react";
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
  const [selectedMethod, setSelectedMethod] = useState<PaymentMethodType>("credit_card");
  const [isProcessing, setIsProcessing] = useState(false);
  const [step, setStep] = useState<"select" | "payment">("select");

  const handleMethodSelect = (method: PaymentMethodType) => {
    setSelectedMethod(method);
    setStep("payment");
  };

  const handlePaymentSuccess = () => {
    setIsProcessing(false);
    toast({
      title: "Success",
      description: "Your subscription has been upgraded successfully!",
    });
    if (onSuccess) {
      onSuccess();
    }
    onOpenChange(false);
    // Reset state
    setStep("select");
    setSelectedMethod("credit_card");
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
                        <div>
                          <div className="font-semibold">Credit Card</div>
                          <div className="text-sm text-muted-foreground">
                            Visa, Mastercard, Amex
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
                        <div>
                          <div className="font-semibold">QRIS</div>
                          <div className="text-sm text-muted-foreground">
                            Scan QR code to pay
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
                        <div>
                          <div className="font-semibold">Virtual Account</div>
                          <div className="text-sm text-muted-foreground">
                            Bank transfer via virtual account
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
      </DialogContent>
    </Dialog>
  );
};

export default UpgradeModal;

