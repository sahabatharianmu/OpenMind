import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Loader2, Building2, Copy, CheckCircle2, AlertCircle } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import api from "@/api/client";

interface VirtualAccountPaymentFormProps {
  amount: number;
  onSuccess: () => void;
  onError: (error: string) => void;
  onCancel: () => void;
}

interface VirtualAccountResponse {
  virtual_account_number: string;
  bank_name: string;
  account_name: string;
  expiry_date: string;
  transaction_id: string;
  order_id: string;
}

const VirtualAccountPaymentForm = ({ amount, onSuccess, onError, onCancel }: VirtualAccountPaymentFormProps) => {
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);
  const [vaData, setVaData] = useState<VirtualAccountResponse | null>(null);
  const [copied, setCopied] = useState(false);
  const [isPolling, setIsPolling] = useState(false);

  const createVirtualAccount = async () => {
    setLoading(true);
    try {
      // TODO: Replace with actual API endpoint for creating virtual account
      const response = await api.post<{ data: VirtualAccountResponse }>("/payments/virtual-account/create", {
        amount: amount,
        currency: "USD",
      });

      setVaData(response.data.data);
      setIsPolling(true);
      // Start polling for payment status
      pollPaymentStatus(response.data.data.transaction_id);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error?.message || err.message || "Failed to create virtual account";
      onError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const pollPaymentStatus = async (transactionId: string) => {
    const maxAttempts = 120; // Poll for up to 10 minutes (5 second intervals)
    let attempts = 0;

    const poll = async () => {
      if (attempts >= maxAttempts) {
        setIsPolling(false);
        toast({
          title: "Payment Timeout",
          description: "Payment verification timed out. Please check your payment status.",
          variant: "destructive",
        });
        return;
      }

      try {
        // TODO: Replace with actual API endpoint for checking payment status
        const response = await api.get<{ data: { status: string } }>(`/payments/virtual-account/status/${transactionId}`);
        
        if (response.data.data.status === "paid" || response.data.data.status === "settled") {
          setIsPolling(false);
          onSuccess();
          return;
        }

        attempts++;
        setTimeout(poll, 5000); // Poll every 5 seconds
      } catch (err) {
        console.error("Error polling payment status:", err);
        attempts++;
        setTimeout(poll, 5000);
      }
    };

    poll();
  };

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      toast({
        title: "Copied",
        description: "Virtual account number copied to clipboard",
      });
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      toast({
        title: "Error",
        description: "Failed to copy to clipboard",
        variant: "destructive",
      });
    }
  };

  useEffect(() => {
    if (!vaData) {
      createVirtualAccount();
    }
  }, []);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Building2 className="w-5 h-5" />
          Virtual Account Payment
        </CardTitle>
        <CardDescription>
          Transfer ${amount.toFixed(2)} to the virtual account below
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {loading && !vaData && (
          <div className="flex flex-col items-center justify-center py-8">
            <Loader2 className="w-8 h-8 animate-spin text-primary mb-4" />
            <p className="text-muted-foreground">Creating virtual account...</p>
          </div>
        )}

        {vaData && (
          <>
            <Alert>
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                Please complete the payment within the expiry time. The virtual account will expire on{" "}
                {new Date(vaData.expiry_date).toLocaleString()}.
              </AlertDescription>
            </Alert>

            <div className="space-y-4">
              <div className="p-4 border rounded-lg bg-muted/50">
                <div className="space-y-3">
                  <div>
                    <Label className="text-sm text-muted-foreground">Bank</Label>
                    <p className="text-lg font-semibold">{vaData.bank_name}</p>
                  </div>

                  <div>
                    <Label className="text-sm text-muted-foreground">Virtual Account Number</Label>
                    <div className="flex items-center gap-2">
                      <p className="text-lg font-mono font-semibold">{vaData.virtual_account_number}</p>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => copyToClipboard(vaData.virtual_account_number)}
                        className="h-8 w-8 p-0"
                      >
                        {copied ? (
                          <CheckCircle2 className="h-4 w-4 text-green-500" />
                        ) : (
                          <Copy className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                  </div>

                  <div>
                    <Label className="text-sm text-muted-foreground">Account Name</Label>
                    <p className="text-lg font-semibold">{vaData.account_name}</p>
                  </div>

                  <div>
                    <Label className="text-sm text-muted-foreground">Amount</Label>
                    <p className="text-2xl font-bold">${amount.toFixed(2)}</p>
                  </div>
                </div>
              </div>

              {isPolling && (
                <Alert>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  <AlertDescription>
                    Waiting for payment confirmation. Please complete the transfer to the virtual account above.
                  </AlertDescription>
                </Alert>
              )}

              <div className="text-center space-y-1">
                <p className="text-xs text-muted-foreground">Transaction ID</p>
                <p className="text-xs font-mono">{vaData.transaction_id}</p>
              </div>
            </div>

            <div className="flex gap-2 pt-4">
              <Button variant="outline" onClick={onCancel} disabled={isPolling} className="flex-1">
                Cancel
              </Button>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  );
};

export default VirtualAccountPaymentForm;

