import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { AuthProvider } from "@/contexts/AuthContext";
import ErrorBoundary from "@/components/ErrorBoundary";
import { ProtectedRoute, AdminRoute } from "@/components/guards/RouteGuards";
import Index from "./pages/Index";
import Auth from "./pages/Auth";
import Dashboard from "./pages/Dashboard";
import Patients from "./pages/Patients";
import Appointments from "./pages/Appointments";
import Notes from "./pages/Notes";
import NoteEditor from "./pages/NoteEditor";
import Billing from "./pages/Billing";
import Settings from "./pages/Settings";
import PatientProfile from "./pages/PatientProfile";
import AuditLogs from "./pages/AuditLogs";
import Teams from "./pages/Teams";
import AcceptInvitation from "./pages/AcceptInvitation";
import Onboarding from "./pages/Onboarding";
import Pricing from "./pages/Pricing";
import PaymentMethods from "./pages/PaymentMethods";
import NotFound from "./pages/NotFound";
import AdminDashboard from "./pages/admin/AdminDashboard";
import SubscriptionPlans from "./pages/admin/SubscriptionPlans";
import Tenants from "./pages/admin/Tenants";

const queryClient = new QueryClient();

const App = () => (
  <QueryClientProvider client={queryClient}>
    <AuthProvider>
      <TooltipProvider>
        <ErrorBoundary>
          <Toaster />
          <Sonner />
          <BrowserRouter>
            <Routes>
              {/* Public routes */}
              <Route path="/" element={<Index />} />
              <Route path="/auth" element={<Auth />} />
              <Route path="/pricing" element={<Pricing />} />
              <Route path="/accept-invitation" element={<AcceptInvitation />} />

              {/* Protected routes — require authentication */}
              <Route path="/onboarding" element={<ProtectedRoute><Onboarding /></ProtectedRoute>} />
              <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
              <Route path="/dashboard/patients" element={<ProtectedRoute><Patients /></ProtectedRoute>} />
              <Route path="/dashboard/patients/:id" element={<ProtectedRoute><PatientProfile /></ProtectedRoute>} />
              <Route path="/dashboard/appointments" element={<ProtectedRoute><Appointments /></ProtectedRoute>} />
              <Route path="/dashboard/notes" element={<ProtectedRoute><Notes /></ProtectedRoute>} />
              <Route path="/dashboard/notes/:id" element={<ProtectedRoute><NoteEditor /></ProtectedRoute>} />
              <Route path="/dashboard/billing" element={<ProtectedRoute><Billing /></ProtectedRoute>} />
              <Route path="/dashboard/settings" element={<ProtectedRoute><Settings /></ProtectedRoute>} />
              <Route path="/dashboard/audit-logs" element={<ProtectedRoute><AuditLogs /></ProtectedRoute>} />
              <Route path="/dashboard/teams" element={<ProtectedRoute><Teams /></ProtectedRoute>} />
              <Route path="/dashboard/payment-methods" element={<ProtectedRoute><PaymentMethods /></ProtectedRoute>} />

              {/* Admin routes — require authentication + system_role=admin */}
              <Route path="/admin" element={<AdminRoute><AdminDashboard /></AdminRoute>} />
              <Route path="/admin/plans" element={<AdminRoute><SubscriptionPlans /></AdminRoute>} />
              <Route path="/admin/tenants" element={<AdminRoute><Tenants /></AdminRoute>} />

              {/* Catch-all */}
              <Route path="*" element={<NotFound />} />
            </Routes>
          </BrowserRouter>
        </ErrorBoundary>
      </TooltipProvider>
    </AuthProvider>
  </QueryClientProvider>
);

export default App;
