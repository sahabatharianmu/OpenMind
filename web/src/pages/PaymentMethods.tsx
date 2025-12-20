import { useState, useEffect } from "react";
import DashboardLayout from "@/components/dashboard/DashboardLayout";
import PaymentMethodForm from "@/components/payment/PaymentMethodForm";
import PaymentMethodList from "@/components/payment/PaymentMethodList";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Plus, CreditCard } from "lucide-react";
import { paymentService, type PaymentMethod } from "@/services/paymentService";
import { useToast } from "@/hooks/use-toast";
import { useAuth } from "@/contexts/AuthContext";

const PaymentMethods = () => {
  const { user } = useAuth();
  const { toast } = useToast();
  const [paymentMethods, setPaymentMethods] = useState<PaymentMethod[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);

  // Check if user can manage payment methods (owner/admin only)
  const canManagePaymentMethods = user?.role === "admin" || user?.role === "owner";

  useEffect(() => {
    if (canManagePaymentMethods) {
      fetchPaymentMethods();
    }
  }, [canManagePaymentMethods]);

  const fetchPaymentMethods = async () => {
    setLoading(true);
    try {
      const data = await paymentService.listPaymentMethods();
      setPaymentMethods(data.payment_methods || []);
    } catch (error: any) {
      const errorMessage = error.response?.data?.error?.message || error.message || "Failed to load payment methods";
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleFormSuccess = () => {
    setShowForm(false);
    fetchPaymentMethods();
  };

  const handleFormCancel = () => {
    setShowForm(false);
  };

  if (!canManagePaymentMethods) {
    return (
      <DashboardLayout>
        <div className="p-4 sm:p-6 lg:p-8">
          <Card>
            <CardContent className="py-8 text-center">
              <CreditCard className="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
              <p className="text-muted-foreground">You don't have permission to manage payment methods.</p>
            </CardContent>
          </Card>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="p-4 sm:p-6 lg:p-8">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
          <div>
            <h1 className="text-xl sm:text-2xl lg:text-3xl font-bold">Payment Methods</h1>
            <p className="text-muted-foreground mt-1 text-sm sm:text-base">
              Manage your organization's payment methods
            </p>
          </div>
          {!showForm && (
            <Button onClick={() => setShowForm(true)} className="gap-2">
              <Plus className="w-4 h-4" />
              Add Payment Method
            </Button>
          )}
        </div>

        {showForm ? (
          <div className="mb-6">
            <PaymentMethodForm onSuccess={handleFormSuccess} onCancel={handleFormCancel} />
          </div>
        ) : null}

        {loading ? (
          <Card>
            <CardContent className="py-8 text-center">
              <p className="text-muted-foreground">Loading payment methods...</p>
            </CardContent>
          </Card>
        ) : (
          <PaymentMethodList paymentMethods={paymentMethods} onUpdate={fetchPaymentMethods} />
        )}
      </div>
    </DashboardLayout>
  );
};

export default PaymentMethods;

