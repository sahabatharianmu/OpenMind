import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";

const Footer = () => {
  const { t } = useTranslation('landing');

  return (
    <footer className="border-t border-border bg-card">
      <div className="container px-4 py-12">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
          {/* Brand */}
          <div className="col-span-2 md:col-span-1">
            <Link to="/" className="flex items-center gap-2 mb-4">
              <img src="/logo_text_horizontal.svg" alt="Closaf" className="h-8" />
            </Link>
            <p className="text-sm text-muted-foreground">
              {t('footer.tagline')}
            </p>
          </div>

          {/* Product */}
          <div>
            <h4 className="font-semibold mb-4">{t('footer.product')}</h4>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li><a href="#" className="hover:text-foreground transition-colors">{t('footer.features')}</a></li>
              <li><a href="#pricing" className="hover:text-foreground transition-colors">{t('pricing.title')}</a></li>
              <li><a href="#" className="hover:text-foreground transition-colors">{t('footer.security')}</a></li>
              <li><a href="#" className="hover:text-foreground transition-colors">{t('footer.roadmap')}</a></li>
            </ul>
          </div>

          {/* Resources */}
          <div>
            <h4 className="font-semibold mb-4">{t('footer.resources')}</h4>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li><a href="#" className="hover:text-foreground transition-colors">{t('footer.documentation')}</a></li>
              <li><a href="#" className="hover:text-foreground transition-colors">{t('footer.apiReference')}</a></li>
              <li><a href="#" className="hover:text-foreground transition-colors">{t('footer.selfHosting')}</a></li>
              <li><a href="#" className="hover:text-foreground transition-colors">GitHub</a></li>
            </ul>
          </div>

          {/* Legal */}
          <div>
            <h4 className="font-semibold mb-4">{t('footer.legal')}</h4>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li><a href="#" className="hover:text-foreground transition-colors">{t('footer.privacy')}</a></li>
              <li><a href="#" className="hover:text-foreground transition-colors">{t('footer.terms')}</a></li>
              <li><a href="#" className="hover:text-foreground transition-colors">{t('trust.hipaa')}</a></li>
              <li><a href="#" className="hover:text-foreground transition-colors">License (AGPLv3)</a></li>
            </ul>
          </div>
        </div>

        <div className="border-t border-border mt-12 pt-8 text-center text-sm text-muted-foreground">
          <p>{t('footer.copyright', { year: new Date().getFullYear() })}</p>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
