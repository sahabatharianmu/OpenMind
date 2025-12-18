import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import DashboardLayout from "@/components/dashboard/DashboardLayout";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { 
  Plus, 
  Search, 
  FileText, 
  Clock, 
  CheckCircle2
} from "lucide-react";
import clinicalNoteService from "@/services/clinicalNoteService";
import patientService from "@/services/patientService";
import { useAuth } from "@/contexts/AuthContext";
import { format } from "date-fns";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ClinicalNote, Patient } from "@/types";

interface UIClinicalNote extends ClinicalNote {
  patient?: Patient;
}

const Notes = () => {
  const navigate = useNavigate();
  const { user } = useAuth();
  const [notes, setNotes] = useState<UIClinicalNote[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [filterType, setFilterType] = useState<string>("all");
  const [filterStatus, setFilterStatus] = useState<string>("all");

  useEffect(() => {
    fetchData();
  }, [user]);

  const fetchData = async () => {
    if (!user) return;

    setLoading(true);
    try {
      const [notesData, patientsData] = await Promise.all([
        clinicalNoteService.list(),
        patientService.list()
      ]);

      const allPatients = patientsData || [];
      const allNotes = notesData || [];

      const enrichedNotes = allNotes.map(note => ({
        ...note,
        patient: allPatients.find(p => p.id === note.patient_id)
      }));

      // Sort by created_at desc
      enrichedNotes.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());

      setNotes(enrichedNotes);
    } catch (error) {
      console.error("Error fetching notes:", error);
    } finally {
      setLoading(false);
    }
  };

  const filteredNotes = notes.filter((note) => {
    const patientName = `${note.patient?.first_name} ${note.patient?.last_name}`.toLowerCase();
    const matchesSearch = patientName.includes(search.toLowerCase());
    const matchesType = filterType === "all" || note.note_type === filterType;
    const matchesStatus = filterStatus === "all" || 
      (filterStatus === "signed" && note.is_signed) ||
      (filterStatus === "unsigned" && !note.is_signed);
    
    return matchesSearch && matchesType && matchesStatus;
  });

  return (
    <DashboardLayout>
      <div className="p-6 lg:p-8">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
          <div>
            <h1 className="text-2xl lg:text-3xl font-bold">Clinical Notes</h1>
            <p className="text-muted-foreground mt-1">
              Manage and review your clinical documentation
            </p>
          </div>
          <Button className="gap-2" onClick={() => navigate("/dashboard/notes/new")}>
            <Plus className="w-4 h-4" />
            New Note
          </Button>
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
              <Select value={filterType} onValueChange={setFilterType}>
                <SelectTrigger className="w-full sm:w-40">
                  <SelectValue placeholder="Note Type" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Types</SelectItem>
                  <SelectItem value="soap">SOAP</SelectItem>
                  <SelectItem value="progress">Progress</SelectItem>
                  <SelectItem value="intake">Intake</SelectItem>
                  <SelectItem value="discharge">Discharge</SelectItem>
                </SelectContent>
              </Select>
              <Select value={filterStatus} onValueChange={setFilterStatus}>
                <SelectTrigger className="w-full sm:w-40">
                  <SelectValue placeholder="Status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Status</SelectItem>
                  <SelectItem value="signed">Signed</SelectItem>
                  <SelectItem value="unsigned">Unsigned</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>

        {/* Notes List */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">
              {filteredNotes.length} Note{filteredNotes.length !== 1 ? "s" : ""}
            </CardTitle>
          </CardHeader>
          <CardContent>
            {loading ? (
              <div className="text-center py-8 text-muted-foreground">
                Loading notes...
              </div>
            ) : filteredNotes.length === 0 ? (
              <div className="text-center py-12 text-muted-foreground">
                <FileText className="w-12 h-12 mx-auto mb-3 opacity-50" />
                <p className="font-medium">No notes found</p>
                <p className="text-sm mt-1">
                  {search || filterType !== "all" || filterStatus !== "all"
                    ? "Try adjusting your filters"
                    : "Create your first clinical note to get started"}
                </p>
                {!search && filterType === "all" && filterStatus === "all" && (
                  <Button 
                    variant="outline" 
                    className="mt-4 gap-2"
                    onClick={() => navigate("/dashboard/notes/new")}
                  >
                    <Plus className="w-4 h-4" />
                    Create Note
                  </Button>
                )}
              </div>
            ) : (
              <div className="space-y-3">
                {filteredNotes.map((note) => (
                  <div
                    key={note.id}
                    className="flex items-center justify-between p-4 rounded-lg border border-border hover:border-primary/50 transition-colors cursor-pointer"
                    onClick={() => navigate(`/dashboard/notes/${note.id}`)}
                  >
                    <div className="flex items-center gap-4">
                      <div className={`p-2 rounded-lg ${note.is_signed ? "bg-green-500/10" : "bg-muted"}`}>
                        {note.is_signed ? (
                          <CheckCircle2 className="w-5 h-5 text-green-600" />
                        ) : (
                          <Clock className="w-5 h-5 text-muted-foreground" />
                        )}
                      </div>
                      <div>
                        <p className="font-medium">
                          {note.patient?.first_name} {note.patient?.last_name}
                        </p>
                        <div className="flex items-center gap-2 mt-1">
                          <Badge variant="outline" className="text-xs capitalize">
                            {note.note_type}
                          </Badge>
                          <span className="text-xs text-muted-foreground">
                            {format(new Date(note.created_at), "MMM d, yyyy")}
                          </span>
                        </div>
                      </div>
                    </div>
                    <div className="text-right">
                      <Badge variant={note.is_signed ? "default" : "secondary"}>
                        {note.is_signed ? "Signed" : "Draft"}
                      </Badge>
                      {note.is_signed && note.signed_at && (
                        <p className="text-xs text-muted-foreground mt-1">
                          Signed {format(new Date(note.signed_at), "MMM d")}
                        </p>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </DashboardLayout>
  );
};

export default Notes;
