import { Button } from "@/components/ui/button";
import { Menu, Globe } from "lucide-react";
import { Link } from "react-router-dom";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Sheet,
  SheetContent,
  SheetTrigger,
} from "@/components/ui/sheet";

const Navbar = () => {
  const [open, setOpen] = useState(false);
  const { t, i18n } = useTranslation('landing');
  const { t: tc } = useTranslation('common');

  const navLinks = [
    { label: t('hero.ctaPricing').replace('See ', '').replace('Lihat ', ''), href: "#features", key: "features" },
    { label: tc('nav.pricing'), href: "#pricing", key: "pricing" },
    { label: tc('nav.docs'), href: "#docs", key: "docs" },
    { label: tc('nav.github'), href: "https://github.com/closaf/practice", key: "github" },
  ];

  const toggleLang = () => {
    i18n.changeLanguage(i18n.language === 'en' ? 'id' : 'en');
  };

  return (
    <header className="sticky top-0 z-50 w-full border-b border-border/40 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-16 items-center justify-between px-4">
        {/* Logo */}
        <Link to="/" className="flex items-center gap-2">
          <img src="/logo_text_horizontal.svg" alt="Closaf" className="h-8" />
        </Link>

        {/* Desktop Nav */}
        <nav className="hidden md:flex items-center gap-6">
          {navLinks.map((link) => (
            <a
              key={link.key}
              href={link.href}
              className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
            >
              {link.label}
            </a>
          ))}
        </nav>

        {/* Desktop CTAs */}
        <div className="hidden md:flex items-center gap-3">
          <Button variant="ghost" size="sm" onClick={toggleLang} className="gap-1.5">
            <Globe className="w-4 h-4" />
            {i18n.language === 'en' ? 'ID' : 'EN'}
          </Button>
          <Link to="/auth">
            <Button variant="ghost">{tc('nav.signIn')}</Button>
          </Link>
          <Link to="/auth?signup=true">
            <Button>{tc('nav.startFree')}</Button>
          </Link>
        </div>

        {/* Mobile Menu */}
        <Sheet open={open} onOpenChange={setOpen}>
          <SheetTrigger asChild className="md:hidden">
            <Button variant="ghost" size="icon">
              <Menu className="w-5 h-5" />
            </Button>
          </SheetTrigger>
          <SheetContent side="right" className="w-[300px]">
            <nav className="flex flex-col gap-4 mt-8">
              {navLinks.map((link) => (
                <a
                  key={link.key}
                  href={link.href}
                  className="text-lg font-medium text-muted-foreground hover:text-foreground transition-colors"
                  onClick={() => setOpen(false)}
                >
                  {link.label}
                </a>
              ))}
              <Button variant="ghost" size="sm" onClick={toggleLang} className="gap-1.5 justify-start">
                <Globe className="w-4 h-4" />
                {tc(`language.${i18n.language === 'en' ? 'id' : 'en'}`)}
              </Button>
              <div className="flex flex-col gap-2 mt-4">
                <Link to="/auth" onClick={() => setOpen(false)}>
                  <Button variant="outline" className="w-full">{tc('nav.signIn')}</Button>
                </Link>
                <Link to="/auth?signup=true" onClick={() => setOpen(false)}>
                  <Button className="w-full">{tc('nav.startFree')}</Button>
                </Link>
              </div>
            </nav>
          </SheetContent>
        </Sheet>
      </div>
    </header>
  );
};

export default Navbar;
