import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Loader2, QrCode, CheckCircle2, AlertCircle } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import api from "@/api/client";

interface QRISPaymentFormProps {
  amount: number;
  onSuccess: () => void;
  onError: (error: string) => void;
  onCancel: () => void;
}

interface QRISPaymentResponse {
  id: string;
  transaction_id: string;
  partner_reference_no: string;
  qr_code: string;
  qr_code_url: string;
  qr_code_image: string;
  amount: number;
  currency: string;
  status: string;
  expires_at?: string;
  created_at: string;
}

interface PaymentStatusResponse {
  id: string;
  transaction_id: string;
  status: string;
  amount: number;
  currency: string;
  paid_at?: string;
  latest_transaction_status?: string;
  transaction_status_desc?: string;
}

const QRISPaymentForm = ({ amount, onSuccess, onError, onCancel }: QRISPaymentFormProps) => {
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);
  const [qrData, setQrData] = useState<QRISPaymentResponse | null>(null);
  const [isPolling, setIsPolling] = useState(false);

  const createQRISPayment = async () => {
    setLoading(true);
    try {
      const response = await api.post<{ data: QRISPaymentResponse }>("/payments/qris/create", {
        amount: amount,
        currency: "USD",
        type: "subscription", // For subscription upgrade
      });

      setQrData(response.data.data);
      setIsPolling(true);
      // Start polling our own database (webhook updates it, we just check our DB)
      pollPaymentStatus(response.data.data.id);
    } catch (err) {
      let errorMessage = "Failed to create QRIS payment";
      if (err && typeof err === "object") {
        const axiosError = err as { response?: { data?: { error?: { message?: string } } }; message?: string };
        errorMessage = axiosError.response?.data?.error?.message || axiosError.message || errorMessage;
      } else if (err instanceof Error) {
        errorMessage = err.message;
      }
      onError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const pollPaymentStatus = async (transactionId: string) => {
    const maxAttempts = 60; // Poll for up to 5 minutes (5 second intervals)
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
        // Poll our own database (webhook updates it, we just check our DB - no Midtrans API calls)
        const response = await api.get<{ data: PaymentStatusResponse }>(`/payments/qris/status/${transactionId}`);
        
        if (response.data.data.status === "paid") {
          setIsPolling(false);
          toast({
            title: "Payment Successful",
            description: "Your payment has been confirmed!",
          });
          onSuccess();
          return;
        }

        if (response.data.data.status === "failed" || response.data.data.status === "cancelled" || response.data.data.status === "expired") {
          setIsPolling(false);
          toast({
            title: "Payment Failed",
            description: `Payment status: ${response.data.data.status}`,
            variant: "destructive",
          });
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

  useEffect(() => {
    if (!qrData) {
      createQRISPayment();
    }
  }, []);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <QrCode className="w-5 h-5" />
          QRIS Payment
        </CardTitle>
        <CardDescription>
          Scan the QR code below to complete your payment of ${amount.toFixed(2)}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {loading && !qrData && (
          <div className="flex flex-col items-center justify-center py-8">
            <Loader2 className="w-8 h-8 animate-spin text-primary mb-4" />
            <p className="text-muted-foreground">Generating QR code...</p>
          </div>
        )}

        {qrData && (
          <>
            <div className="flex flex-col items-center space-y-4">
              {qrData.qr_code_image ? (
                <img
                  src={`data:image/png;base64,${qrData.qr_code_image}`}
                  alt="QRIS Payment QR Code"
                  className="w-64 h-64 border rounded-lg p-4 bg-white"
                />
              ) : qrData.qr_code ? (
                <div className="w-64 h-64 border rounded-lg p-4 bg-white flex items-center justify-center">
                  <div className="text-center">
                    <QrCode className="w-32 h-32 text-muted-foreground mx-auto mb-2" />
                    <p className="text-xs text-muted-foreground">Scan with your QRIS app</p>
                    <p className="text-xs font-mono mt-2 break-all">{qrData.qr_code}</p>
                  </div>
                </div>
              ) : (
                <div className="w-64 h-64 border rounded-lg p-4 bg-white flex items-center justify-center">
                  <QrCode className="w-32 h-32 text-muted-foreground" />
                </div>
              )}

              {isPolling && (
                <Alert>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  <AlertDescription>
                    Waiting for payment confirmation. Please complete the payment by scanning the QR code.
                  </AlertDescription>
                </Alert>
              )}

              <div className="text-center space-y-2">
                <p className="text-sm text-muted-foreground">Amount</p>
                <p className="text-2xl font-bold">${amount.toFixed(2)}</p>
              </div>

              <div className="text-center space-y-1">
                <p className="text-xs text-muted-foreground">Transaction ID</p>
                <p className="text-xs font-mono">{qrData.transaction_id}</p>
                {qrData.expires_at && (
                  <p className="text-xs text-muted-foreground mt-2">
                    Expires: {new Date(qrData.expires_at).toLocaleString()}
                  </p>
                )}
              </div>
            </div>

            <div className="flex gap-2 pt-4">
                     <Button variant="outline" onClick={onCancel} disabled={isPolling} className="flex-1">
                       Cancel
                     </Button>
              {qrData.qr_code_url && (
                <Button
                  variant="outline"
                  onClick={() => window.open(qrData.qr_code_url, "_blank")}
                  className="flex-1"
                >
                  Open QR Code
                </Button>
              )}
            </div>
          </>
        )}
      </CardContent>
    </Card>
  );
};

export default QRISPaymentForm;

