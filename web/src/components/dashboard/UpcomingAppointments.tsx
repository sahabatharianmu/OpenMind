import { useAppointments, usePatients } from "@/hooks/useDashboardQueries";
import { Clock, Video, MapPin, Plus, Calendar as CalendarIcon, User } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { format, parseISO } from "date-fns";
import { Appointment } from "@/types";
import { useAuth } from "@/contexts/AuthContext";

interface UIAppointment extends Appointment {
  patients: {
    first_name: string;
    last_name: string;
  } | null;
}

const UpcomingAppointments = () => {
  const { user } = useAuth();
  const { data: appointments, isLoading: appointmentsLoading } = useAppointments();
  const { data: patients, isLoading: patientsLoading } = usePatients();
  
  const loading = appointmentsLoading || patientsLoading;
  
  // Calculate upcoming appointments
  const allAppointments = appointments || [];
  const allPatients = patients || [];

  // Filter for upcoming (start_time >= today)
  const now = new Date();
  now.setHours(0, 0, 0, 0); // Start of today

  const upcoming = allAppointments.filter(apt => {
    try {
      const start = parseISO(apt.start_time);
      return start >= now;
    } catch { return false; }
  });

  // Sort by start_time ascending
  upcoming.sort((a, b) => {
    const dateA = new Date(a.start_time).getTime();
    const dateB = new Date(b.start_time).getTime();
    return (isNaN(dateA) ? 0 : dateA) - (isNaN(dateB) ? 0 : dateB);
  });

  // Take top 5
  const limited = upcoming.slice(0, 5);

  // Join with patients
  const enrichedAppointments: UIAppointment[] = limited.map(apt => {
    const patient = allPatients.find(p => p.id === apt.patient_id);
    return {
      ...apt,
      patients: patient ? {
        first_name: patient.first_name,
        last_name: patient.last_name
      } : null
    };
  });

  const formatDuration = (start: string, end: string) => {
    try {
      const startDate = parseISO(start);
      const endDate = parseISO(end);
      const diffMs = endDate.getTime() - startDate.getTime();
      const diffMins = Math.round(diffMs / 60000);
      return `${diffMins} min`;
    } catch {
      return "";
    }
  };

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-xl">Upcoming Appointments</CardTitle>
          <Button variant="outline" size="sm">
            View Calendar
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="text-center py-8 text-muted-foreground">Loading appointments...</div>
        ) : enrichedAppointments.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <CalendarIcon className="w-12 h-12 mx-auto mb-3 opacity-50" />
            <p>No upcoming appointments</p>
            <Button variant="outline" size="sm" className="mt-3 gap-1">
              <Plus className="w-4 h-4" />
              Schedule Appointment
            </Button>
          </div>
        ) : (
          <div className="space-y-3">
            {enrichedAppointments.map((apt) => (
              <div
                key={apt.id}
                className="flex items-center justify-between p-3 rounded-lg border border-border hover:border-primary/50 transition-colors"
              >
                <div className="flex items-center gap-4 flex-1 min-w-0">
                  <div className="text-center min-w-[70px] flex-shrink-0">
                    <div className="flex items-center justify-center gap-1 mb-1">
                      <Clock className="w-4 h-4 text-primary" />
                      <p className="font-semibold text-primary">
                        {(() => {
                          try {
                            return format(parseISO(apt.start_time), "h:mm a");
                          } catch {
                            return "--:--";
                          }
                        })()}
                      </p>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      {formatDuration(apt.start_time, apt.end_time)}
                    </p>
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <User className="w-4 h-4 text-muted-foreground flex-shrink-0" />
                      <p className="font-medium truncate">
                        {apt.patients?.first_name} {apt.patients?.last_name}
                      </p>
                    </div>
                    {(user?.role === "admin" || user?.role === "owner") && apt.clinician_name && (
                      <div className="flex items-center gap-2 mb-1">
                        <User className="w-4 h-4 text-muted-foreground flex-shrink-0" />
                        <p className="text-sm text-muted-foreground truncate">
                          Dr. {apt.clinician_name}
                        </p>
                      </div>
                    )}
                    <div className="flex items-center gap-2 flex-wrap">
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
                  </div>
                </div>
                <Button variant="ghost" size="sm">
                  Details
                </Button>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
};

export default UpcomingAppointments;
