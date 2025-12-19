import { ReactNode, useState, useEffect } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { 
  Brain, 
  Users, 
  Calendar, 
  FileText, 
  Receipt, 
  Settings,
  LogOut,
  ChevronLeft,
  ChevronRight,
  LayoutDashboard,
  Shield,
  Menu
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { useAuth } from "@/contexts/AuthContext";
import { useIsMobile } from "@/hooks/use-mobile";
import {
  Sheet,
  SheetContent,
  SheetTrigger,
} from "@/components/ui/sheet";

interface DashboardLayoutProps {
  children: ReactNode;
}

const navItems = [
  { icon: LayoutDashboard, label: "Dashboard", path: "/dashboard" },
  { icon: Users, label: "Patients", path: "/dashboard/patients" },
  { icon: Calendar, label: "Appointments", path: "/dashboard/appointments" },
  { icon: FileText, label: "Notes", path: "/dashboard/notes" },
  { icon: Receipt, label: "Billing", path: "/dashboard/billing" },
  { icon: Shield, label: "Audit Logs", path: "/dashboard/audit-logs" },
  { icon: Settings, label: "Settings", path: "/dashboard/settings" },
];

const DashboardLayout = ({ children }: DashboardLayoutProps) => {
  const [collapsed, setCollapsed] = useState(false);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const location = useLocation();
  const navigate = useNavigate();
  const { user, signOut, loading } = useAuth();
  const isMobile = useIsMobile();

  useEffect(() => {
    if (!loading && !user) {
      navigate("/auth");
    }
  }, [user, loading, navigate]);

  const handleSignOut = async () => {
    await signOut();
    navigate("/");
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="animate-pulse">Loading...</div>
      </div>
    );
  }

  if (!user) {
    return null;
  }

  const initials = user?.full_name
    ?.split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase() || user?.email?.[0]?.toUpperCase() || "U";

  const SidebarContent = ({ onNavigate }: { onNavigate?: () => void }) => (
    <div className="flex flex-col h-full">
      {/* Logo */}
      <div className="h-16 flex items-center justify-between px-4 border-b border-sidebar-border flex-shrink-0">
        <Link to="/" className="flex items-center gap-2" onClick={onNavigate}>
          <div className="w-8 h-8 rounded-lg bg-sidebar-primary flex items-center justify-center flex-shrink-0">
            <img src="/SahariIcon.svg" alt="OpenMind" className="w-5 h-5" />
          </div>
          {(!collapsed || isMobile) && <span className="font-bold text-sidebar-foreground">OpenMind</span>}
        </Link>
        {!isMobile && (
          <Button 
            variant="ghost" 
            size="icon" 
            className="h-10 w-10 min-h-[44px] min-w-[44px]"
            onClick={() => setCollapsed(!collapsed)}
          >
            {collapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
          </Button>
        )}
      </div>

      {/* Nav Items */}
      <nav className="p-2 space-y-1 flex-1 overflow-y-auto">
        {navItems.map((item) => {
          const isActive = location.pathname === item.path || 
            (item.path !== "/dashboard" && location.pathname.startsWith(item.path));
          
          return (
            <Link key={item.path} to={item.path} onClick={onNavigate}>
              <Button
                variant={isActive ? "secondary" : "ghost"}
                className={cn(
                  "w-full justify-start gap-3 h-11 min-h-[44px]",
                  collapsed && !isMobile && "justify-center px-2"
                )}
              >
                <item.icon className="w-5 h-5 flex-shrink-0" />
                {(!collapsed || isMobile) && <span>{item.label}</span>}
              </Button>
            </Link>
          );
        })}
      </nav>

      {/* Bottom section */}
      <div className="p-4 border-t border-sidebar-border flex-shrink-0">
        <div className={cn(
          "flex items-center gap-3",
          collapsed && !isMobile && "justify-center"
        )}>
          <Avatar className="h-10 w-10 min-h-[44px] min-w-[44px]">
            <AvatarFallback className="bg-primary text-primary-foreground text-xs">
              {initials}
            </AvatarFallback>
          </Avatar>
          {(!collapsed || isMobile) && (
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium truncate">{user?.full_name || "User"}</p>
              <p className="text-xs text-muted-foreground truncate">{user?.email}</p>
            </div>
          )}
          {(!collapsed || isMobile) && (
            <Button 
              variant="ghost" 
              size="icon" 
              className="h-10 w-10 min-h-[44px] min-w-[44px]" 
              onClick={handleSignOut}
            >
              <LogOut className="w-4 h-4" />
            </Button>
          )}
        </div>
      </div>
    </div>
  );

  return (
    <div className="min-h-screen bg-background flex">
      {/* Desktop Sidebar */}
      {!isMobile && (
        <aside 
          className={cn(
            "fixed left-0 top-0 h-full bg-sidebar border-r border-sidebar-border transition-all duration-300 z-40",
            collapsed ? "w-16" : "w-64"
          )}
        >
          <SidebarContent />
        </aside>
      )}

      {/* Mobile Header with Menu Button */}
      {isMobile && (
        <header className="fixed top-0 left-0 right-0 h-16 bg-sidebar border-b border-sidebar-border z-50 flex items-center justify-between px-4">
          <Link to="/" className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-sidebar-primary flex items-center justify-center">
              <img src="/SahariIcon.svg" alt="OpenMind" className="w-5 h-5" />
            </div>
            <span className="font-bold text-sidebar-foreground">OpenMind</span>
          </Link>
          <Sheet open={mobileMenuOpen} onOpenChange={setMobileMenuOpen}>
            <SheetTrigger asChild>
              <Button variant="ghost" size="icon" className="h-10 w-10 min-h-[44px] min-w-[44px]">
                <Menu className="w-5 h-5" />
              </Button>
            </SheetTrigger>
            <SheetContent side="left" className="w-64 p-0 bg-sidebar">
              <div className="relative h-full flex flex-col">
                <SidebarContent onNavigate={() => setMobileMenuOpen(false)} />
              </div>
            </SheetContent>
          </Sheet>
        </header>
      )}

      {/* Main Content */}
      <main className={cn(
        "flex-1 transition-all duration-300",
        !isMobile && (collapsed ? "ml-16" : "ml-64"),
        isMobile && "mt-16"
      )}>
        {children}
      </main>
    </div>
  );
};

export default DashboardLayout;
