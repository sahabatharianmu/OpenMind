import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Loader2, CreditCard, AlertCircle } from "lucide-react";
import { paymentService } from "@/services/paymentService";
import { useToast } from "@/hooks/use-toast";

interface CreditCardPaymentFormProps {
  amount: number;
  onSuccess: () => void;
  onError: (error: string) => void;
  onCancel: () => void;
}

const CreditCardPaymentForm = ({ amount, onSuccess, onError, onCancel }: CreditCardPaymentFormProps) => {
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [formData, setFormData] = useState({
    cardNumber: "",
    expiryMonth: "",
    expiryYear: "",
    cvv: "",
    cardholderName: "",
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      // Validate form
      if (!formData.cardNumber || !formData.expiryMonth || !formData.expiryYear || !formData.cvv || !formData.cardholderName) {
        setError("Please fill in all card details");
        setLoading(false);
        return;
      }

      // Format card number (remove spaces)
      const cardNumber = formData.cardNumber.replace(/\s/g, "");

      // Create a token from card details
      // Note: For PCI compliance, this should use provider-specific tokenization
      // The backend will handle provider-specific tokenization
      const token = `stripe_${cardNumber}_${formData.expiryMonth}${formData.expiryYear}_${formData.cvv}`;

      // Create payment method
      await paymentService.createPaymentMethod({
        token: token,
        provider: "stripe",
      });

      // TODO: Create subscription with this payment method
      // For now, just show success
      toast({
        title: "Payment Method Added",
        description: "Your card has been added. Processing subscription...",
      });

      // Simulate subscription creation
      // In production, this would call the subscription API
      setTimeout(() => {
        onSuccess();
      }, 1000);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error?.message || err.message || "Failed to process payment";
      setError(errorMessage);
      onError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const formatCardNumber = (value: string) => {
    const digits = value.replace(/\D/g, "");
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
          Credit Card Payment
        </CardTitle>
        <CardDescription>
          Pay ${amount.toFixed(2)}/month using your credit card
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
            <Label htmlFor="cardholderName">Cardholder Name</Label>
            <Input
              id="cardholderName"
              type="text"
              placeholder="John Doe"
              value={formData.cardholderName}
              onChange={(e) => setFormData({ ...formData, cardholderName: e.target.value })}
              required
            />
          </div>

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

          <div className="flex gap-2 pt-4">
            <Button type="button" variant="outline" onClick={onCancel} disabled={loading} className="flex-1">
              Cancel
            </Button>
            <Button type="submit" disabled={loading} className="flex-1">
              {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Pay ${amount.toFixed(2)}/month
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
};

export default CreditCardPaymentForm;

