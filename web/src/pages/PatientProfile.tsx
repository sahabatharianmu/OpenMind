import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import DashboardLayout from "@/components/dashboard/DashboardLayout";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Separator } from "@/components/ui/separator";
import {
  ArrowLeft,
  Mail,
  Phone,
  MapPin,
  Calendar,
  FileText,
  Receipt,
  Plus,
  User,
  Clock,
  Video,
} from "lucide-react";
import patientService from "@/services/patientService";
import appointmentService from "@/services/appointmentService";
import clinicalNoteService from "@/services/clinicalNoteService";
import invoiceService from "@/services/invoiceService";
import { useAuth } from "@/contexts/AuthContext";
import { format } from "date-fns";
import { Patient, Appointment, ClinicalNote, Invoice } from "@/types";
import PatientAssignments from "@/components/patient/PatientAssignments";
import PatientHandoff from "@/components/patient/PatientHandoff";
import { UserCog } from "lucide-react";

const PatientProfile = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [patient, setPatient] = useState<Patient | null>(null);
  const [appointments, setAppointments] = useState<Appointment[]>([]);
  const [notes, setNotes] = useState<ClinicalNote[]>([]);
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [loading, setLoading] = useState(true);
  const [isAssigned, setIsAssigned] = useState(false);
  const [checkingAssignment, setCheckingAssignment] = useState(true);

  useEffect(() => {
    if (user && id) {
      fetchPatientData();
    }
  }, [user, id]);

  const fetchPatientData = async () => {
    if (!id) return;
    setLoading(true);

    try {
      // Parallel fetch
      const [patientData, allAppointments, allNotes, allInvoices] = await Promise.all([
        patientService.get(id),
        // TEMPORARY: Fetching ALL and filtering client side until I update services.
        // Ideally: appointmentService.list({ patient_id: id })
        appointmentService.list(),
        clinicalNoteService.list(),
        invoiceService.list()
      ]);

      if (patientData) {
        setPatient(patientData);
        
        // Filter client-side for now (Optimize later by updating API/Services)
        // Handle null/undefined responses by defaulting to empty arrays
        setAppointments((allAppointments || []).filter(a => a.patient_id === id));
        
        // Check if user is assigned to this patient
        setCheckingAssignment(true);
        try {
          const assigned = await patientService.isAssigned(id, user?.id);
          console.log("Assignment status for patient:", id, "user:", user?.id, "isAssigned:", assigned);
          setIsAssigned(assigned);
          
          // Only fetch clinical notes if user is assigned
          if (assigned) {
            setNotes((allNotes || []).filter(n => n.patient_id === id));
          } else {
            setNotes([]);
          }
        } catch (error) {
          console.error("Error checking assignment:", error);
          setIsAssigned(false);
          setNotes([]);
        } finally {
          setCheckingAssignment(false);
        }
        
        setInvoices((allInvoices || []).filter(i => i.patient_id === id));
      }
    } catch (error: unknown) {
      console.error("Error fetching patient details:", error);
      const err = error as { response?: { status?: number; data?: { error?: { message?: string } } } };
      if (err.response?.status === 403 || err.response?.status === 404) {
        // Patient not found or access denied - will show "not found" message
        setPatient(null);
      } else if (err.response?.status === 401) {
        // 401 will be handled by API interceptor, but we should still set patient to null
        // to prevent rendering issues
        setPatient(null);
      }
    } finally {
      setLoading(false);
      setCheckingAssignment(false);
    }
  };

  const formatCurrency = (cents: number) => {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
    }).format(cents / 100);
  };

  const getStatusBadge = (status: string) => {
    const variants: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
      active: "default",
      inactive: "secondary",
      scheduled: "outline",
      completed: "default",
      cancelled: "destructive",
      paid: "default",
      draft: "secondary",
      sent: "outline",
      overdue: "destructive",
    };
    return <Badge variant={variants[status] || "secondary"}>{status}</Badge>;
  };

  if (loading) {
    return (
      <DashboardLayout>
        <div className="p-6 lg:p-8">
          <div className="text-center py-12 text-muted-foreground">
            Loading patient...
          </div>
        </div>
      </DashboardLayout>
    );
  }

  if (!patient) {
    return (
      <DashboardLayout>
        <div className="p-6 lg:p-8">
          <div className="text-center py-12">
            <p className="font-medium">Patient not found</p>
            <Button variant="link" onClick={() => navigate("/dashboard/patients")}>
              Back to Patients
            </Button>
          </div>
        </div>
      </DashboardLayout>
    );
  }

  const age = Math.floor(
    (new Date().getTime() - new Date(patient.date_of_birth).getTime()) /
      (365.25 * 24 * 60 * 60 * 1000)
  );

  return (
    <DashboardLayout>
      <div className="p-6 lg:p-8">
        {/* Header */}
        <div className="flex items-center gap-4 mb-6">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => navigate("/dashboard/patients")}
          >
            <ArrowLeft className="w-5 h-5" />
          </Button>
          <div className="flex items-center gap-4 flex-1">
            <Avatar className="h-16 w-16">
              <AvatarFallback className="bg-primary/10 text-primary text-xl">
                {patient.first_name[0]}
                {patient.last_name[0]}
              </AvatarFallback>
            </Avatar>
            <div>
              <h1 className="text-2xl lg:text-3xl font-bold">
                {patient.first_name} {patient.last_name}
              </h1>
              <div className="flex items-center gap-3 mt-1">
                {getStatusBadge(patient.status)}
                <span className="text-muted-foreground">
                  {age} years old • DOB: {format(new Date(patient.date_of_birth), "MMM d, yyyy")}
                </span>
              </div>
            </div>
          </div>
          <div className="flex gap-2">
            {isAssigned && (
              <Button
                variant="outline"
                onClick={() => navigate(`/dashboard/notes/new?patient=${patient.id}`)}
              >
                <FileText className="w-4 h-4 mr-2" />
                Add Note
              </Button>
            )}
            <Button onClick={() => navigate("/dashboard/appointments")}>
              <Calendar className="w-4 h-4 mr-2" />
              Schedule
            </Button>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Contact Info Sidebar */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Contact Information</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {patient.email && (
                <div className="flex items-center gap-3">
                  <Mail className="w-4 h-4 text-muted-foreground" />
                  <span className="text-sm">{patient.email}</span>
                </div>
              )}
              {patient.phone && (
                <div className="flex items-center gap-3">
                  <Phone className="w-4 h-4 text-muted-foreground" />
                  <span className="text-sm">{patient.phone}</span>
                </div>
              )}
              {patient.address && (
                <div className="flex items-center gap-3">
                  <MapPin className="w-4 h-4 text-muted-foreground" />
                  <span className="text-sm">{patient.address}</span>
                </div>
              )}
              {!patient.email && !patient.phone && !patient.address && (
                <p className="text-sm text-muted-foreground">No contact info on file</p>
              )}
              <Separator />
              <div>
                <p className="text-xs text-muted-foreground">Patient Since</p>
                <p className="text-sm font-medium">
                  {format(new Date(patient.created_at), "MMMM d, yyyy")}
                </p>
              </div>
            </CardContent>
          </Card>

          {/* Main Content */}
          <div className="lg:col-span-2 space-y-6">
            {/* Assigned Clinicians */}
            <PatientAssignments patientId={patient.id} onAssignmentsUpdated={fetchPatientData} />

            {/* Patient Handoffs */}
            <PatientHandoff patientId={patient.id} onHandoffUpdated={fetchPatientData} />

            <Tabs defaultValue="appointments">
              <TabsList>
                <TabsTrigger value="appointments" className="gap-2">
                  <Calendar className="w-4 h-4" />
                  Appointments ({appointments.length})
                </TabsTrigger>
                <TabsTrigger value="notes" className="gap-2" disabled={!isAssigned && !checkingAssignment}>
                  <FileText className="w-4 h-4" />
                  Notes ({notes.length})
                </TabsTrigger>
                <TabsTrigger value="billing" className="gap-2">
                  <Receipt className="w-4 h-4" />
                  Billing ({invoices.length})
                </TabsTrigger>
              </TabsList>

              <TabsContent value="appointments" className="mt-4">
                <Card>
                  <CardContent className="p-0">
                    {appointments.length === 0 ? (
                      <div className="text-center py-8">
                        <Calendar className="w-10 h-10 mx-auto mb-2 text-muted-foreground opacity-50" />
                        <p className="text-sm text-muted-foreground">No appointments yet</p>
                      </div>
                    ) : (
                      <div className="divide-y">
                        {appointments.map((apt) => (
                          <div key={apt.id} className="p-4 flex items-start justify-between gap-4 hover:bg-muted/50 transition-colors">
                            <div className="flex-1 space-y-2">
                              <div className="flex items-center gap-2">
                                <Clock className="w-4 h-4 text-primary flex-shrink-0" />
                                <div>
                                  <p className="font-medium">
                                    {format(new Date(apt.start_time), "MMM d, yyyy")} at{" "}
                                    {format(new Date(apt.start_time), "h:mm a")}
                                  </p>
                                  <p className="text-xs text-muted-foreground mt-0.5">
                                    {format(new Date(apt.start_time), "h:mm a")} - {format(new Date(apt.end_time), "h:mm a")}
                                  </p>
                                </div>
                              </div>
                              <div className="flex items-center gap-2">
                                <Badge variant="outline" className="text-xs">
                                  {apt.appointment_type}
                                </Badge>
                                <span className="flex items-center gap-1 text-xs text-muted-foreground">
                                  {apt.mode === "video" ? (
                                    <>
                                      <Video className="w-3 h-3" />
                                      Video
                                    </>
                                  ) : (
                                    <>
                                      <MapPin className="w-3 h-3" />
                                      In-person
                                    </>
                                  )}
                                </span>
                              </div>
                              {(user?.role === "admin" || user?.role === "owner") && apt.clinician_name && (
                                <div className="flex items-center gap-2 pt-1">
                                  <User className="w-4 h-4 text-muted-foreground flex-shrink-0" />
                                  <div>
                                    <p className="text-xs text-muted-foreground">Clinician</p>
                                    <p className="text-sm font-medium">
                                      Dr. {apt.clinician_name}
                                    </p>
                                  </div>
                                </div>
                              )}
                            </div>
                            <div className="flex-shrink-0">
                              {getStatusBadge(apt.status)}
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="notes" className="mt-4">
                <Card>
                  <CardContent className="p-0">
                    {!isAssigned && !checkingAssignment ? (
                      <div className="text-center py-8">
                        <FileText className="w-10 h-10 mx-auto mb-2 text-muted-foreground opacity-50" />
                        <p className="text-sm font-medium mb-1">Access Restricted</p>
                        <p className="text-sm text-muted-foreground">
                          You must be assigned to this patient to view clinical notes.
                        </p>
                      </div>
                    ) : notes.length === 0 ? (
                      <div className="text-center py-8">
                        <FileText className="w-10 h-10 mx-auto mb-2 text-muted-foreground opacity-50" />
                        <p className="text-sm text-muted-foreground">No clinical notes yet</p>
                        <Button
                          variant="link"
                          size="sm"
                          onClick={() => navigate(`/dashboard/notes/new?patient=${patient.id}`)}
                        >
                          <Plus className="w-3 h-3 mr-1" />
                          Create First Note
                        </Button>
                      </div>
                    ) : (
                      <div className="divide-y">
                        {notes.map((note) => (
                          <div
                            key={note.id}
                            className="p-4 hover:bg-muted/50 cursor-pointer"
                            onClick={() => navigate(`/dashboard/notes/${note.id}`)}
                          >
                            <div className="flex items-center justify-between mb-1">
                              <p className="font-medium capitalize">{note.note_type} Note</p>
                              <div className="flex items-center gap-2">
                                {note.is_signed && (
                                  <Badge variant="default" className="text-xs">Signed</Badge>
                                )}
                                <span className="text-xs text-muted-foreground">
                                  {format(new Date(note.created_at), "MMM d, yyyy")}
                                </span>
                              </div>
                            </div>
                            {note.assessment && (
                              <p className="text-sm text-muted-foreground line-clamp-2">
                                {note.assessment}
                              </p>
                            )}
                          </div>
                        ))}
                      </div>
                    )}
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="billing" className="mt-4">
                <Card>
                  <CardContent className="p-0">
                    {invoices.length === 0 ? (
                      <div className="text-center py-8">
                        <Receipt className="w-10 h-10 mx-auto mb-2 text-muted-foreground opacity-50" />
                        <p className="text-sm text-muted-foreground">No invoices yet</p>
                      </div>
                    ) : (
                      <div className="divide-y">
                        {invoices.map((inv) => (
                          <div key={inv.id} className="p-4 flex items-center justify-between">
                            <div>
                              <p className="font-medium">{formatCurrency(inv.amount_cents)}</p>
                              <p className="text-sm text-muted-foreground">
                                {format(new Date(inv.created_at), "MMM d, yyyy")}
                                {inv.due_date && ` • Due ${format(new Date(inv.due_date), "MMM d")}`}
                              </p>
                            </div>
                            {getStatusBadge(inv.status)}
                          </div>
                        ))}
                      </div>
                    )}
                  </CardContent>
                </Card>
              </TabsContent>
            </Tabs>
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
};

export default PatientProfile;
