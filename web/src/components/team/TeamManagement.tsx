import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
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
import { teamService, type TeamInvitation } from "@/services/teamService";
import { subscriptionService, UpgradePrompt as UpgradePromptType } from "@/services/subscriptionService";
import UpgradePrompt from "@/components/subscription/UpgradePrompt";
import { UserPlus, X, RotateCcw, Clock, CheckCircle2, XCircle } from "lucide-react";
import { format } from "date-fns";

const TeamManagement = () => {
  const { toast } = useToast();
  const [invitations, setInvitations] = useState<TeamInvitation[]>([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [pageSize] = useState(20);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState("member");
  const [sending, setSending] = useState(false);
  const [showUpgradePrompt, setShowUpgradePrompt] = useState(false);
  const [upgradePrompt, setUpgradePrompt] = useState<UpgradePromptType | null>(null);
  const [usageStats, setUsageStats] = useState<{ clinician_count: number; clinician_limit: number } | null>(null);

  useEffect(() => {
    loadInvitations();
    loadUsageStats();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page]);

  const loadUsageStats = async () => {
    try {
      const stats = await subscriptionService.getUsageStats();
      if (stats) {
        setUsageStats({
          clinician_count: stats.clinician_count,
          clinician_limit: stats.clinician_limit,
        });
      }
    } catch (error) {
      console.error("Failed to load usage stats", error);
    }
  };

  const loadInvitations = async () => {
    setLoading(true);
    try {
      const data = await teamService.listInvitations(page, pageSize);
      setInvitations(data.invitations);
      setTotal(data.total);
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: { message?: string } } }; message?: string };
      const message = err.response?.data?.error?.message || err.message || "Failed to load invitations";
      toast({
        title: "Error",
        description: message,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleSendInvitation = async () => {
    if (!inviteEmail.trim()) {
      toast({
        title: "Error",
        description: "Email is required",
        variant: "destructive",
      });
      return;
    }

    if (!inviteEmail.includes("@")) {
      toast({
        title: "Error",
        description: "Please enter a valid email address",
        variant: "destructive",
      });
      return;
    }

    setSending(true);
    try {
      await teamService.sendInvitation({
        email: inviteEmail,
        role: inviteRole,
      });
      toast({
        title: "Success",
        description: "Invitation sent successfully",
      });
      setIsDialogOpen(false);
      setInviteEmail("");
      setInviteRole("member");
      loadInvitations();
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: { details?: { upgrade_prompt?: UpgradePromptType }; message?: string } } }; message?: string };
      // Check for upgrade prompt in error response
      const errorDetails = err.response?.data?.error?.details;
      if (errorDetails?.upgrade_prompt) {
        setUpgradePrompt(errorDetails.upgrade_prompt);
        setShowUpgradePrompt(true);
      } else {
        const message = err.response?.data?.error?.message || err.message || "Failed to send invitation";
        toast({
          title: "Error",
          description: message,
          variant: "destructive",
        });
      }
    } finally {
      setSending(false);
    }
  };

  const handleCancelInvitation = async (invitationId: string) => {
    if (!confirm("Are you sure you want to cancel this invitation?")) {
      return;
    }

    try {
      await teamService.cancelInvitation(invitationId);
      toast({
        title: "Success",
        description: "Invitation cancelled",
      });
      loadInvitations();
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: { message?: string } } }; message?: string };
      const message = err.response?.data?.error?.message || err.message || "Failed to cancel invitation";
      toast({
        title: "Error",
        description: message,
        variant: "destructive",
      });
    }
  };

  const handleResendInvitation = async (invitationId: string) => {
    try {
      await teamService.resendInvitation(invitationId);
      toast({
        title: "Success",
        description: "Invitation resent successfully",
      });
      loadInvitations();
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: { message?: string } } }; message?: string };
      const message = err.response?.data?.error?.message || err.message || "Failed to resend invitation";
      toast({
        title: "Error",
        description: message,
        variant: "destructive",
      });
    }
  };

  const getStatusBadge = (status: string, expiresAt: string) => {
    const isExpired = new Date(expiresAt) < new Date();
    
    if (status === "accepted") {
      return (
        <Badge variant="default" className="bg-green-500">
          <CheckCircle2 className="w-3 h-3 mr-1" />
          Accepted
        </Badge>
      );
    }
    if (status === "cancelled") {
      return (
        <Badge variant="secondary">
          <XCircle className="w-3 h-3 mr-1" />
          Cancelled
        </Badge>
      );
    }
    if (status === "expired" || isExpired) {
      return (
        <Badge variant="destructive">
          <Clock className="w-3 h-3 mr-1" />
          Expired
        </Badge>
      );
    }
    return (
      <Badge variant="outline">
        <Clock className="w-3 h-3 mr-1" />
        Pending
      </Badge>
    );
  };

  const getRoleBadge = (role: string) => {
    const roleColors: Record<string, string> = {
      owner: "bg-purple-500",
      admin: "bg-blue-500",
      clinician: "bg-green-500",
      case_manager: "bg-orange-500",
      member: "bg-gray-500",
    };

    return (
      <Badge className={roleColors[role] || "bg-gray-500"}>
        {role.charAt(0).toUpperCase() + role.slice(1).replace("_", " ")}
      </Badge>
    );
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Invitations</CardTitle>
              <CardDescription>
                Invite team members and manage their access to your organization
              </CardDescription>
            </div>
            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
              <DialogTrigger asChild>
                <Button 
                  className="gap-2"
                  disabled={usageStats && subscriptionService.isAtLimit(usageStats.clinician_count, usageStats.clinician_limit)}
                >
                  <UserPlus className="w-4 h-4" />
                  Invite Member
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Send Team Invitation</DialogTitle>
                  <DialogDescription>
                    Send an invitation email to a team member. They will receive a link to join your organization.
                  </DialogDescription>
                </DialogHeader>
                <div className="space-y-4 py-4">
                  <div className="space-y-2">
                    <Label htmlFor="invite-email">Email Address</Label>
                    <Input
                      id="invite-email"
                      type="email"
                      placeholder="colleague@example.com"
                      value={inviteEmail}
                      onChange={(e) => setInviteEmail(e.target.value)}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="invite-role">Role</Label>
                    <Select value={inviteRole} onValueChange={setInviteRole}>
                      <SelectTrigger id="invite-role">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="admin">Admin - Full access to all features</SelectItem>
                        <SelectItem value="clinician">Clinician - Clinical operations access</SelectItem>
                        <SelectItem value="case_manager">Case Manager - Case management access</SelectItem>
                        <SelectItem value="member">Member - Read-only access</SelectItem>
                      </SelectContent>
                    </Select>
                    <p className="text-xs text-muted-foreground">
                      {inviteRole === "admin" && "Can manage organization, audit logs, and all data"}
                      {inviteRole === "clinician" && "Can create/edit patients, appointments, and clinical notes"}
                      {inviteRole === "case_manager" && "Can manage patient cases and coordination"}
                      {inviteRole === "member" && "Can view data but cannot create or edit"}
                    </p>
                  </div>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                    Cancel
                  </Button>
                  <Button onClick={handleSendInvitation} disabled={sending}>
                    {sending ? "Sending..." : "Send Invitation"}
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>
        </CardHeader>
        <CardContent>
          {(() => {
            if (loading) {
              return <div className="text-center py-8 text-muted-foreground">Loading invitations...</div>;
            }
            if (invitations.length === 0) {
              return (
                <div className="text-center py-8 text-muted-foreground">
                  No invitations yet. Invite your first team member to get started.
                </div>
              );
            }
            return (
              <>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Email</TableHead>
                      <TableHead>Role</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Expires</TableHead>
                      <TableHead>Sent</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {invitations.map((invitation) => (
                      <TableRow key={invitation.id}>
                        <TableCell className="font-medium">{invitation.email}</TableCell>
                        <TableCell>{getRoleBadge(invitation.role)}</TableCell>
                        <TableCell>{getStatusBadge(invitation.status, invitation.expires_at)}</TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {format(new Date(invitation.expires_at), "MMM d, yyyy")}
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {format(new Date(invitation.created_at), "MMM d, yyyy")}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex justify-end gap-2">
                            {invitation.status === "pending" && (
                              <>
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => handleResendInvitation(invitation.id)}
                                  className="gap-1"
                                >
                                  <RotateCcw className="w-3 h-3" />
                                  Resend
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => handleCancelInvitation(invitation.id)}
                                  className="gap-1 text-destructive"
                                >
                                  <X className="w-3 h-3" />
                                  Cancel
                                </Button>
                              </>
                            )}
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
                {total > pageSize && (
                  <div className="flex items-center justify-between mt-4">
                    <div className="text-sm text-muted-foreground">
                      Showing {(page - 1) * pageSize + 1} to {Math.min(page * pageSize, total)} of {total} invitations
                    </div>
                    <div className="flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setPage((p) => Math.max(1, p - 1))}
                        disabled={page === 1}
                      >
                        Previous
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setPage((p) => p + 1)}
                        disabled={page * pageSize >= total}
                      >
                        Next
                      </Button>
                    </div>
                  </div>
                )}
              </>
            );
          })()}
        </CardContent>
      </Card>
      <UpgradePrompt
        isOpen={showUpgradePrompt}
        onClose={() => setShowUpgradePrompt(false)}
        upgradePrompt={upgradePrompt}
      />
    </div>
  );
};

export default TeamManagement;

