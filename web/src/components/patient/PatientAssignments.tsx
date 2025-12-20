import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
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
import { Badge } from "@/components/ui/badge";
import { useToast } from "@/hooks/use-toast";
import { useAuth } from "@/contexts/AuthContext";
import patientService, { type ClinicianAssignment } from "@/services/patientService";
import { organizationService, type TeamMember } from "@/services/organizationService";
import { UserPlus, X, User } from "lucide-react";
import { format } from "date-fns";

interface PatientAssignmentsProps {
  patientId: string;
}

const PatientAssignments = ({ patientId }: PatientAssignmentsProps) => {
  const { user } = useAuth();
  const { toast } = useToast();
  const [assignments, setAssignments] = useState<ClinicianAssignment[]>([]);
  const [teamMembers, setTeamMembers] = useState<TeamMember[]>([]);
  const [loading, setLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(true);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [selectedClinicianId, setSelectedClinicianId] = useState("");
  const [selectedRole, setSelectedRole] = useState<"primary" | "secondary">("primary");

  useEffect(() => {
    if (patientId && user) {
      loadAssignments();
      loadTeamMembers();
    }
  }, [patientId, user]);

  const loadAssignments = async () => {
    if (!patientId) {
      setInitialLoading(false);
      return;
    }
    try {
      const data = await patientService.getAssignedClinicians(patientId);
      setAssignments(data || []);
    } catch (error) {
      console.error("Error loading assignments:", error);
      // Don't show toast on initial load errors, just log
      setAssignments([]);
    } finally {
      setInitialLoading(false);
    }
  };

  const loadTeamMembers = async () => {
    try {
      // Only load team members if user has admin/owner role
      // This endpoint requires admin/owner permissions
      if (user?.role !== "admin" && user?.role !== "owner") {
        setTeamMembers([]);
        return;
      }
      const data = await organizationService.listTeamMembers();
      setTeamMembers(data || []);
    } catch (error: unknown) {
      console.error("Error loading team members:", error);
      const err = error as { response?: { status?: number } };
      // Don't let 401/403 errors from team members affect the page
      // Just set empty array and continue
      if (err.response?.status === 401 || err.response?.status === 403) {
        setTeamMembers([]);
        return;
      }
      setTeamMembers([]);
    }
  };

  const handleAssign = async () => {
    if (!selectedClinicianId) {
      toast({
        title: "Error",
        description: "Please select a clinician",
        variant: "destructive",
      });
      return;
    }

    if (!patientId) {
      toast({
        title: "Error",
        description: "Patient ID is missing",
        variant: "destructive",
      });
      return;
    }

    setLoading(true);
    try {
      await patientService.assignClinician(patientId, {
        clinician_id: selectedClinicianId,
        role: selectedRole,
      });
      toast({
        title: "Success",
        description: "Clinician assigned successfully",
      });
      setIsDialogOpen(false);
      setSelectedClinicianId("");
      setSelectedRole("primary");
      await loadAssignments();
    } catch (error: unknown) {
      console.error("Error assigning clinician:", error);
      const err = error as { response?: { data?: { error?: { message?: string } } }; message?: string };
      const message = err.response?.data?.error?.message || err.message || "Failed to assign clinician";
      toast({
        title: "Error",
        description: message,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleUnassign = async (clinicianId: string) => {
    if (!confirm("Are you sure you want to remove this clinician from the patient?")) {
      return;
    }

    setLoading(true);
    try {
      await patientService.unassignClinician(patientId, clinicianId);
      toast({
        title: "Success",
        description: "Clinician unassigned successfully",
      });
      loadAssignments();
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: { message?: string } } }; message?: string };
      const message = err.response?.data?.error?.message || err.message || "Failed to unassign clinician";
      toast({
        title: "Error",
        description: message,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  // Get available clinicians (not already assigned)
  // Only include users with clinician roles: admin, owner, clinician, case_manager
  const availableClinicians = teamMembers.filter(
    (member) => {
      const isClinician = member.role === "admin" || 
                         member.role === "owner" || 
                         member.role === "clinician" || 
                         member.role === "case_manager";
      const isNotAssigned = !assignments.some((assignment) => assignment.clinician_id === member.user_id);
      return isClinician && isNotAssigned;
    }
  );

  if (initialLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Assigned Clinicians</CardTitle>
          <CardDescription>
            Manage which clinicians have access to this patient
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8 text-muted-foreground">Loading...</div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Assigned Clinicians</CardTitle>
            <CardDescription>
              Manage which clinicians have access to this patient
            </CardDescription>
          </div>
          <Dialog open={isDialogOpen} onOpenChange={(open) => {
            setIsDialogOpen(open);
            if (!open) {
              // Reset form when dialog closes
              setSelectedClinicianId("");
              setSelectedRole("primary");
            }
          }}>
            <DialogTrigger asChild>
              <Button size="sm" className="gap-2" disabled={loading}>
                <UserPlus className="w-4 h-4" />
                Assign Clinician
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Assign Clinician</DialogTitle>
                <DialogDescription>
                  Select a clinician to assign to this patient
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4 py-4">
                <div className="space-y-2">
                  <label className="text-sm font-medium">Clinician</label>
                  <Select 
                    value={selectedClinicianId} 
                    onValueChange={setSelectedClinicianId}
                    disabled={loading || availableClinicians.length === 0}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select a clinician" />
                    </SelectTrigger>
                    <SelectContent>
                      {availableClinicians.length === 0 ? (
                        <div className="p-2 text-sm text-muted-foreground">
                          No available clinicians to assign. All clinicians are already assigned.
                        </div>
                      ) : (
                        availableClinicians.map((member) => (
                          <SelectItem key={member.user_id} value={member.user_id}>
                            {member.full_name || "Unknown"} ({member.email || "No email"})
                          </SelectItem>
                        ))
                      )}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Role</label>
                  <Select
                    value={selectedRole}
                    onValueChange={(value) => setSelectedRole(value as "primary" | "secondary")}
                    disabled={loading}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="primary">Primary</SelectItem>
                      <SelectItem value="secondary">Secondary</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
              <DialogFooter>
                <Button 
                  variant="outline" 
                  onClick={() => setIsDialogOpen(false)}
                  disabled={loading}
                >
                  Cancel
                </Button>
                <Button 
                  onClick={handleAssign} 
                  disabled={loading || !selectedClinicianId || availableClinicians.length === 0}
                >
                  {loading ? "Assigning..." : "Assign"}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </CardHeader>
      <CardContent>
        {assignments.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            No clinicians assigned yet. Assign a clinician to grant access to this patient.
          </div>
        ) : (
          <div className="space-y-3">
            {assignments.map((assignment) => (
              <div
                key={assignment.clinician_id}
                className="flex items-center justify-between p-3 rounded-lg border"
              >
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
                    <User className="w-5 h-5 text-primary" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <p className="font-medium">{assignment.full_name || "Unknown Clinician"}</p>
                      <Badge variant={assignment.role === "primary" ? "default" : "secondary"}>
                        {assignment.role || "secondary"}
                      </Badge>
                    </div>
                    <p className="text-sm text-muted-foreground">{assignment.email || "No email"}</p>
                    {assignment.assigned_at && (
                      <p className="text-xs text-muted-foreground">
                        Assigned {format(new Date(assignment.assigned_at), "MMM d, yyyy")}
                      </p>
                    )}
                  </div>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleUnassign(assignment.clinician_id)}
                  disabled={loading}
                  className="text-destructive hover:text-destructive"
                >
                  <X className="w-4 h-4" />
                </Button>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
};

export default PatientAssignments;

