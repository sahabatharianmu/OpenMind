import { useState, useEffect } from "react";
import DashboardLayout from "@/components/dashboard/DashboardLayout";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
  Plus, 
  Video,
  MapPin,
  ChevronLeft,
  ChevronRight
} from "lucide-react";
import appointmentService from "@/services/appointmentService";
import patientService from "@/services/patientService";
import { useAuth } from "@/contexts/AuthContext";
import { useToast } from "@/hooks/use-toast";
import { format, startOfWeek, addDays, isSameDay, parseISO, addWeeks, subWeeks } from "date-fns";
import { Appointment, Patient } from "@/types";

// Extended Appointment type for UI to include resolved patient details
interface UIAppointment extends Appointment {
  patient?: Patient;
}

const Appointments = () => {
  const { user } = useAuth();
  const { toast } = useToast();
  const [appointments, setAppointments] = useState<UIAppointment[]>([]);
  const [patients, setPatients] = useState<Patient[]>([]);
  const [loading, setLoading] = useState(true);
  const [currentWeekStart, setCurrentWeekStart] = useState(startOfWeek(new Date(), { weekStartsOn: 1 }));
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // New appointment form
  const [selectedPatientId, setSelectedPatientId] = useState("");
  const [selectedDate, setSelectedDate] = useState("");
  const [selectedTime, setSelectedTime] = useState("");
  const [duration, setDuration] = useState("50");
  const [appointmentType, setAppointmentType] = useState("session");
  const [mode, setMode] = useState("in-person");
  const [cptCode, setCptCode] = useState("90837");

  useEffect(() => {
    fetchData();
  }, [user, currentWeekStart]);

  const fetchData = async () => {
    if (!user) return;
    setLoading(true);

    try {
      const [appointmentsData, patientsData] = await Promise.all([
        appointmentService.list(),
        patientService.list()
      ]);

      const allPatients = patientsData || [];
      const allAppointments = appointmentsData || [];

      // Map patient details to appointments
      const enrichedAppointments = allAppointments.map(apt => ({
        ...apt,
        patient: allPatients.find(p => p.id === apt.patient_id)
      }));

      // Filter for current week (Client-side filtering for now)
      const weekEnd = addDays(currentWeekStart, 7);
      const weeklyAppointments = enrichedAppointments.filter(apt => {
        const start = new Date(apt.start_time);
        return start >= currentWeekStart && start < weekEnd;
      });

      setAppointments(weeklyAppointments);
      setPatients(allPatients.filter(p => p.status === 'active'));

    } catch (error) {
      console.error("Error fetching data:", error);
      toast({
        title: "Error",
        description: "Failed to load schedule.",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleAddAppointment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!user || !selectedPatientId || !selectedDate || !selectedTime) return;

    setIsSubmitting(true);

    const startTime = new Date(`${selectedDate}T${selectedTime}`);
    const endTime = new Date(startTime.getTime() + parseInt(duration) * 60000);

    try {
      await appointmentService.create({
        patient_id: selectedPatientId,
        clinician_id: user.id,
        start_time: startTime.toISOString(),
        end_time: endTime.toISOString(),
        appointment_type: appointmentType,
        mode: mode,
        cpt_code: cptCode,
        status: 'scheduled'
      });

      toast({
        title: "Appointment Scheduled",
        description: "The appointment has been added to your calendar.",
      });
      setIsAddDialogOpen(false);
      resetForm();
      fetchData();
    } catch (error: any) {
       toast({
        title: "Error",
        description: error.response?.data?.error?.message || "Failed to schedule appointment.",
        variant: "destructive",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const resetForm = () => {
    setSelectedPatientId("");
    setSelectedDate("");
    setSelectedTime("");
    setDuration("50");
    setAppointmentType("session");
    setMode("in-person");
  };

  const weekDays = Array.from({ length: 7 }, (_, i) => addDays(currentWeekStart, i));

  const getAppointmentsForDay = (date: Date) => {
    return appointments.filter((apt) => isSameDay(parseISO(apt.start_time), date));
  };

  return (
    <DashboardLayout>
      <div className="p-6 lg:p-8">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
          <div>
            <h1 className="text-2xl lg:text-3xl font-bold">Appointments</h1>
            <p className="text-muted-foreground mt-1">
              Manage your practice schedule
            </p>
          </div>
          <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
            <DialogTrigger asChild>
              <Button className="gap-2">
                <Plus className="w-4 h-4" />
                New Appointment
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Schedule Appointment</DialogTitle>
                <DialogDescription>
                  Create a new appointment for a patient.
                </DialogDescription>
              </DialogHeader>
              <form onSubmit={handleAddAppointment} className="space-y-4 mt-4">
                <div className="space-y-2">
                  <Label>Patient</Label>
                  <Select value={selectedPatientId} onValueChange={setSelectedPatientId}>
                    <SelectTrigger>
                      <SelectValue placeholder="Select a patient" />
                    </SelectTrigger>
                    <SelectContent>
                      {patients.map((patient) => (
                        <SelectItem key={patient.id} value={patient.id}>
                          {patient.last_name}, {patient.first_name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Date</Label>
                    <Input
                      type="date"
                      value={selectedDate}
                      onChange={(e) => setSelectedDate(e.target.value)}
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Time</Label>
                    <Input
                      type="time"
                      value={selectedTime}
                      onChange={(e) => setSelectedTime(e.target.value)}
                      required
                    />
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Duration</Label>
                    <Select value={duration} onValueChange={setDuration}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="25">25 minutes</SelectItem>
                        <SelectItem value="50">50 minutes</SelectItem>
                        <SelectItem value="80">80 minutes</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-2">
                    <Label>Type</Label>
                    <Select value={appointmentType} onValueChange={setAppointmentType}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="session">Therapy Session</SelectItem>
                        <SelectItem value="intake">Initial Intake</SelectItem>
                        <SelectItem value="follow-up">Follow-up</SelectItem>
                        <SelectItem value="consultation">Consultation</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Mode</Label>
                    <Select value={mode} onValueChange={setMode}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="in-person">In-person</SelectItem>
                        <SelectItem value="video">Video Call</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="cptCode">CPT Code</Label>
                    <Input
                      id="cptCode"
                      placeholder="e.g. 90837"
                      value={cptCode}
                      onChange={(e) => setCptCode(e.target.value)}
                    />
                  </div>
                </div>
                <Button type="submit" className="w-full" disabled={isSubmitting}>
                  {isSubmitting ? "Scheduling..." : "Schedule Appointment"}
                </Button>
              </form>
            </DialogContent>
          </Dialog>
        </div>

        {/* Week Navigation */}
        <Card className="mb-6">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <Button
                variant="outline"
                size="icon"
                onClick={() => setCurrentWeekStart(subWeeks(currentWeekStart, 1))}
              >
                <ChevronLeft className="w-4 h-4" />
              </Button>
              <h2 className="font-semibold">
                {format(currentWeekStart, "MMMM d")} - {format(addDays(currentWeekStart, 6), "MMMM d, yyyy")}
              </h2>
              <Button
                variant="outline"
                size="icon"
                onClick={() => setCurrentWeekStart(addWeeks(currentWeekStart, 1))}
              >
                <ChevronRight className="w-4 h-4" />
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Calendar Grid */}
        {loading ? (
          <div className="text-center py-12 text-muted-foreground">
            Loading appointments...
          </div>
        ) : (
          <div className="grid grid-cols-7 gap-2">
            {weekDays.map((day) => {
              const dayAppointments = getAppointmentsForDay(day);
              const isToday = isSameDay(day, new Date());

              return (
                <Card key={day.toISOString()} className={isToday ? "border-primary" : ""}>
                  <CardHeader className="p-3 pb-2">
                    <CardTitle className={`text-sm font-medium ${isToday ? "text-primary" : ""}`}>
                      <div>{format(day, "EEE")}</div>
                      <div className="text-2xl">{format(day, "d")}</div>
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="p-2 pt-0 space-y-2 min-h-[200px]">
                    {dayAppointments.length === 0 ? (
                      <p className="text-xs text-muted-foreground text-center py-4">
                        No appointments
                      </p>
                    ) : (
                      dayAppointments.map((apt) => (
                        <div
                          key={apt.id}
                          className="p-2 rounded bg-primary/10 border border-primary/20 text-xs cursor-pointer hover:bg-primary/20 transition-colors"
                        >
                          <div className="font-medium text-primary">
                            {format(parseISO(apt.start_time), "h:mm a")}
                          </div>
                          <div className="truncate">
                            {apt.patient?.first_name} {apt.patient?.last_name?.[0]}.
                          </div>
                          <div className="flex items-center gap-1 mt-1 text-muted-foreground">
                            {apt.mode === "video" ? (
                              <Video className="w-3 h-3" />
                            ) : (
                              <MapPin className="w-3 h-3" />
                            )}
                          </div>
                        </div>
                      ))
                    )}
                  </CardContent>
                </Card>
              );
            })}
          </div>
        )}
      </div>
    </DashboardLayout>
  );
};

export default Appointments;
