import { Users, Calendar, FileText, DollarSign } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { isSameDay, parseISO, subDays, isWithinInterval, startOfMonth } from "date-fns";
import { usePatients, useAppointments, useClinicalNotes, useInvoices } from "@/hooks/useDashboardQueries";
import { useEffect, useState } from "react";
import { organizationService, Organization } from "@/services/organizationService";

interface Stats {
  activePatients: number;
  todaySessions: number;
  weeklyNotes: number;
  monthlyRevenue: number;
}

const StatsCards = () => {
  const { data: patients, isLoading: patientsLoading } = usePatients();
  const { data: appointments, isLoading: appointmentsLoading } = useAppointments();
  const { data: notes, isLoading: notesLoading } = useClinicalNotes();
  const { data: invoices, isLoading: invoicesLoading } = useInvoices();
  const [organization, setOrganization] = useState<Organization | null>(null);

  useEffect(() => {
    organizationService.getMyOrganization().then(setOrganization).catch(console.error);
  }, []);

  const loading = patientsLoading || appointmentsLoading || notesLoading || invoicesLoading;

  // Calculate stats
  // 1. Active Patients
  const activePatients = (patients || []).filter(p => p.status === 'active').length;

  // 2. Today's Sessions
  const today = new Date();
  const todaySessions = (appointments || []).filter(apt => {
    try {
      if (!apt.start_time) return false;
      return isSameDay(parseISO(apt.start_time), today);
    } catch { return false; }
  }).length;

  // 3. Weekly Notes (Last 7 days)
  const weekStart = subDays(today, 7);
  const weeklyNotes = (notes || []).filter(note => {
    try {
      if (!note.created_at) return false;
      return isWithinInterval(parseISO(note.created_at), { start: weekStart, end: today });
    } catch { return false; }
  }).length;

  // 4. Monthly Revenue (Paid invoices this month)
  const monthStart = startOfMonth(today);
  const monthlyRevenue = (invoices || []).reduce((sum, inv) => {
    if (inv.status === 'paid' && inv.paid_at) {
      try {
        const paidDate = parseISO(inv.paid_at);
        if (isWithinInterval(paidDate, { start: monthStart, end: today })) {
          return sum + inv.amount_cents;
        }
      } catch { return sum; }
    }
    return sum;
  }, 0) / 100;

  const statsConfig = [
    {
      label: "Active Patients",
      value: activePatients.toString(),
      icon: Users,
      color: "text-primary",
      bg: "bg-primary/10",
    },
    {
      label: "Today's Sessions",
      value: todaySessions.toString(),
      icon: Calendar,
      color: "text-chart-2",
      bg: "bg-chart-2/10",
    },
    {
      label: "Notes This Week",
      value: weeklyNotes.toString(),
      icon: FileText,
      color: "text-chart-3",
      bg: "bg-chart-3/10",
    },
    {
      label: "Revenue (MTD)",
      value: new Intl.NumberFormat(organization?.locale || "en-US", {
        style: "currency",
        currency: organization?.currency || "USD",
      }).format(monthlyRevenue),
      icon: DollarSign,
      color: "text-chart-1",
      bg: "bg-chart-1/10",
    },
  ];

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      {statsConfig.map((stat, index) => (
        <Card key={index} className="hover:shadow-md transition-shadow">
          <CardContent className="p-6">
            <div className="flex items-start justify-between">
              <div>
                <p className="text-sm text-muted-foreground">{stat.label}</p>
                <p className="text-3xl font-bold mt-1">
                  {loading ? "..." : stat.value}
                </p>
              </div>
              <div className={`p-3 rounded-lg ${stat.bg}`}>
                <stat.icon className={`w-5 h-5 ${stat.color}`} />
              </div>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
};

export default StatsCards;
