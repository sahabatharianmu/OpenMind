import { useState, useEffect } from "react";
import { format } from "date-fns";
import { 
  CreditCard, 
  ArrowRight, 
  Clock, 
  CheckCircle2, 
  XCircle,
  AlertCircle,
  RefreshCw,
  Zap,
  Ticket,
  ExternalLink
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from "@/components/ui/card";
import { useToast } from "@/hooks/use-toast";
import { paymentService, type PaymentTransaction } from "@/services/paymentService";
import { useFormatters } from "@/hooks/useFormatters";
import type { Organization } from "@/services/organizationService";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import QRISPaymentForm from "@/components/payment/QRISPaymentForm";

const getStatusBadgeVariant = (status: string) => {
  switch (status.toLowerCase()) {
    case 'paid':
      return 'default'; // Or a custom green success variant if you have it
    case 'pending':
      return 'secondary';
    case 'failed':
    case 'cancelled':
      return 'destructive';
    default:
      return 'outline';
  }
};

const getStatusIcon = (status: string) => {
  switch (status.toLowerCase()) {
    case 'paid':
      return <CheckCircle2 className="w-3 h-3 mr-1" />;
    case 'pending':
      return <Clock className="w-3 h-3 mr-1" />;
    case 'failed':
    case 'cancelled':
      return <XCircle className="w-3 h-3 mr-1" />;
    default:
      return <AlertCircle className="w-3 h-3 mr-1" />;
  }
};

const getPaymentMethodLabel = (method: string) => {
  switch (method.toLowerCase()) {
    case 'qris': return 'QRIS';
    case 'credit_card': return 'Credit Card';
    case 'bank_transfer': return 'Bank Transfer';
    case 'virtual_account': return 'Virtual Account';
    default: return method.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase());
  }
};

interface SubscriptionBillingProps {
  organization: Organization | null;
}

