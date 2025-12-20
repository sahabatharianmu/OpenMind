import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import DashboardLayout from "@/components/dashboard/DashboardLayout";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { 
  Plus, 
  Search, 
  Users, 
  MoreHorizontal,
  Mail,
  Phone,
  Calendar,
  FileText,
  Receipt
} from "lucide-react";
import patientService from "@/services/patientService";
import { useAuth } from "@/contexts/AuthContext";
import { useToast } from "@/hooks/use-toast";
import { format } from "date-fns";
import { Patient } from "@/types";

const Patients = () => {
  const navigate = useNavigate();
  const { user } = useAuth();
  const { toast } = useToast();
  const [patients, setPatients] = useState<Patient[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [assignedPatientIds, setAssignedPatientIds] = useState<Set<string>>(new Set());

  // New patient form
  const [newFirstName, setNewFirstName] = useState("");
  const [newLastName, setNewLastName] = useState("");
  const [newDob, setNewDob] = useState("");
  const [newEmail, setNewEmail] = useState("");
  const [newPhone, setNewPhone] = useState("");
  const [newAddress, setNewAddress] = useState("");

  useEffect(() => {
    fetchPatients();
  }, [user]);

  const fetchPatients = async () => {
    if (!user) return; // Wait for user to be loaded

    setLoading(true);
    try {
      const data = await patientService.list();
      setPatients(data || []);
      
      // Check assignment status for each patient
      const assignedIds = new Set<string>();
      for (const patient of data || []) {
        try {
          const isAssigned = await patientService.isAssigned(patient.id, user?.id);
          if (isAssigned) {
            assignedIds.add(patient.id);
          }
        } catch (error) {
          console.error(`Error checking assignment for patient ${patient.id}:`, error);
        }
      }
      setAssignedPatientIds(assignedIds);
    } catch (error) {
      console.error("Error fetching patients:", error);
      toast({
        title: "Error",
        description: "Failed to load patients.",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleAddPatient = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!user) return;

    setIsSubmitting(true);

    try {
      const newPatient = await patientService.create({
        first_name: newFirstName,
        last_name: newLastName,
        date_of_birth: newDob,
        email: newEmail || undefined,
        phone: newPhone || undefined,
        address: newAddress || undefined,
      });

      toast({
        title: "Patient Added",
        description: `${newFirstName} ${newLastName} has been added.`,
      });
      setIsAddDialogOpen(false);
      resetForm();
      fetchPatients();
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to add patient.",
        variant: "destructive",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const resetForm = () => {
    setNewFirstName("");
    setNewLastName("");
    setNewDob("");
    setNewEmail("");
    setNewPhone("");
    setNewAddress("");
  };

  const filteredPatients = patients.filter((p) =>
    `${p.first_name} ${p.last_name}`.toLowerCase().includes(search.toLowerCase())
  );

  const activeCount = patients.filter(p => p.status === "active").length;

  return (
    <DashboardLayout>
      <div className="p-4 sm:p-6 lg:p-8">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-4 sm:mb-6">
          <div>
            <h1 className="text-xl sm:text-2xl lg:text-3xl font-bold">Patients</h1>
            <p className="text-muted-foreground mt-1 text-sm sm:text-base">
              {activeCount} active patient{activeCount !== 1 ? "s" : ""} in your practice
            </p>
          </div>
          <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
            <DialogTrigger asChild>
              <Button className="gap-2 h-11 min-h-[44px] w-full sm:w-auto">
                <Plus className="w-4 h-4" />
                Add Patient
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-md">
              <DialogHeader>
                <DialogTitle>Add New Patient</DialogTitle>
                <DialogDescription>
                  Enter the patient's information to add them to your practice.
                </DialogDescription>
              </DialogHeader>
              <form onSubmit={handleAddPatient} className="space-y-4 mt-4">
                <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                    <Label htmlFor="firstName">First Name</Label>
                    <Input
                      id="firstName"
                      value={newFirstName}
                      onChange={(e) => setNewFirstName(e.target.value)}
                      required
                      className="h-11 min-h-[44px]"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="lastName">Last Name</Label>
                    <Input
                      id="lastName"
                      value={newLastName}
                      onChange={(e) => setNewLastName(e.target.value)}
                      required
                      className="h-11 min-h-[44px]"
                    />
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="dob">Date of Birth</Label>
                  <Input
                    id="dob"
                    type="date"
                    value={newDob}
                    onChange={(e) => setNewDob(e.target.value)}
                    required
                    className="h-11 min-h-[44px]"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    id="email"
                    type="email"
                    value={newEmail}
                    onChange={(e) => setNewEmail(e.target.value)}
                    className="h-11 min-h-[44px]"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="phone">Phone</Label>
                  <Input
                    id="phone"
                    type="tel"
                    value={newPhone}
                    onChange={(e) => setNewPhone(e.target.value)}
                    className="h-11 min-h-[44px]"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="address">Address</Label>
                  <Input
                    id="address"
                    value={newAddress}
                    onChange={(e) => setNewAddress(e.target.value)}
                    className="h-11 min-h-[44px]"
                  />
                </div>
                <Button type="submit" className="w-full h-11 min-h-[44px]" disabled={isSubmitting}>
                  {isSubmitting ? "Adding..." : "Add Patient"}
                </Button>
              </form>
            </DialogContent>
          </Dialog>
        </div>

        {/* Search */}
        <Card className="mb-4 sm:mb-6">
          <CardContent className="p-4">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                placeholder="Search patients by name..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-9 h-11 min-h-[44px]"
              />
            </div>
          </CardContent>
        </Card>

        {/* Patients Grid */}
        {loading ? (
          <div className="text-center py-12 text-muted-foreground">
            Loading patients...
          </div>
        ) : filteredPatients.length === 0 ? (
          <Card>
            <CardContent className="text-center py-12">
              <Users className="w-12 h-12 mx-auto mb-3 text-muted-foreground opacity-50" />
              <p className="font-medium">
                {search ? "No patients found" : "No patients yet"}
              </p>
              <p className="text-sm text-muted-foreground mt-1">
                {search ? "Try a different search term" : "Add your first patient to get started"}
              </p>
            </CardContent>
          </Card>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredPatients.map((patient) => (
              <Card key={patient.id} className="hover:shadow-md transition-shadow cursor-pointer" onClick={() => navigate(`/dashboard/patients/${patient.id}`)}>
                <CardContent className="p-4">
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-3">
                      <Avatar className="h-12 w-12">
                        <AvatarFallback className="bg-primary/10 text-primary">
                          {patient.first_name[0]}{patient.last_name[0]}
                        </AvatarFallback>
                      </Avatar>
                      <div>
                        <div className="flex items-center gap-2">
                          <p className="font-semibold">
                            {patient.first_name} {patient.last_name}
                          </p>
                          {assignedPatientIds.has(patient.id) && (
                            <Badge variant="outline" className="text-xs">
                              Assigned
                            </Badge>
                          )}
                        </div>
                        <Badge 
                          variant={patient.status === "active" ? "default" : "secondary"}
                          className="mt-1"
                        >
                          {patient.status}
                        </Badge>
                      </div>
                    </div>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-10 w-10 min-h-[44px] min-w-[44px]">
                          <MoreHorizontal className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        {assignedPatientIds.has(patient.id) && (
                          <DropdownMenuItem onClick={(e) => { e.stopPropagation(); navigate(`/dashboard/notes/new?patient=${patient.id}`); }}>
                            <FileText className="w-4 h-4 mr-2" />
                            Add Note
                          </DropdownMenuItem>
                        )}
                        <DropdownMenuItem onClick={(e) => { e.stopPropagation(); navigate(`/dashboard/appointments/new?patient=${patient.id}`); }}>
                          <Calendar className="w-4 h-4 mr-2" />
                          Schedule Appointment
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={(e) => { e.stopPropagation(); navigate(`/dashboard/invoices/new?patient=${patient.id}`); }}>
                          <Receipt className="w-4 h-4 mr-2" />
                          Create Invoice
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem onClick={(e) => { e.stopPropagation(); navigate(`/dashboard/patients/${patient.id}`); }}>
                          View Full Profile
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>

                  <div className="mt-4 space-y-2 text-sm text-muted-foreground">
                    <div className="flex items-center gap-2">
                      <Calendar className="w-4 h-4" />
                      <span>DOB: {format(new Date(patient.date_of_birth), "MMM d, yyyy")}</span>
                    </div>
                    {patient.email && (
                      <div className="flex items-center gap-2">
                        <Mail className="w-4 h-4" />
                        <span className="truncate">{patient.email}</span>
                      </div>
                    )}
                    {patient.phone && (
                      <div className="flex items-center gap-2">
                        <Phone className="w-4 h-4" />
                        <span>{patient.phone}</span>
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>
    </DashboardLayout>
  );
};

export default Patients;
