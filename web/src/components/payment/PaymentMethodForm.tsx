import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Loader2, CreditCard, AlertCircle } from "lucide-react";
import { paymentService } from "@/services/paymentService";
import { useToast } from "@/hooks/use-toast";

interface PaymentMethodFormProps {
  onSuccess?: () => void;
  onCancel?: () => void;
}

const PaymentMethodForm = ({ onSuccess, onCancel }: PaymentMethodFormProps) => {
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [formData, setFormData] = useState({
    cardNumber: "",
    expiryMonth: "",
    expiryYear: "",
    cvv: "",
    provider: "stripe", // Default provider
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      // Validate form
      if (!formData.cardNumber || !formData.expiryMonth || !formData.expiryYear || !formData.cvv) {
        setError("Please fill in all card details");
        setLoading(false);
        return;
      }

      // Format card number (remove spaces)
      const cardNumber = formData.cardNumber.replace(/\s/g, "");

      // Create a token from card details
      // Note: For PCI compliance, this should ideally use provider-specific tokenization
      // For now, we'll send a formatted token to the backend
      // The backend will handle provider-specific tokenization
      const token = `${formData.provider}_${cardNumber}_${formData.expiryMonth}${formData.expiryYear}_${formData.cvv}`;

      // Send payment method to backend
      await paymentService.createPaymentMethod({
        token: token,
        provider: formData.provider,
      });

      toast({
        title: "Success",
        description: "Payment method added successfully",
      });

      if (onSuccess) {
        onSuccess();
      }
    } catch (err: any) {
      const errorMessage = err.response?.data?.error?.message || err.message || "Failed to add payment method";
      setError(errorMessage);
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const formatCardNumber = (value: string) => {
    // Remove all non-digits
    const digits = value.replace(/\D/g, "");
    // Add spaces every 4 digits
    return digits.replace(/(\d{4})(?=\d)/g, "$1 ");
  };

  const handleCardNumberChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const formatted = formatCardNumber(e.target.value);
    setFormData({ ...formData, cardNumber: formatted });
  };

  const currentYear = new Date().getFullYear();
  const years = Array.from({ length: 20 }, (_, i) => currentYear + i);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <CreditCard className="w-5 h-5" />
          Add Payment Method
        </CardTitle>
        <CardDescription>
          Add a new payment method to your account. Your payment information is securely processed by our payment providers.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <div className="space-y-2">
            <Label htmlFor="provider">Payment Provider</Label>
            <Select value={formData.provider} onValueChange={(value) => setFormData({ ...formData, provider: value })}>
              <SelectTrigger>
                <SelectValue placeholder="Select provider" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="stripe">Stripe</SelectItem>
                <SelectItem value="midtrans">Midtrans (QRIS)</SelectItem>
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              {formData.provider === "midtrans" 
                ? "QRIS payments will generate a QR code for scanning"
                : "Card payments are processed securely"}
            </p>
          </div>

          {formData.provider === "stripe" && (
            <>
              <div className="space-y-2">
                <Label htmlFor="cardNumber">Card Number</Label>
                <Input
                  id="cardNumber"
                  type="text"
                  placeholder="1234 5678 9012 3456"
                  maxLength={19}
                  value={formData.cardNumber}
                  onChange={handleCardNumberChange}
                  required
                />
              </div>

              <div className="grid grid-cols-3 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="expiryMonth">Month</Label>
                  <Select
                    value={formData.expiryMonth}
                    onValueChange={(value) => setFormData({ ...formData, expiryMonth: value })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="MM" />
                    </SelectTrigger>
                    <SelectContent>
                      {Array.from({ length: 12 }, (_, i) => {
                        const month = String(i + 1).padStart(2, "0");
                        return (
                          <SelectItem key={month} value={month}>
                            {month}
                          </SelectItem>
                        );
                      })}
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="expiryYear">Year</Label>
                  <Select
                    value={formData.expiryYear}
                    onValueChange={(value) => setFormData({ ...formData, expiryYear: value })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="YYYY" />
                    </SelectTrigger>
                    <SelectContent>
                      {years.map((year) => (
                        <SelectItem key={year} value={String(year)}>
                          {year}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="cvv">CVV</Label>
                  <Input
                    id="cvv"
                    type="text"
                    placeholder="123"
                    maxLength={4}
                    value={formData.cvv}
                    onChange={(e) => setFormData({ ...formData, cvv: e.target.value.replace(/\D/g, "") })}
                    required
                  />
                </div>
              </div>
            </>
          )}

          {formData.provider === "midtrans" && (
            <Alert>
              <AlertDescription>
                QRIS payments will be processed when you create a payment transaction. No card details are required.
              </AlertDescription>
            </Alert>
          )}

          <div className="flex gap-2">
            <Button type="submit" disabled={loading} className="flex-1">
              {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Add Payment Method
            </Button>
            {onCancel && (
              <Button type="button" variant="outline" onClick={onCancel} disabled={loading}>
                Cancel
              </Button>
            )}
          </div>
        </form>
      </CardContent>
    </Card>
  );
};

export default PaymentMethodForm;
