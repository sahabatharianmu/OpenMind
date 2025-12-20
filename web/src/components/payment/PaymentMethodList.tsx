import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { CreditCard, Trash2, Star, StarOff, Loader2 } from "lucide-react";
import { paymentService, type PaymentMethod } from "@/services/paymentService";
import { useToast } from "@/hooks/use-toast";
import { format } from "date-fns";

interface PaymentMethodListProps {
  paymentMethods: PaymentMethod[];
  onUpdate: () => void;
}

const PaymentMethodList = ({ paymentMethods, onUpdate }: PaymentMethodListProps) => {
  const { toast } = useToast();
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [settingDefault, setSettingDefault] = useState<string | null>(null);

  const handleDelete = async (id: string) => {
    setDeletingId(id);
    try {
      await paymentService.deletePaymentMethod(id);
      toast({
        title: "Success",
        description: "Payment method deleted successfully",
      });
      onUpdate();
    } catch (error: any) {
      const errorMessage = error.response?.data?.error?.message || error.message || "Failed to delete payment method";
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      });
    } finally {
      setDeletingId(null);
      setShowDeleteDialog(false);
      setSelectedId(null);
    }
  };

  const handleSetDefault = async (id: string) => {
    setSettingDefault(id);
    try {
      await paymentService.setDefaultPaymentMethod(id);
      toast({
        title: "Success",
        description: "Default payment method updated",
      });
      onUpdate();
    } catch (error: any) {
      const errorMessage = error.response?.data?.error?.message || error.message || "Failed to set default payment method";
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      });
    } finally {
      setSettingDefault(null);
    }
  };

  const getBrandIcon = (brand: string) => {
    const brandLower = brand.toLowerCase();
    if (brandLower.includes("visa")) return "💳";
    if (brandLower.includes("mastercard")) return "💳";
    if (brandLower.includes("amex") || brandLower.includes("american")) return "💳";
    if (brandLower.includes("discover")) return "💳";
    return "💳";
  };

  const formatExpiry = (month: number, year: number) => {
    return `${String(month).padStart(2, "0")}/${String(year).slice(-2)}`;
  };

  if (paymentMethods.length === 0) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <CreditCard className="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
          <p className="text-muted-foreground">No payment methods added yet</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Card</TableHead>
                <TableHead>Brand</TableHead>
                <TableHead>Expires</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Added</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {paymentMethods.map((pm) => (
                <TableRow key={pm.id}>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <span className="text-xl">{getBrandIcon(pm.brand)}</span>
                      <span className="font-mono">•••• •••• •••• {pm.last4}</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" className="capitalize">
                      {pm.brand}
                    </Badge>
                  </TableCell>
                  <TableCell>{formatExpiry(pm.expiry_month, pm.expiry_year)}</TableCell>
                  <TableCell>
                    {pm.is_default ? (
                      <Badge className="bg-primary">
                        <Star className="w-3 h-3 mr-1" />
                        Default
                      </Badge>
                    ) : (
                      <Badge variant="secondary">Active</Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {format(new Date(pm.created_at), "MMM d, yyyy")}
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-2">
                      {!pm.is_default && (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleSetDefault(pm.id)}
                          disabled={settingDefault === pm.id}
                        >
                        {settingDefault === pm.id ? (
                          <Loader2 className="w-4 h-4 animate-spin mr-1" />
                        ) : (
                          <StarOff className="w-4 h-4 mr-1" />
                        )}
                          Set Default
                        </Button>
                      )}
                      <Button
                        variant="destructive"
                        size="sm"
                        onClick={() => {
                          setSelectedId(pm.id);
                          setShowDeleteDialog(true);
                        }}
                        disabled={deletingId === pm.id}
                      >
                        {deletingId === pm.id ? (
                          <Loader2 className="w-4 h-4 animate-spin" />
                        ) : (
                          <Trash2 className="w-4 h-4" />
                        )}
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Payment Method</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete this payment method? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => selectedId && handleDelete(selectedId)}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
};

export default PaymentMethodList;

