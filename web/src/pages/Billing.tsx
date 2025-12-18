import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import DashboardLayout from "@/components/dashboard/DashboardLayout";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { 
  Plus, 
  Search, 
  Receipt, 
  MoreHorizontal,
  DollarSign,
  Clock,
  CheckCircle,
  XCircle,
  FileText,
  Download,
  FileDown
} from "lucide-react";
import invoiceService from "@/services/invoiceService";
import patientService from "@/services/patientService";
import appointmentService from "@/services/appointmentService";
import { useAuth } from "@/contexts/AuthContext";
import { useToast } from "@/hooks/use-toast";
import { format } from "date-fns";
import { InvoiceDetailDialog } from "@/components/billing/InvoiceDetailDialog";
import { RevenueChart } from "@/components/billing/RevenueChart";
import { exportInvoicesToCSV } from "@/components/billing/exportInvoices";
import { Invoice, Patient, Appointment } from "@/types";

// Extended types for UI
interface UIInvoice extends Invoice {
  patients: { // Matching the structure expected by children components like RevenueChart
    first_name: string;
    last_name: string;
  };
  appointments?: {
    start_time: string;
    appointment_type: string;
  } | null;
}

const Billing = () => {
  const navigate = useNavigate();
  const { user } = useAuth();
  const { toast } = useToast();
  const [invoices, setInvoices] = useState<UIInvoice[]>([]);
  const [patients, setPatients] = useState<Patient[]>([]);
  const [appointments, setAppointments] = useState<Appointment[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [selectedInvoice, setSelectedInvoice] = useState<UIInvoice | null>(null);
  const [isDetailOpen, setIsDetailOpen] = useState(false);

  // New invoice form
  const [selectedPatient, setSelectedPatient] = useState("");
  const [selectedAppointment, setSelectedAppointment] = useState("");
  const [amount, setAmount] = useState("");
  const [dueDate, setDueDate] = useState("");
  const [notes, setNotes] = useState("");

  useEffect(() => {
    if (user) {
      fetchData();
    }
  }, [user]);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [invoicesData, patientsData, appointmentsData] = await Promise.all([
        invoiceService.list(),
        patientService.list(),
        appointmentService.list()
      ]);

      const allPatients = patientsData || [];
      const allAppointments = appointmentsData || [];
      const allInvoices = invoicesData || [];

      // Join data
      const enrichedInvoices: UIInvoice[] = allInvoices.map(inv => {
        const patient = allPatients.find(p => p.id === inv.patient_id);
        const appointment = allAppointments.find(a => a.id === inv.appointment_id);
        return {
          ...inv,
          patients: {
            first_name: patient?.first_name || "Unknown",
            last_name: patient?.last_name || "Patient"
          },
          appointments: appointment ? {
            start_time: appointment.start_time,
            appointment_type: appointment.appointment_type
          } : null
        };
      });

      // Sort by created_at desc
      enrichedInvoices.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());

      setInvoices(enrichedInvoices);
      setPatients(allPatients.filter(p => p.status === 'active'));
      
      // Filter recent completed appointments for dropdown
      const completedAppointments = allAppointments
        .filter(a => a.status === 'completed')
        .sort((a, b) => new Date(b.start_time).getTime() - new Date(a.start_time).getTime())
        .slice(0, 50);
        
      setAppointments(completedAppointments);

    } catch (error) {
      console.error("Error fetching billing data:", error);
      toast({
        title: "Error",
        description: "Failed to load billing information.",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  // Filter appointments for selected patient
  const patientAppointments = appointments.filter(
    (apt) => apt.patient_id === selectedPatient
  );

  const handleCreateInvoice = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!user || !selectedPatient || !amount) return;

    setIsSubmitting(true);
    const amountCents = Math.round(parseFloat(amount) * 100);

    try {
      await invoiceService.create({
        patient_id: selectedPatient,
        appointment_id: selectedAppointment || undefined,
        amount_cents: amountCents,
        due_date: dueDate || undefined,
        notes: notes || undefined,
        status: "draft",
      });

      toast({
        title: "Invoice Created",
        description: "New invoice has been created.",
      });
      setIsCreateDialogOpen(false);
      resetForm();
      fetchData();
    } catch (error) {
       toast({
        title: "Error",
        description: "Failed to create invoice.",
        variant: "destructive",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const updateInvoiceStatus = async (invoiceId: string, newStatus: string) => {
    const updates: { status: string; paid_at?: string } = { status: newStatus };
    
    if (newStatus === "paid") {
      updates.paid_at = new Date().toISOString();
    }

    try {
      await invoiceService.update(invoiceId, updates);
      toast({
        title: "Status Updated",
        description: `Invoice marked as ${newStatus}.`,
      });
      fetchData();
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to update invoice status.",
        variant: "destructive",
      });
    }
  };

  const handleDownloadSuperbill = async (invoiceId: string) => {
    toast({
      title: "Coming Soon",
      description: "Superbill generation will be available in a future update.",
    });
  };

  const handleViewInvoice = (invoice: UIInvoice) => {
    setSelectedInvoice(invoice);
    setIsDetailOpen(true);
  };

  const handleExportCSV = () => {
    exportInvoicesToCSV(invoices);
    toast({
      title: "Export Complete",
      description: "Invoices exported to CSV file.",
    });
  };

  const resetForm = () => {
    setSelectedPatient("");
    setSelectedAppointment("");
    setAmount("");
    setDueDate("");
    setNotes("");
  };

  const filteredInvoices = invoices.filter((inv) => {
    const matchesSearch = `${inv.patients.first_name} ${inv.patients.last_name}`
      .toLowerCase()
      .includes(search.toLowerCase());
    const matchesStatus = statusFilter === "all" || inv.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const formatCurrency = (cents: number) => {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
    }).format(cents / 100);
  };

  const getStatusBadge = (status: string) => {
    const variants: Record<string, { variant: "default" | "secondary" | "destructive" | "outline"; icon: React.ReactNode }> = {
      draft: { variant: "secondary", icon: <FileText className="w-3 h-3" /> },
      sent: { variant: "outline", icon: <Clock className="w-3 h-3" /> },
      paid: { variant: "default", icon: <CheckCircle className="w-3 h-3" /> },
      overdue: { variant: "destructive", icon: <XCircle className="w-3 h-3" /> },
      cancelled: { variant: "secondary", icon: <XCircle className="w-3 h-3" /> },
    };

    const config = variants[status] || variants.draft;

    return (
      <Badge variant={config.variant} className="gap-1">
        {config.icon}
        {status.charAt(0).toUpperCase() + status.slice(1)}
      </Badge>
    );
  };

  // Stats calculations
  const totalOutstanding = invoices
    .filter(i => ["sent", "overdue"].includes(i.status))
    .reduce((sum, i) => sum + i.amount_cents, 0);
  
  const totalPaid = invoices
    .filter(i => i.status === "paid")
    .reduce((sum, i) => sum + i.amount_cents, 0);

  const overdueCount = invoices.filter(i => i.status === "overdue").length;

  return (
    <DashboardLayout>
      <div className="p-6 lg:p-8">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
          <div>
            <h1 className="text-2xl lg:text-3xl font-bold">Billing</h1>
            <p className="text-muted-foreground mt-1">
              Manage invoices and payments
            </p>
          </div>
          <div className="flex gap-2">
            <Button variant="outline" onClick={handleExportCSV} className="gap-2">
              <FileDown className="w-4 h-4" />
              Export CSV
            </Button>
            <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
              <DialogTrigger asChild>
                <Button className="gap-2">
                  <Plus className="w-4 h-4" />
                  Create Invoice
                </Button>
              </DialogTrigger>
              <DialogContent className="max-w-md">
                <DialogHeader>
                  <DialogTitle>Create Invoice</DialogTitle>
                  <DialogDescription>
                    Create a new invoice for a patient.
                  </DialogDescription>
                </DialogHeader>
                <form onSubmit={handleCreateInvoice} className="space-y-4 mt-4">
                  <div className="space-y-2">
                    <Label>Patient</Label>
                    <Select value={selectedPatient} onValueChange={(val) => {
                      setSelectedPatient(val);
                      setSelectedAppointment("");
                    }}>
                      <SelectTrigger>
                        <SelectValue placeholder="Select patient" />
                      </SelectTrigger>
                      <SelectContent>
                        {patients.map((patient) => (
                          <SelectItem key={patient.id} value={patient.id}>
                            {patient.first_name} {patient.last_name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  {selectedPatient && patientAppointments.length > 0 && (
                    <div className="space-y-2">
                      <Label>Link to Appointment (Optional)</Label>
                      <Select value={selectedAppointment} onValueChange={setSelectedAppointment}>
                        <SelectTrigger>
                          <SelectValue placeholder="Select appointment" />
                        </SelectTrigger>
                        <SelectContent>
                          {patientAppointments.map((apt) => (
                            <SelectItem key={apt.id} value={apt.id}>
                              {format(new Date(apt.start_time), "MMM d, yyyy")} - {apt.appointment_type}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  )}
                  <div className="space-y-2">
                    <Label htmlFor="amount">Amount ($)</Label>
                    <Input
                      id="amount"
                      type="number"
                      step="0.01"
                      min="0"
                      value={amount}
                      onChange={(e) => setAmount(e.target.value)}
                      placeholder="0.00"
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="dueDate">Due Date</Label>
                    <Input
                      id="dueDate"
                      type="date"
                      value={dueDate}
                      onChange={(e) => setDueDate(e.target.value)}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="notes">Notes</Label>
                    <Input
                      id="notes"
                      value={notes}
                      onChange={(e) => setNotes(e.target.value)}
                      placeholder="Session fee, copay, etc."
                    />
                  </div>
                  <Button type="submit" className="w-full" disabled={isSubmitting || !selectedPatient}>
                    {isSubmitting ? "Creating..." : "Create Invoice"}
                  </Button>
                </form>
              </DialogContent>
            </Dialog>
          </div>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-primary/10">
                  <DollarSign className="w-5 h-5 text-primary" />
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Outstanding</p>
                  <p className="text-xl font-bold">{formatCurrency(totalOutstanding)}</p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-green-500/10">
                  <CheckCircle className="w-5 h-5 text-green-500" />
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Collected</p>
                  <p className="text-xl font-bold">{formatCurrency(totalPaid)}</p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-destructive/10">
                  <XCircle className="w-5 h-5 text-destructive" />
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Overdue</p>
                  <p className="text-xl font-bold">{overdueCount} invoice{overdueCount !== 1 ? "s" : ""}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Revenue Chart */}
        <div className="mb-6">
          <RevenueChart invoices={invoices} />
        </div>

        {/* Filters */}
        <Card className="mb-6">
          <CardContent className="p-4">
            <div className="flex flex-col sm:flex-row gap-4">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  placeholder="Search by patient name..."
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className="pl-9"
                />
              </div>
              <Select value={statusFilter} onValueChange={setStatusFilter}>
                <SelectTrigger className="w-full sm:w-40">
                  <SelectValue placeholder="Status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Status</SelectItem>
                  <SelectItem value="draft">Draft</SelectItem>
                  <SelectItem value="sent">Sent</SelectItem>
                  <SelectItem value="paid">Paid</SelectItem>
                  <SelectItem value="overdue">Overdue</SelectItem>
                  <SelectItem value="cancelled">Cancelled</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>

        {/* Invoices Table */}
        <Card>
          <CardContent className="p-0">
            {loading ? (
              <div className="text-center py-12 text-muted-foreground">
                Loading invoices...
              </div>
            ) : filteredInvoices.length === 0 ? (
              <div className="text-center py-12">
                <Receipt className="w-12 h-12 mx-auto mb-3 text-muted-foreground opacity-50" />
                <p className="font-medium">
                  {search || statusFilter !== "all" ? "No invoices found" : "No invoices yet"}
                </p>
                <p className="text-sm text-muted-foreground mt-1">
                  {search || statusFilter !== "all" 
                    ? "Try adjusting your filters" 
                    : "Create your first invoice to get started"}
                </p>
              </div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Patient</TableHead>
                    <TableHead>Amount</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Due Date</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead className="w-12"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredInvoices.map((invoice) => (
                    <TableRow 
                      key={invoice.id} 
                      className="cursor-pointer"
                      onClick={() => handleViewInvoice(invoice)}
                    >
                      <TableCell className="font-medium">
                        {invoice.patients.first_name} {invoice.patients.last_name}
                      </TableCell>
                      <TableCell>{formatCurrency(invoice.amount_cents)}</TableCell>
                      <TableCell>{getStatusBadge(invoice.status)}</TableCell>
                      <TableCell>
                        {invoice.due_date 
                          ? format(new Date(invoice.due_date), "MMM d, yyyy")
                          : "-"}
                      </TableCell>
                      <TableCell>
                        {format(new Date(invoice.created_at), "MMM d, yyyy")}
                      </TableCell>
                      <TableCell onClick={(e) => e.stopPropagation()}>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon">
                              <MoreHorizontal className="w-4 h-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => handleViewInvoice(invoice)}>
                              <FileText className="w-4 h-4 mr-2" />
                              View Details
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() => handleDownloadSuperbill(invoice.id)}>
                              <Download className="w-4 h-4 mr-2" />
                              Download Superbill
                            </DropdownMenuItem>
                            {invoice.status === "draft" && (
                              <DropdownMenuItem onClick={() => updateInvoiceStatus(invoice.id, "sent")}>
                                Mark as Sent
                              </DropdownMenuItem>
                            )}
                            {["draft", "sent", "overdue"].includes(invoice.status) && (
                              <DropdownMenuItem onClick={() => updateInvoiceStatus(invoice.id, "paid")}>
                                Mark as Paid
                              </DropdownMenuItem>
                            )}
                            {invoice.status !== "cancelled" && invoice.status !== "paid" && (
                              <DropdownMenuItem onClick={() => updateInvoiceStatus(invoice.id, "cancelled")}>
                                Cancel Invoice
                              </DropdownMenuItem>
                            )}
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>

        {/* Invoice Detail Dialog */}
        <InvoiceDetailDialog
          invoice={selectedInvoice}
          open={isDetailOpen}
          onOpenChange={setIsDetailOpen}
          onUpdate={fetchData}
        />
      </div>
    </DashboardLayout>
  );
};

export default Billing;
