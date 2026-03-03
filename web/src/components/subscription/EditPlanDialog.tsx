import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { adminPlanService, CreatePlanRequest, SubscriptionPlan } from "@/services/adminPlanService";
import { Loader2 } from "lucide-react";

interface EditPlanDialogProps {
  plan: SubscriptionPlan;
  onPlanUpdated: () => void;
}

export function EditPlanDialog({ plan, onPlanUpdated }: EditPlanDialogProps) {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [priceUSD, setPriceUSD] = useState(0);
  const [priceIDR, setPriceIDR] = useState(0);
  const [formData, setFormData] = useState<CreatePlanRequest>({
    name: plan.name,
    price: plan.price,
    currency: plan.currency,
    is_active: plan.is_active,
    limits: plan.limits || {
      patient_limit: 10,
      clinician_limit: 1,
    }
  });

  // Initialize prices when dialog opens or plan changes
  useEffect(() => {
    if (plan.prices) {
      setPriceUSD((plan.prices.USD || 0) / 100);
      setPriceIDR((plan.prices.IDR || 0) / 100);
    } else {
        // Fallback to base price if prices json is missing
        if (plan.currency === 'USD') setPriceUSD(plan.price / 100);
        if (plan.currency === 'IDR') setPriceIDR(plan.price / 100);
    }
  }, [plan]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      // Convert to subunits
      const prices = {
          USD: Number(priceUSD) * 100, // Cents
          IDR: Number(priceIDR) * 100  // Subunits for IDR format
      };

      const payload = {
          ...formData,
          price: prices.USD, // Default fallback price
          currency: "USD", // Default fallback currency
          prices: prices
      };
      
      await adminPlanService.updatePlan(plan.id, payload);
      setOpen(false);
      onPlanUpdated();
    } catch (error) {
      console.error("Failed to create plan", error);
      // TODO: Show toast error
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" className="w-full relative">
          Edit Plan
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Edit Subscription Plan</DialogTitle>
          <DialogDescription>
            Update the plan pricing or limits. Click save when you're done.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="name" className="text-right">
                Name
              </Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                className="col-span-3"
                required
              />
            </div>
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="priceUSD" className="text-right">
                Price (USD $)
              </Label>
              <Input
                id="priceUSD"
                type="number"
                min="0"
                step="0.01"
                value={priceUSD}
                onChange={(e) => setPriceUSD(Number(e.target.value))}
                className="col-span-3"
                required
              />
            </div>
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="priceIDR" className="text-right">
                Price (IDR Rp)
              </Label>
              <Input
                id="priceIDR"
                type="number"
                min="0"
                step="100"
                value={priceIDR}
                onChange={(e) => setPriceIDR(Number(e.target.value))}
                className="col-span-3"
                required
              />
            </div>
              <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="patient_limit" className="text-right">
                Patients
              </Label>
              <div className="col-span-3 flex items-center gap-2">
                 <Input
                    id="patient_limit"
                    type="number"
                    value={formData.limits.patient_limit === -1 ? "" : formData.limits.patient_limit}
                    onChange={(e) => setFormData({ 
                        ...formData, 
                        limits: { ...formData.limits, patient_limit: Number(e.target.value) } 
                    })}
                    disabled={formData.limits.patient_limit === -1}
                    className="flex-1"
                    placeholder="Limit"
                  />
                  <div className="flex items-center space-x-2">
                    <Checkbox 
                        id="unlimited_patients" 
                        checked={formData.limits.patient_limit === -1}
                        onCheckedChange={(checked) => {
                             setFormData({ 
                                ...formData, 
                                limits: { ...formData.limits, patient_limit: checked ? -1 : 10 } 
                            })
                        }}
                    />
                    <label
                        htmlFor="unlimited_patients"
                        className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                    >
                        Unlimited
                    </label>
                  </div>
              </div>
            </div>
             <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="clinician_limit" className="text-right">
                Team
              </Label>
              <div className="col-span-3 flex items-center gap-2">
                <Input
                    id="clinician_limit"
                    type="number"
                    value={formData.limits.clinician_limit === -1 ? "" : formData.limits.clinician_limit}
                    onChange={(e) => setFormData({ 
                        ...formData, 
                        limits: { ...formData.limits, clinician_limit: Number(e.target.value) } 
                    })}
                    disabled={formData.limits.clinician_limit === -1}
                    className="flex-1"
                    placeholder="Limit"
                />
                 <div className="flex items-center space-x-2">
                    <Checkbox 
                        id="unlimited_clinicians" 
                        checked={formData.limits.clinician_limit === -1}
                        onCheckedChange={(checked) => {
                             setFormData({ 
                                ...formData, 
                                limits: { ...formData.limits, clinician_limit: checked ? -1 : 1 } 
                            })
                        }}
                    />
                    <label
                        htmlFor="unlimited_clinicians"
                        className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                    >
                        Unlimited
                    </label>
                  </div>
              </div>
            </div>
             <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="is_active" className="text-right">
                Active
              </Label>
              <Checkbox 
                id="is_active" 
                checked={formData.is_active}
                onCheckedChange={(checked) => setFormData({...formData, is_active: checked as boolean})}
              />
            </div>
          </div>
          <DialogFooter>
            <Button type="submit" disabled={loading}>
              {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Save changes
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
