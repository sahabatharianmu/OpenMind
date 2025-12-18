import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import DashboardLayout from "@/components/dashboard/DashboardLayout";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { 
  ArrowLeft, 
  Save, 
  CheckCircle2, 
  Lock,
  FileText,
  Plus,
  Shield
} from "lucide-react";
import clinicalNoteService from "@/services/clinicalNoteService";
import patientService from "@/services/patientService";
import { useAuth } from "@/contexts/AuthContext";
import { useToast } from "@/hooks/use-toast";
import { format } from "date-fns";
import { Patient } from "@/types";

const NoteEditor = () => {
  const navigate = useNavigate();
  const { id } = useParams();
  const { user } = useAuth();
  const { toast } = useToast();
  const isNew = !id || id === "new";

  const [loading, setLoading] = useState(!isNew);
  const [saving, setSaving] = useState(false);
  const [patients, setPatients] = useState<Patient[]>([]);
  
  // Form state
  const [patientId, setPatientId] = useState("");
  const [noteType, setNoteType] = useState("soap");
  const [subjective, setSubjective] = useState("");
  const [objective, setObjective] = useState("");
  const [assessment, setAssessment] = useState("");
  const [plan, setPlan] = useState("");
  const [isSigned, setIsSigned] = useState(false);
  const [signedAt, setSignedAt] = useState<string | null>(null);
  const [createdAt, setCreatedAt] = useState<string | null>(null);
  const [addendums, setAddendums] = useState<any[]>([]);
  const [newAddendum, setNewAddendum] = useState("");
  const [addingAddendum, setAddingAddendum] = useState(false);

  useEffect(() => {
    fetchPatients();
    if (!isNew && id) {
      fetchNote(id);
    }
  }, [id, isNew]);

  const fetchPatients = async () => {
    try {
      const data = await patientService.list();
      setPatients(data?.filter(p => p.status === 'active') || []);
    } catch (error) {
      console.error("Error fetching patients", error);
    }
  };

  const fetchNote = async (noteId: string) => {
    setLoading(true);
    try {
      const data = await clinicalNoteService.get(noteId);
      if (data) {
        setPatientId(data.patient_id);
        setNoteType(data.note_type);
        setSubjective(data.subjective || "");
        setObjective(data.objective || "");
        setAssessment(data.assessment || "");
        setPlan(data.plan || "");
        setIsSigned(data.is_signed);
        setSignedAt(data.signed_at || null);
        setCreatedAt(data.created_at);
        setAddendums(data.addendums || []);
      }
    } catch (error) {
      toast({
        title: "Error",
        description: "Could not load the note.",
        variant: "destructive",
      });
      navigate("/dashboard/notes");
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async (sign = false) => {
    if (!user || !patientId) {
      toast({
        title: "Error",
        description: "Please select a patient.",
        variant: "destructive",
      });
      return;
    }

    setSaving(true);
    
    // Construct the payload matching the Create/Update/Service Request types
    const noteData = {
      patient_id: patientId,
      clinician_id: user.id, // Service might pull this from context/token but good to verify
      note_type: noteType,
      subjective,
      objective,
      assessment,
      plan,
      is_signed: sign,
      // signed_at is handled by backend usually if is_signed is true, or we pass it? 
      // Checking service definition, UpdateClinicalNoteRequest takes fields.
      // Backend should set signed_at if is_signed becomes true.
    };

    try {
      if (isNew) {
        await clinicalNoteService.create(noteData);
      } else if (id) {
        await clinicalNoteService.update(id, noteData);
      }

      toast({
        title: sign ? "Note Signed" : "Note Saved",
        description: sign 
          ? "Your clinical note has been signed and locked."
          : "Your changes have been saved.",
      });

      if (sign) {
        setIsSigned(true);
        setSignedAt(new Date().toISOString());
      }
      
      if (isNew) {
        navigate("/dashboard/notes");
      }
    } catch (error: any) {
      console.error("Error saving note:", error);
       toast({
        title: "Error",
        description: "Failed to save the note. " + (error.message || ""),
        variant: "destructive",
      });
    } finally {
      setSaving(false);
    }
  };

  const handleAddAddendum = async () => {
    if (!id || !newAddendum.trim()) return;

    setAddingAddendum(true);
    try {
      const data = await clinicalNoteService.addAddendum(id, newAddendum);
      setAddendums([...addendums, data]);
      setNewAddendum("");
      toast({
        title: "Addendum Added",
        description: "The addendum has been successfully added to this note.",
      });
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to add addendum.",
        variant: "destructive",
      });
    } finally {
      setAddingAddendum(false);
    }
  };

  if (loading) {
    return (
      <DashboardLayout>
        <div className="p-6 lg:p-8">
          <div className="animate-pulse">Loading note...</div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="p-6 lg:p-8 max-w-4xl">
        {/* Header */}
        <div className="flex items-center gap-4 mb-6">
          <Button 
            variant="ghost" 
            size="icon"
            onClick={() => navigate("/dashboard/notes")}
          >
            <ArrowLeft className="w-5 h-5" />
          </Button>
          <div className="flex-1">
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold">
                {isNew ? "New Clinical Note" : "Edit Note"}
              </h1>
              <Badge variant="outline" className="gap-1 bg-green-50 text-green-700 border-green-200">
                <Shield className="w-3 h-3" />
                AES-256 Encrypted
              </Badge>
              {isSigned && (
                <Badge className="gap-1">
                  <Lock className="w-3 h-3" />
                  Signed
                </Badge>
              )}
            </div>
            {createdAt && (
              <p className="text-sm text-muted-foreground mt-1">
                Created {format(new Date(createdAt), "MMMM d, yyyy 'at' h:mm a")}
              </p>
            )}
          </div>
        </div>

        {isSigned && (
          <Card className="mb-6 border-primary/50 bg-primary/5">
            <CardContent className="p-4 flex items-center gap-3">
              <CheckCircle2 className="w-5 h-5 text-primary" />
              <div>
                <p className="font-medium">This note has been signed and locked</p>
                <p className="text-sm text-muted-foreground">
                  Signed on {signedAt && format(new Date(signedAt), "MMMM d, yyyy 'at' h:mm a")}
                </p>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Patient & Type Selection */}
        <Card className="mb-6">
          <CardHeader>
            <CardTitle className="text-lg">Note Details</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Patient</Label>
                <Select 
                  value={patientId} 
                  onValueChange={setPatientId}
                  disabled={isSigned || !isNew}
                >
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
              <div className="space-y-2">
                <Label>Note Type</Label>
                <Select 
                  value={noteType} 
                  onValueChange={setNoteType}
                  disabled={isSigned}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="soap">SOAP Note</SelectItem>
                    <SelectItem value="progress">Progress Note</SelectItem>
                    <SelectItem value="intake">Intake Assessment</SelectItem>
                    <SelectItem value="discharge">Discharge Summary</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* SOAP Fields */}
        <Card className="mb-6">
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <FileText className="w-5 h-5" />
              SOAP Documentation
            </CardTitle>
            <CardDescription>
              Document your clinical observations and treatment plan
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="space-y-2">
              <Label htmlFor="subjective" className="text-base font-semibold">
                Subjective
              </Label>
              <p className="text-sm text-muted-foreground">
                Patient's reported symptoms, feelings, and concerns
              </p>
              <Textarea
                id="subjective"
                placeholder="Document the patient's subjective experience..."
                value={subjective}
                onChange={(e) => setSubjective(e.target.value)}
                disabled={isSigned}
                className="min-h-[120px]"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="objective" className="text-base font-semibold">
                Objective
              </Label>
              <p className="text-sm text-muted-foreground">
                Clinical observations, mental status exam findings
              </p>
              <Textarea
                id="objective"
                placeholder="Document your clinical observations..."
                value={objective}
                onChange={(e) => setObjective(e.target.value)}
                disabled={isSigned}
                className="min-h-[120px]"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="assessment" className="text-base font-semibold">
                Assessment
              </Label>
              <p className="text-sm text-muted-foreground">
                Clinical interpretation, diagnosis, and progress evaluation
              </p>
              <Textarea
                id="assessment"
                placeholder="Document your clinical assessment..."
                value={assessment}
                onChange={(e) => setAssessment(e.target.value)}
                disabled={isSigned}
                className="min-h-[120px]"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="plan" className="text-base font-semibold">
                Plan
              </Label>
              <p className="text-sm text-muted-foreground">
                Treatment plan, interventions, and next steps
              </p>
              <Textarea
                id="plan"
                placeholder="Document the treatment plan..."
                value={plan}
                onChange={(e) => setPlan(e.target.value)}
                disabled={isSigned}
                className="min-h-[120px]"
              />
            </div>
          </CardContent>
        </Card>

        {/* Addendums Section */}
        {isSigned && (
          <div className="space-y-6 mb-6">
            <h3 className="text-xl font-bold flex items-center gap-2">
              <Plus className="w-5 h-5" />
              Addendums
            </h3>
            
            {addendums.length > 0 ? (
              <div className="space-y-4">
                {addendums.map((addendum, index) => (
                  <Card key={addendum.id || index} className="border-l-4 border-l-primary">
                    <CardHeader className="py-3">
                      <div className="flex justify-between items-center">
                        <CardTitle className="text-sm font-medium">
                          Addendum #{index + 1}
                        </CardTitle>
                        <span className="text-xs text-muted-foreground">
                          {format(new Date(addendum.signed_at), "MMMM d, yyyy 'at' h:mm a")}
                        </span>
                      </div>
                    </CardHeader>
                    <CardContent>
                      <p className="text-sm whitespace-pre-wrap">{addendum.content}</p>
                    </CardContent>
                  </Card>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground italic">No addendums yet.</p>
            )}

            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Add New Addendum</CardTitle>
                <CardDescription>
                  Enter additional information or corrections for this signed note.
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <Textarea
                  placeholder="Enter addendum content..."
                  value={newAddendum}
                  onChange={(e) => setNewAddendum(e.target.value)}
                  className="min-h-[100px]"
                />
                <div className="flex justify-end">
                  <Button 
                    onClick={handleAddAddendum} 
                    disabled={addingAddendum || !newAddendum.trim()}
                    className="gap-2"
                  >
                    <Save className="w-4 h-4" />
                    {addingAddendum ? "Adding..." : "Save Addendum"}
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        )}

        {/* Actions */}
        {!isSigned && (
          <div className="flex items-center justify-end gap-3">
            <Button
              variant="outline"
              onClick={() => navigate("/dashboard/notes")}
            >
              Cancel
            </Button>
            <Button
              variant="secondary"
              onClick={() => handleSave(false)}
              disabled={saving}
              className="gap-2"
            >
              <Save className="w-4 h-4" />
              Save Draft
            </Button>
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button disabled={saving || !patientId} className="gap-2">
                  <CheckCircle2 className="w-4 h-4" />
                  Sign & Lock
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Sign Clinical Note</AlertDialogTitle>
                  <AlertDialogDescription>
                    Once signed, this note will be locked and cannot be edited.
                    Any changes will need to be added as an addendum.
                    Are you sure you want to sign this note?
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction onClick={() => handleSave(true)}>
                    Sign Note
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        )}
      </div>
    </DashboardLayout>
  );
};

export default NoteEditor;
