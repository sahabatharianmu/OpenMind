import { useState } from "react";
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
import { adminPlanService, CreatePlanRequest } from "@/services/adminPlanService";
import { Plus, Loader2 } from "lucide-react";

interface CreatePlanDialogProps {
  onPlanCreated: () => void;
}

export function CreatePlanDialog({ onPlanCreated }: CreatePlanDialogProps) {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState<CreatePlanRequest>({
    name: "",
    prices: [{ currency: "USD", price: 0 }],
    is_active: true,
    limits: {
      patient_limit: 10,
      clinician_limit: 1,
    }
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      // Convert price to cents if input is in dollars (simplified logic)
      const payload = {
          ...formData,
          prices: formData.prices.map(p => ({
              ...p,
              price: Number(p.price) * 100
          }))
      };
      
      await adminPlanService.createPlan(payload);
      setOpen(false);
      onPlanCreated();
      // Reset form
      setFormData({
        name: "",
        prices: [{ currency: "USD", price: 0 }],
        is_active: true,
        limits: {
            patient_limit: 10,
            clinician_limit: 1,
        }
      });
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
        <Button>
          <Plus className="mr-2 h-4 w-4" /> Create Plan
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Create Subscription Plan</DialogTitle>
          <DialogDescription>
            Add a new plan to your offering. Click save when you're done.
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
            <div className="grid grid-cols-4 items-start gap-4">
              <Label className="text-right pt-2">
                Prices
              </Label>
              <div className="col-span-3 space-y-2">
                {formData.prices.map((price, index) => (
                    <div key={index} className="flex gap-2">
                        <Input
                            placeholder="Cur"
                            value={price.currency}
                            onChange={(e) => {
                                const newPrices = [...formData.prices];
                                newPrices[index].currency = e.target.value.toUpperCase();
                                setFormData({ ...formData, prices: newPrices });
                            }}
                            className="w-20"
                            maxLength={3}
                            required
                        />
                         <Input
                            type="number"
                            min="0"
                            step="0.01"
                            value={price.price}
                            onChange={(e) => {
                                const newPrices = [...formData.prices];
                                newPrices[index].price = Number(e.target.value);
                                setFormData({ ...formData, prices: newPrices });
                            }}
                            className="flex-1"
                            required
                         />
                         {index > 0 && (
                             <Button 
                                type="button" 
                                variant="destructive" 
                                size="icon"
                                onClick={() => {
                                    const newPrices = formData.prices.filter((_, i) => i !== index);
                                    setFormData({ ...formData, prices: newPrices });
                                }}
                             >
                                 <span className="sr-only">Remove</span>
                                 &times;
                             </Button>
                         )}
                    </div>
                ))}
                <Button 
                    type="button" 
                    variant="outline" 
                    size="sm"
                    onClick={() => {
                        setFormData({ 
                            ...formData, 
                            prices: [...formData.prices, { currency: "", price: 0 }] 
                        });
                    }}
                >
                    Add Price
                </Button>
              </div>
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
