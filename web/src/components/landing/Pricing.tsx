import { useState, useEffect } from "react";
import { Check, ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Link } from "react-router-dom";
import { publicPlanService } from "@/services/publicPlanService";
import { SubscriptionPlan } from "@/services/adminPlanService";
import { useTranslation } from "react-i18next";

const Pricing = () => {
  const { i18n } = useTranslation();
  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
  const [loading, setLoading] = useState(true);
  const [currency, setCurrency] = useState<string>("USD");

  useEffect(() => {
    // Default to IDR if language is Indonesian
    if (i18n.language && i18n.language.startsWith('id')) {
        setCurrency("IDR");
    }
  }, [i18n.language]);

  useEffect(() => {
    const fetchPlans = async () => {
      try {
        const data = await publicPlanService.listActivePlans();
        setPlans(data);
      } catch (error) {
        console.error("Failed to load plans", error);
      } finally {
        setLoading(false);
      }
    };
    fetchPlans();
  }, []);

  // Helper to get display price based on selected currency
  const getDisplayPrice = (plan: SubscriptionPlan) => {
      if (!plan.prices || plan.prices.length === 0) {
           // Fallback for legacy simple price if exists and matches currency (unlikely but safe)
           if ((plan as any).price !== undefined && (plan as any).currency === currency) {
               return { price: (plan as any).price, currency: currency };
           }
           // Fallback to USD 0 if really nothing
           return { price: 0, currency: currency };
      }
      
      const priceObj = plan.prices.find(p => p.currency === currency);
      // If price not found for selected currency, fallback to USD or first available
      // But ideally we return 0/NA for that currency to prompt "Contact Sales" or similar? 
      // For now fallback to USD if IDR missing, or first.
      const fallback = plan.prices.find(p => p.currency === 'USD') || plan.prices[0];
      
      const finalPrice = priceObj || fallback;
      
      return { 
          price: finalPrice?.price ?? 0, 
          currency: finalPrice?.currency || 'USD' 
      };
  }

  // Helper to get plan features from limits
  const getPlanFeatures = (plan: SubscriptionPlan) => {
      const displayPrice = getDisplayPrice(plan);
      return [
          { name: "Patients", value: plan.limits?.patient_limit === -1 ? "Unlimited" : plan.limits?.patient_limit },
          { name: "Team Members", value: plan.limits?.clinician_limit === -1 ? "Unlimited" : plan.limits?.clinician_limit },
          { name: "Clinical Notes", value: "Unlimited" },
          { name: "HIPAA Compliance", value: "Included" },
          { name: "Support", value: displayPrice.price > 0 ? "Priority" : "Community" },
      ];
  };

  return (
    <section id="pricing" className="py-24 bg-slate-50 relative overflow-hidden">
        {/* Background Decoration */}
        <div className="absolute top-0 left-0 w-full h-full overflow-hidden pointer-events-none">
            <div className="absolute top-[-10%] right-[-5%] w-[500px] h-[500px] bg-primary/5 rounded-full blur-3xl opacity-50"></div>
            <div className="absolute bottom-[-10%] left-[-5%] w-[500px] h-[500px] bg-blue-500/5 rounded-full blur-3xl opacity-50"></div>
        </div>

      <div className="container px-4 relative z-10">
        <div className="text-center mb-10">
          <h2 className="text-4xl md:text-5xl font-extrabold font-heading mb-6 text-foreground tracking-tight">
            Simple, Transparent Pricing
          </h2>
          <p className="text-xl text-muted-foreground max-w-2xl mx-auto font-light leading-relaxed mb-8">
            Start free, upgrade as you grow. No hidden fees or long-term contracts.
          </p>

          {/* Currency Toggle */}
          <div className="inline-flex items-center p-1 bg-white border border-border rounded-full shadow-sm">
            <button 
                onClick={() => setCurrency("IDR")}
                className={`px-4 py-2 rounded-full text-sm font-bold transition-all ${currency === "IDR" ? "bg-primary text-primary-foreground shadow-md" : "text-muted-foreground hover:text-foreground"}`}
            >
                IDR (Rp)
            </button>
            <button 
                onClick={() => setCurrency("USD")}
                className={`px-4 py-2 rounded-full text-sm font-bold transition-all ${currency === "USD" ? "bg-primary text-primary-foreground shadow-md" : "text-muted-foreground hover:text-foreground"}`}
            >
                USD ($)
            </button>
          </div>
        </div>
        
        {loading ? (
           <div className="flex justify-center py-20">
             <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
           </div>
        ) : (
            <div className={`grid grid-cols-1 gap-8 max-w-5xl mx-auto ${plans.length > 1 ? 'md:grid-cols-' + Math.min(plans.length, 3) : 'md:grid-cols-1'}`}>
            {plans.map((plan) => {
                const displayPrice = getDisplayPrice(plan);
                // If the selected currency is NOT available for this plan (e.g. only has USD), and we are in IDR mode, 
                // displayPrice might have returned a fallback USD price. 
                // We should probably check if the returned currency matches the selected one for UI consistency,
                // or just display whatever currency code we got.
                
                const isPlanFree = displayPrice.price === 0;
                const isHighlighted = !isPlanFree;

                return (
                <div 
                    key={plan.id} 
                    className={`
                    relative rounded-3xl p-8 transition-all duration-300 flex flex-col
                    ${isHighlighted 
                        ? "bg-white border-2 border-primary/20 shadow-2xl shadow-primary/10 scale-105 z-10" 
                        : "bg-white/80 border border-border/50 hover:border-primary/20 hover:shadow-xl hover:-translate-y-2 backdrop-blur-sm"
                    }
                    `}
                >
                    {isHighlighted && (
                    <div className="absolute -top-5 left-1/2 -translate-x-1/2">
                        <span className="bg-primary text-primary-foreground text-sm font-bold px-4 py-1.5 rounded-full uppercase tracking-wider shadow-lg shadow-primary/20">
                        Most Popular
                        </span>
                    </div>
                    )}

                    <div className="mb-8">
                    <h3 className="font-heading text-2xl font-bold text-foreground mb-2">{plan.name}</h3>
                    <p className="text-muted-foreground">
                        {isPlanFree ? "Perfect for independent clinicians" : "For growing practices and teams"}
                    </p>
                    </div>

                    <div className="mb-8 flex items-baseline gap-1">
                    <span className="text-5xl font-extrabold font-heading text-foreground tracking-tight">
                        {(displayPrice.price / 100).toLocaleString(i18n.language || 'en-US', { style: 'currency', currency: displayPrice.currency, minimumFractionDigits: 0 })}
                    </span>
                    <span className="text-lg font-medium text-muted-foreground">/mo</span>
                    </div>

                    <ul className="space-y-4 mb-8 flex-1">
                        {getPlanFeatures(plan).map((feature, i) => (
                        <li key={i} className="flex items-start gap-3 group">
                            <div className={`
                                mt-0.5 rounded-full p-0.5 flex-shrink-0
                                ${isHighlighted ? "bg-primary/10 text-primary" : "bg-gray-100 text-gray-400 group-hover:text-primary transition-colors"}
                            `}>
                                <Check className="w-4 h-4" />
                            </div>
                            <span className="text-foreground/80">
                                <span className="font-semibold text-foreground">{feature.value}</span> {feature.name}
                            </span>
                        </li>
                        ))}
                    </ul>

                    <Link to={`/auth?signup=true${!isPlanFree ? '&plan=paid' : ''}`} className="w-full">
                        <Button 
                            className={`
                                w-full h-12 text-base font-bold rounded-xl transition-all
                                ${isHighlighted 
                                    ? "bg-primary hover:bg-primary/90 shadow-lg shadow-primary/25 hover:shadow-xl hover:shadow-primary/30" 
                                    : "bg-gray-900 text-white hover:bg-gray-800"
                                }
                            `}
                        >
                            {isPlanFree ? "Start for Free" : "Get Started Now"}
                            <ArrowRight className="ml-2 w-4 h-4" />
                        </Button>
                    </Link>
                </div>
                );
            })}
            </div>
        )}
      </div>
    </section>
  );
};

export default Pricing;
