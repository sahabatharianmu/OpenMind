import { Button } from "@/components/ui/button";
import { Shield, Lock, ArrowRight } from "lucide-react";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";

const Hero = () => {
  const { t } = useTranslation('landing');

  return (
    <section className="relative min-h-[90vh] flex items-center justify-center overflow-hidden">
      {/* Background gradient */}
      <div className="absolute inset-0 bg-gradient-to-br from-accent via-background to-background" />
      
      {/* Floating orbs */}
      <div className="absolute top-20 left-20 w-72 h-72 bg-primary/10 rounded-full blur-3xl animate-pulse" />
      <div className="absolute bottom-20 right-20 w-96 h-96 bg-primary/5 rounded-full blur-3xl animate-pulse delay-1000" />
      
      <div className="container relative z-10 px-4 py-20">
        <div className="max-w-4xl mx-auto text-center">
          {/* Badge */}
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-accent border border-primary/20 mb-8">
            <Shield className="w-4 h-4 text-primary" />
            <span className="text-sm font-medium text-accent-foreground">{t('hero.badge')}</span>
          </div>
          
          {/* Headline */}
          <h1 className="text-4xl md:text-6xl lg:text-7xl font-bold tracking-tight mb-6">
            {t('hero.headline1')}{" "}
            <span className="text-primary">{t('hero.headline2')}</span>
            <br />
            <span className="text-muted-foreground">{t('hero.headline3')}</span>
          </h1>
          
          {/* Subheadline */}
          <p className="text-lg md:text-xl text-muted-foreground max-w-2xl mx-auto mb-6">
            {t('hero.subheadline')}
            <span className="block mt-2 font-semibold text-foreground">
              {t('hero.freeForever')}
            </span>
          </p>
          
          {/* CTAs */}
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4 mb-6">
            <Link to="/auth?signup=true">
              <Button size="lg" className="gap-2 text-lg px-8 py-6">
                {t('hero.ctaStart')}
                <ArrowRight className="w-5 h-5" />
              </Button>
            </Link>
            <a href="#pricing">
              <Button variant="outline" size="lg" className="gap-2 text-lg px-8 py-6">
                {t('hero.ctaPricing')}
              </Button>
            </a>
            <Button variant="ghost" size="lg" className="gap-2 text-lg px-8 py-6">
              <Lock className="w-5 h-5" />
              {t('hero.ctaSecurity')}
            </Button>
          </div>

          {/* Free Tier Benefits */}
          <div className="flex flex-wrap items-center justify-center gap-6 text-sm text-muted-foreground mb-8">
            <div className="flex items-center gap-2">
              <span className="text-primary font-semibold">✓</span>
              <span>{t('hero.benefit1')}</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-primary font-semibold">✓</span>
              <span>{t('hero.benefit2')}</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-primary font-semibold">✓</span>
              <span>{t('hero.benefit3')}</span>
            </div>
          </div>
          
          {/* Trust indicators */}
          <div className="mt-16 flex flex-wrap items-center justify-center gap-8 text-sm text-muted-foreground">
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 rounded-full bg-green-500" />
              {t('trust.hipaa')}
            </div>
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 rounded-full bg-green-500" />
              {t('trust.gdpr')}
            </div>
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 rounded-full bg-green-500" />
              {t('trust.openSource')}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default Hero;
