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
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { useToast } from "@/hooks/use-toast";
import { useAuth } from "@/contexts/AuthContext";
import patientService, { type Handoff, type RequestHandoffRequest } from "@/services/patientService";
import { organizationService, type TeamMember } from "@/services/organizationService";
import { UserPlus, X, CheckCircle2, XCircle, Clock, ArrowRight } from "lucide-react";
import { format } from "date-fns";

interface PatientHandoffProps {
  patientId: string;
  onHandoffUpdated?: () => void;
}

const PatientHandoff = ({ patientId, onHandoffUpdated }: PatientHandoffProps) => {
  const { user } = useAuth();
  const { toast } = useToast();
  const [handoffs, setHandoffs] = useState<Handoff[]>([]);
  const [teamMembers, setTeamMembers] = useState<TeamMember[]>([]);
  const [loading, setLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(true);
  const [isRequestDialogOpen, setIsRequestDialogOpen] = useState(false);
  const [selectedClinicianId, setSelectedClinicianId] = useState("");
  const [selectedRole, setSelectedRole] = useState<"primary" | "secondary" | undefined>(undefined);
  const [message, setMessage] = useState("");

  useEffect(() => {
    if (patientId && user) {
      loadHandoffs();
      loadTeamMembers();
    }
  }, [patientId, user]);

  const loadHandoffs = async () => {
    if (!patientId) {
      setInitialLoading(false);
      return;
    }
    try {
      const data = await patientService.listHandoffs(patientId);
      setHandoffs(data || []);
    } catch (error) {
      console.error("Error loading handoffs:", error);
      setHandoffs([]);
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
      // Filter to only include roles that can be assigned as clinicians
      const assignableRoles = ["admin", "owner", "clinician", "case_manager"];
      setTeamMembers(data.filter(member => assignableRoles.includes(member.role)) || []);
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

  const handleRequestHandoff = async () => {
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
      // Build request data, only including defined fields
      const requestData: RequestHandoffRequest = {
        receiving_clinician_id: selectedClinicianId,
      };
      
      if (message && message.trim()) {
        requestData.message = message.trim();
      }
      
      if (selectedRole) {
        requestData.role = selectedRole;
      }
      
      await patientService.requestHandoff(patientId, requestData);
      toast({
        title: "Success",
        description: "Handoff request sent successfully",
      });
      setIsRequestDialogOpen(false);
      setSelectedClinicianId("");
      setSelectedRole(undefined);
      setMessage("");
      loadHandoffs();
      onHandoffUpdated?.();
    } catch (error: unknown) {
      console.error("Error requesting handoff:", error);
      const err = error as { response?: { status?: number; data?: { error?: { message?: string } } }; message?: string };
      
      // Don't show error toast for 401/403 - let the auth system handle it
      if (err.response?.status === 401 || err.response?.status === 403) {
        // The interceptor will handle the redirect
        return;
      }
      
      const errorMessage = err.response?.data?.error?.message || err.message || "Failed to request handoff";
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleApprove = async (handoffId: string) => {
    if (!confirm("Are you sure you want to approve this handoff? The patient will be transferred to you.")) {
      return;
    }

    setLoading(true);
    try {
      await patientService.approveHandoff(handoffId, {});
      toast({
        title: "Success",
        description: "Handoff approved successfully",
      });
      loadHandoffs();
      onHandoffUpdated?.();
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: { message?: string } } }; message?: string };
      const errorMessage = err.response?.data?.error?.message || err.message || "Failed to approve handoff";
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleReject = async (handoffId: string) => {
    const reason = prompt("Please provide a reason for rejecting this handoff:");
    if (reason === null) {
      return; // User cancelled
    }

    if (!reason.trim()) {
      toast({
        title: "Error",
        description: "Reason is required",
        variant: "destructive",
      });
      return;
    }

    setLoading(true);
    try {
      await patientService.rejectHandoff(handoffId, { reason: reason.trim() });
      toast({
        title: "Success",
        description: "Handoff rejected successfully",
      });
      loadHandoffs();
      onHandoffUpdated?.();
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: { message?: string } } }; message?: string };
      const errorMessage = err.response?.data?.error?.message || err.message || "Failed to reject handoff";
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = async (handoffId: string) => {
    if (!confirm("Are you sure you want to cancel this handoff request?")) {
      return;
    }

    setLoading(true);
    try {
      await patientService.cancelHandoff(handoffId);
      toast({
        title: "Success",
        description: "Handoff cancelled successfully",
      });
      loadHandoffs();
      onHandoffUpdated?.();
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: { message?: string } } }; message?: string };
      const errorMessage = err.response?.data?.error?.message || err.message || "Failed to cancel handoff";
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const getStatusBadge = (status: Handoff["status"]) => {
    switch (status) {
      case "requested":
        return <Badge variant="outline" className="gap-1"><Clock className="w-3 h-3" />Pending</Badge>;
      case "approved":
        return <Badge variant="default" className="gap-1 bg-green-600"><CheckCircle2 className="w-3 h-3" />Approved</Badge>;
      case "rejected":
        return <Badge variant="destructive" className="gap-1"><XCircle className="w-3 h-3" />Rejected</Badge>;
      case "cancelled":
        return <Badge variant="secondary" className="gap-1"><X className="w-3 h-3" />Cancelled</Badge>;
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  // Get available clinicians (not the current user)
  const availableClinicians = teamMembers.filter(
    (member) => member.user_id !== user?.id
  );

  // Check if user has a pending handoff request
  const hasPendingRequest = handoffs.some(
    (h) => h.requesting_clinician_id === user?.id && h.status === "requested"
  );

  if (initialLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Patient Handoffs</CardTitle>
          <CardDescription>
            Manage patient handoff requests between clinicians
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
            <CardTitle>Patient Handoffs</CardTitle>
            <CardDescription>
              Manage patient handoff requests between clinicians
            </CardDescription>
          </div>
          {!hasPendingRequest && (
            <Dialog open={isRequestDialogOpen} onOpenChange={setIsRequestDialogOpen}>
              <DialogTrigger asChild>
                <Button size="sm" className="gap-2">
                  <UserPlus className="w-4 h-4" />
                  Request Handoff
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Request Patient Handoff</DialogTitle>
                  <DialogDescription>
                    Request to transfer this patient to another clinician
                  </DialogDescription>
                </DialogHeader>
                <div className="space-y-4 py-4">
                  <div className="space-y-2">
                    <label className="text-sm font-medium">Receiving Clinician</label>
                    <Select value={selectedClinicianId} onValueChange={setSelectedClinicianId}>
                      <SelectTrigger>
                        <SelectValue placeholder="Select a clinician" />
                      </SelectTrigger>
                      <SelectContent>
                        {availableClinicians.length === 0 ? (
                          <div className="p-2 text-sm text-muted-foreground">
                            No available clinicians
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
                    <label className="text-sm font-medium">Role (Optional)</label>
                    <Select
                      value={selectedRole || "inherit"}
                      onValueChange={(value) => setSelectedRole(value === "inherit" ? undefined : value as "primary" | "secondary")}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Inherit from current role" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="inherit">Inherit from current role</SelectItem>
                        <SelectItem value="primary">Primary</SelectItem>
                        <SelectItem value="secondary">Secondary</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-2">
                    <label className="text-sm font-medium">Message (Optional)</label>
                    <Textarea
                      placeholder="Add a note about why you're requesting this handoff..."
                      value={message}
                      onChange={(e) => setMessage(e.target.value)}
                      rows={3}
                    />
                  </div>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setIsRequestDialogOpen(false)}>
                    Cancel
                  </Button>
                  <Button onClick={handleRequestHandoff} disabled={loading || !selectedClinicianId}>
                    {loading ? "Requesting..." : "Request Handoff"}
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          )}
        </div>
      </CardHeader>
      <CardContent>
        {handoffs.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            No handoff requests yet. Request a handoff to transfer this patient to another clinician.
          </div>
        ) : (
          <div className="space-y-3">
            {handoffs.map((handoff) => {
              const isRequesting = handoff.requesting_clinician_id === user?.id;
              const isReceiving = handoff.receiving_clinician_id === user?.id;
              const canApprove = isReceiving && handoff.status === "requested";
              const canReject = isReceiving && handoff.status === "requested";
              const canCancel = isRequesting && handoff.status === "requested";

              return (
                <div
                  key={handoff.id}
                  className="flex items-start justify-between p-4 rounded-lg border"
                >
                  <div className="flex-1 space-y-2">
                    <div className="flex items-center gap-2">
                      {getStatusBadge(handoff.status)}
                      <span className="text-sm text-muted-foreground">
                        {format(new Date(handoff.requested_at), "MMM d, yyyy 'at' h:mm a")}
                      </span>
                    </div>
                    <div className="flex items-center gap-2 text-sm">
                      <span className="font-medium">{handoff.requesting_clinician_name}</span>
                      <ArrowRight className="w-4 h-4 text-muted-foreground" />
                      <span className="font-medium">{handoff.receiving_clinician_name}</span>
                    </div>
                    {handoff.message && (
                      <p className="text-sm text-muted-foreground italic">
                        "{handoff.message}"
                      </p>
                    )}
                    {handoff.requested_role && (
                      <p className="text-xs text-muted-foreground">
                        Requested role: <span className="font-medium">{handoff.requested_role}</span>
                      </p>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    {canApprove && (
                      <Button
                        variant="default"
                        size="sm"
                        onClick={() => handleApprove(handoff.id)}
                        disabled={loading}
                        className="gap-1"
                      >
                        <CheckCircle2 className="w-4 h-4" />
                        Approve
                      </Button>
                    )}
                    {canReject && (
                      <Button
                        variant="destructive"
                        size="sm"
                        onClick={() => handleReject(handoff.id)}
                        disabled={loading}
                        className="gap-1"
                      >
                        <XCircle className="w-4 h-4" />
                        Reject
                      </Button>
                    )}
                    {canCancel && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleCancel(handoff.id)}
                        disabled={loading}
                        className="gap-1"
                      >
                        <X className="w-4 h-4" />
                        Cancel
                      </Button>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
};

export default PatientHandoff;

