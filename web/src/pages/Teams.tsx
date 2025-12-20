import { useState, useEffect } from "react";
import DashboardLayout from "@/components/dashboard/DashboardLayout";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
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
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Label } from "@/components/ui/label";
import { useToast } from "@/hooks/use-toast";
import { useAuth } from "@/contexts/AuthContext";
import { organizationService, type TeamMember } from "@/services/organizationService";
import { Edit, Trash2 } from "lucide-react";
import { format } from "date-fns";
import TeamManagement from "@/components/team/TeamManagement";

const Teams = () => {
  const { user } = useAuth();
  const { toast } = useToast();
  const [members, setMembers] = useState<TeamMember[]>([]);
  const [loading, setLoading] = useState(false);
  const [editingMember, setEditingMember] = useState<TeamMember | null>(null);
  const [newRole, setNewRole] = useState("");

  const loadData = async () => {
    setLoading(true);
    try {
      const membersData = await organizationService.listTeamMembers();
      setMembers(membersData);
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: { message?: string } } }; message?: string };
      const message = err.response?.data?.error?.message || err.message || "Failed to load team data";
      toast({
        title: "Error",
        description: message,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleUpdateRole = async () => {
    if (!editingMember || !newRole) return;

    try {
      await organizationService.updateMemberRole(editingMember.user_id, newRole);
      toast({
        title: "Success",
        description: "Member role updated successfully",
      });
      setEditingMember(null);
      setNewRole("");
      loadData();
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: { message?: string } } }; message?: string };
      const message = err.response?.data?.error?.message || err.message || "Failed to update role";
      toast({
        title: "Error",
        description: message,
        variant: "destructive",
      });
    }
  };

  const handleRemoveMember = async (member: TeamMember) => {
    if (!confirm(`Are you sure you want to remove ${member.full_name} from the team?`)) {
      return;
    }

    try {
      await organizationService.removeMember(member.user_id);
      toast({
        title: "Success",
        description: "Member removed successfully",
      });
      loadData();
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: { message?: string } } }; message?: string };
      const message = err.response?.data?.error?.message || err.message || "Failed to remove member";
      toast({
        title: "Error",
        description: message,
        variant: "destructive",
      });
    }
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

  const canEdit = user?.role === "admin" || user?.role === "owner";
  const isOwner = (member: TeamMember) => member.role === "owner";

  return (
    <DashboardLayout>
      <div className="p-6 lg:p-8 max-w-6xl">
        <div className="mb-6">
          <h1 className="text-2xl lg:text-3xl font-bold">Team Management</h1>
          <p className="text-muted-foreground mt-1">
            Manage your team members, roles, and invitations
          </p>
        </div>

        <div className="space-y-6">
          {/* Current Members */}
          <Card>
            <CardHeader>
              <CardTitle>Team Members</CardTitle>
              <CardDescription>
                Current members of your organization
              </CardDescription>
            </CardHeader>
            <CardContent>
              {(() => {
                if (loading) {
                  return <div className="text-center py-8 text-muted-foreground">Loading members...</div>;
                }
                if (members.length === 0) {
                  return (
                    <div className="text-center py-8 text-muted-foreground">
                      No team members yet.
                    </div>
                  );
                }
                return (
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Name</TableHead>
                        <TableHead>Email</TableHead>
                        <TableHead>Role</TableHead>
                        <TableHead>Joined</TableHead>
                        {canEdit && <TableHead className="text-right">Actions</TableHead>}
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {members.map((member) => (
                        <TableRow key={member.user_id}>
                          <TableCell className="font-medium">{member.full_name}</TableCell>
                          <TableCell>{member.email}</TableCell>
                          <TableCell>{getRoleBadge(member.role)}</TableCell>
                          <TableCell className="text-sm text-muted-foreground">
                            {format(new Date(member.joined_at), "MMM d, yyyy")}
                          </TableCell>
                          {canEdit && (
                            <TableCell className="text-right">
                              <div className="flex justify-end gap-2">
                                {!isOwner(member) && (
                                  <>
                                    <Button
                                      variant="ghost"
                                      size="sm"
                                      onClick={() => {
                                        setEditingMember(member);
                                        setNewRole(member.role);
                                      }}
                                      className="gap-1"
                                    >
                                      <Edit className="w-3 h-3" />
                                      Edit Role
                                    </Button>
                                    {member.user_id !== user?.id && (
                                      <Button
                                        variant="ghost"
                                        size="sm"
                                        onClick={() => handleRemoveMember(member)}
                                        className="gap-1 text-destructive"
                                      >
                                        <Trash2 className="w-3 h-3" />
                                        Remove
                                      </Button>
                                    )}
                                  </>
                                )}
                              </div>
                            </TableCell>
                          )}
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                );
              })()}
            </CardContent>
          </Card>

          {/* Invitations */}
          {canEdit && <TeamManagement />}
        </div>

        {/* Edit Role Dialog */}
        <Dialog open={!!editingMember} onOpenChange={(open) => !open && setEditingMember(null)}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Update Member Role</DialogTitle>
              <DialogDescription>
                Change the role for {editingMember?.full_name}
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="role">Role</Label>
                <Select value={newRole} onValueChange={setNewRole}>
                  <SelectTrigger id="role">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="admin">Admin - Full access to all features</SelectItem>
                    <SelectItem value="clinician">Clinician - Clinical operations access</SelectItem>
                    <SelectItem value="case_manager">Case Manager - Case management access</SelectItem>
                    <SelectItem value="member">Member - Read-only access</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setEditingMember(null)}>
                Cancel
              </Button>
              <Button onClick={handleUpdateRole}>Update Role</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </DashboardLayout>
  );
};

export default Teams;