export default function SubscriptionBilling({ organization }: SubscriptionBillingProps) {
  const { toast } = useToast();
  const { formatCurrencyRaw } = useFormatters();
  const [transactions, setTransactions] = useState<PaymentTransaction[]>([]);
  const [loading, setLoading] = useState(true);
  
  // Pagination
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const limit = 10;
  const totalPages = Math.ceil(total / limit) || 1;

  // Resuming Payment State
  const [resumeTransactionId, setResumeTransactionId] = useState<string | null>(null);

  useEffect(() => {
    fetchTransactions();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page]);

  const fetchTransactions = async () => {
    setLoading(true);
    try {
      const offset = (page - 1) * limit;
      const resp = await paymentService.listTransactions(limit, offset);
      setTransactions(resp.data || []);
      setTotal(resp.total || 0);
    } catch (error: any) {
      toast({
        title: "Failed to load transactions",
        description: error.message || "An unexpected error occurred.",
        variant: "destructive"
      });
    } finally {
      setLoading(false);
    }
  };

  const handleResumePayment = (transactionId: string) => {
    setResumeTransactionId(transactionId);
  };

  const handlePaymentSuccess = () => {
    toast({
      title: "Payment Successful",
      description: "Your subscription has been upgraded.",
    });
    setResumeTransactionId(null);
    fetchTransactions();
  };

  const handlePaymentCancel = () => {
    setResumeTransactionId(null);
    fetchTransactions(); // Refresh in case it was explicitly cancelled
  };

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      
      {/* Overview Cards */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card className="bg-gradient-to-br from-primary/10 via-background to-background border-primary/20 shadow-sm relative overflow-hidden">
          <div className="absolute -right-4 -top-4 w-24 h-24 bg-primary/10 rounded-full blur-2xl pointer-events-none" />
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <Zap className="w-4 h-4 text-primary" /> Current Plan
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold tracking-tight capitalize">
              {organization?.subscription_tier ? `${organization.subscription_tier} Tier` : 'Loading...'}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              Active until end of billing cycle
            </p>
          </CardContent>
        </Card>
        
        <Card className="shadow-sm">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <Ticket className="w-4 h-4" /> Next Invoice
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold tracking-tight">{formatCurrencyRaw(20, 'USD')}</div>
            <p className="text-xs text-muted-foreground mt-1">
              Due on Dec 1st, 2026
            </p>
          </CardContent>
        </Card>

        <Card className="shadow-sm">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <CreditCard className="w-4 h-4" /> Payment Method
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold tracking-tight">QRIS</div>
            <p className="text-xs text-muted-foreground mt-1 cursor-pointer hover:text-primary transition-colors inline-flex items-center gap-1">
              Manage methods <ExternalLink className="w-3 h-3" />
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Transaction History Table */}
      <Card className="shadow-sm border-border/50">
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <div className="space-y-1">
            <CardTitle>Billing History</CardTitle>
            <CardDescription>
              View your recent subscription payments and invoices
            </CardDescription>
          </div>
          <Button variant="outline" size="sm" onClick={fetchTransactions} disabled={loading}>
            <RefreshCw className={`w-4 h-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </CardHeader>
        <CardContent>
          <div className="rounded-md border overflow-hidden">
            <Table>
              <TableHeader className="bg-muted/50">
                <TableRow>
                  <TableHead>Date</TableHead>
                  <TableHead>Description</TableHead>
                  <TableHead>Amount</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Method</TableHead>
                  <TableHead className="text-right">Action</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading ? (
                  Array.from({ length: 3 }).map((_, i) => (
                    <TableRow key={`skeleton-${i}`}>
                      <TableCell><div className="h-4 w-24 bg-muted animate-pulse rounded" /></TableCell>
                      <TableCell><div className="h-4 w-32 bg-muted animate-pulse rounded" /></TableCell>
                      <TableCell><div className="h-4 w-16 bg-muted animate-pulse rounded" /></TableCell>
                      <TableCell><div className="h-5 w-20 bg-muted animate-pulse rounded-full" /></TableCell>
                      <TableCell><div className="h-4 w-16 bg-muted animate-pulse rounded" /></TableCell>
                      <TableCell><div className="h-8 w-20 bg-muted animate-pulse rounded ml-auto" /></TableCell>
                    </TableRow>
                  ))
                ) : transactions.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="h-32 text-center text-muted-foreground">
                      No transaction history found.
                    </TableCell>
                  </TableRow>
                ) : (
                  transactions.map((tx) => (
                    <TableRow key={tx.id} className="group hover:bg-muted/30 transition-colors">
                      <TableCell className="font-medium">
                        {format(new Date(tx.created_at), "MMM d, yyyy")}
                      </TableCell>
                      <TableCell>
                        <div className="flex flex-col">
                          <span>{tx.type === 'subscription' ? 'Pro Plan Subscription' : 'One-time Payment'}</span>
                          <span className="text-xs text-muted-foreground truncate max-w-[200px]" title={tx.partner_reference_no}>
                            Ref: {tx.partner_reference_no}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell className="font-medium whitespace-nowrap">
                        {formatCurrencyRaw(tx.amount, tx.currency)}
                      </TableCell>
                      <TableCell>
                        <Badge 
                          variant={getStatusBadgeVariant(tx.status) as any}
                          className="capitalize inline-flex items-center px-2 py-0.5"
                        >
                          {getStatusIcon(tx.status)}
                          {tx.status}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {getPaymentMethodLabel(tx.payment_method)}
                      </TableCell>
                      <TableCell className="text-right">
                        {tx.status === 'pending' && tx.payment_method === 'qris' ? (
                          <Button 
                            variant="default" 
                            size="sm" 
                            className="bg-primary hover:bg-primary/90 shadow-sm transition-all group-hover:shadow-md"
                            onClick={() => handleResumePayment(tx.id)}
                          >
                            Pay Now <ArrowRight className="w-3 h-3 ml-1.5" />
                          </Button>
                        ) : tx.status === 'paid' ? (
                          <Button variant="ghost" size="sm" className="text-muted-foreground hover:text-foreground">
                            Receipt
                          </Button>
                        ) : null}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
          
          {/* Pagination Controls */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between mt-4">
              <p className="text-sm text-muted-foreground">
                Showing {transactions.length} of {total} transactions
              </p>
              <div className="flex gap-2">
                <Button 
                  variant="outline" 
                  size="sm" 
                  onClick={() => setPage(p => Math.max(1, p - 1))}
                  disabled={page === 1 || loading}
                >
                  Previous
                </Button>
                <Button 
                  variant="outline" 
                  size="sm"
                  onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                  disabled={page >= totalPages || loading}
                >
                  Next
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Resume Payment Modal */}
      <Dialog open={!!resumeTransactionId} onOpenChange={(open) => !open && setResumeTransactionId(null)}>
        <DialogContent className="sm:max-w-md p-0 overflow-hidden border-none shadow-2xl">
          {resumeTransactionId && (
            <QRISPaymentForm
              existingTransactionId={resumeTransactionId}
              onSuccess={handlePaymentSuccess}
              onCancel={handlePaymentCancel}
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
