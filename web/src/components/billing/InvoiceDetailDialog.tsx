import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { 
  User, 
  Calendar, 
  DollarSign, 
  CreditCard,
  FileText,
  CheckCircle
} from "lucide-react";
import { format } from "date-fns";
import { useToast } from "@/hooks/use-toast";
import invoiceService from "@/services/invoiceService";
import { Invoice } from "@/types";

interface UIInvoice extends Invoice {
  patients: {
    first_name: string;
    last_name: string;
  };
  appointments?: {
    start_time: string;
    appointment_type: string;
  } | null;
}

interface InvoiceDetailDialogProps {
  invoice: UIInvoice | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onUpdate: () => void;
}

const paymentMethods = [
  { value: "cash", label: "Cash" },
  { value: "check", label: "Check" },
  { value: "credit_card", label: "Credit Card" },
  { value: "insurance", label: "Insurance" },
  { value: "other", label: "Other" },
];

export const InvoiceDetailDialog = ({
  invoice,
  open,
  onOpenChange,
  onUpdate,
}: InvoiceDetailDialogProps) => {
  const { toast } = useToast();
  const [isRecordingPayment, setIsRecordingPayment] = useState(false);
  const [paymentMethod, setPaymentMethod] = useState<string>("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  if (!invoice) return null;

  const formatCurrency = (cents: number) => {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
    }).format(cents / 100);
  };

  const handleRecordPayment = async () => {
    if (!paymentMethod) {
      toast({
        title: "Payment method required",
        description: "Please select a payment method.",
        variant: "destructive",
      });
      return;
    }

    setIsSubmitting(true);

    try {
      await invoiceService.update(invoice.id, {
        status: "paid",
        paid_at: new Date().toISOString(),
        payment_method: paymentMethod,
      });

      toast({
        title: "Payment Recorded",
        description: "Invoice has been marked as paid.",
      });
      setIsRecordingPayment(false);
      setPaymentMethod("");
      onUpdate();
    } catch (error) {
       toast({
        title: "Error",
        description: "Failed to record payment.",
        variant: "destructive",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const getStatusBadge = (status: string) => {
    const variants: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
      draft: "secondary",
      sent: "outline",
      paid: "default",
      overdue: "destructive",
      cancelled: "secondary",
    };

    return (
      <Badge variant={variants[status] || "secondary"}>
        {status.charAt(0).toUpperCase() + status.slice(1)}
      </Badge>
    );
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <FileText className="w-5 h-5" />
            Invoice Details
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {/* Patient Info */}
          <div className="flex items-start gap-3 p-3 rounded-lg bg-muted/50">
            <User className="w-5 h-5 text-muted-foreground mt-0.5" />
            <div>
              <p className="font-medium">
                {invoice.patients.first_name} {invoice.patients.last_name}
              </p>
              <p className="text-sm text-muted-foreground">Patient</p>
            </div>
          </div>

          {/* Amount & Status */}
          <div className="grid grid-cols-2 gap-4">
            <div className="flex items-start gap-3 p-3 rounded-lg bg-muted/50">
              <DollarSign className="w-5 h-5 text-muted-foreground mt-0.5" />
              <div>
                <p className="font-medium text-lg">{formatCurrency(invoice.amount_cents)}</p>
                <p className="text-sm text-muted-foreground">Amount</p>
              </div>
            </div>
            <div className="flex items-start gap-3 p-3 rounded-lg bg-muted/50">
              <CheckCircle className="w-5 h-5 text-muted-foreground mt-0.5" />
              <div>
                {getStatusBadge(invoice.status)}
                <p className="text-sm text-muted-foreground mt-1">Status</p>
              </div>
            </div>
          </div>

          {/* Dates */}
          <div className="grid grid-cols-2 gap-4">
            <div className="flex items-start gap-3 p-3 rounded-lg bg-muted/50">
              <Calendar className="w-5 h-5 text-muted-foreground mt-0.5" />
              <div>
                <p className="font-medium">
                  {format(new Date(invoice.created_at), "MMM d, yyyy")}
                </p>
                <p className="text-sm text-muted-foreground">Created</p>
              </div>
            </div>
            <div className="flex items-start gap-3 p-3 rounded-lg bg-muted/50">
              <Calendar className="w-5 h-5 text-muted-foreground mt-0.5" />
              <div>
                <p className="font-medium">
                  {invoice.due_date
                    ? format(new Date(invoice.due_date), "MMM d, yyyy")
                    : "Not set"}
                </p>
                <p className="text-sm text-muted-foreground">Due Date</p>
              </div>
            </div>
          </div>

          {/* Payment Info (if paid) */}
          {invoice.status === "paid" && (
            <div className="flex items-start gap-3 p-3 rounded-lg bg-green-500/10 border border-green-500/20">
              <CreditCard className="w-5 h-5 text-green-600 mt-0.5" />
              <div>
                <p className="font-medium text-green-600">
                  Paid on {invoice.paid_at ? format(new Date(invoice.paid_at), "MMM d, yyyy") : "N/A"}
                </p>
                {invoice.payment_method && (
                  <p className="text-sm text-muted-foreground">
                    via {paymentMethods.find(m => m.value === invoice.payment_method)?.label || invoice.payment_method}
                  </p>
                )}
              </div>
            </div>
          )}

          {/* Notes */}
          {invoice.notes && (
            <>
              <Separator />
              <div>
                <Label className="text-muted-foreground">Notes</Label>
                <p className="mt-1">{invoice.notes}</p>
              </div>
            </>
          )}

          {/* Record Payment Section */}
          {["draft", "sent", "overdue"].includes(invoice.status) && (
            <>
              <Separator />
              {isRecordingPayment ? (
                <div className="space-y-3">
                  <Label>Payment Method</Label>
                  <Select value={paymentMethod} onValueChange={setPaymentMethod}>
                    <SelectTrigger>
                      <SelectValue placeholder="Select payment method" />
                    </SelectTrigger>
                    <SelectContent>
                      {paymentMethods.map((method) => (
                        <SelectItem key={method.value} value={method.value}>
                          {method.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <div className="flex gap-2">
                    <Button
                      onClick={handleRecordPayment}
                      disabled={isSubmitting || !paymentMethod}
                      className="flex-1"
                    >
                      {isSubmitting ? "Recording..." : "Confirm Payment"}
                    </Button>
                    <Button
                      variant="outline"
                      onClick={() => {
                        setIsRecordingPayment(false);
                        setPaymentMethod("");
                      }}
                    >
                      Cancel
                    </Button>
                  </div>
                </div>
              ) : (
                <Button
                  onClick={() => setIsRecordingPayment(true)}
                  className="w-full gap-2"
                >
                  <CreditCard className="w-4 h-4" />
                  Record Payment
                </Button>
              )}
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
};
