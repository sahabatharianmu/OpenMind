export interface User {
  id: string;
  email: string;
  full_name: string;
  role: string;
  created_at: string;
  updated_at: string;
}

export interface Patient {
  id: string;
  organization_id: string;
  first_name: string;
  last_name: string;
  date_of_birth: string;
  email?: string;
  phone?: string;
  address?: string;
  status: "active" | "inactive" | "archived";
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface Appointment {
  id: string;
  organization_id: string;
  patient_id: string;
  clinician_id: string;
  clinician_name?: string;
  clinician_email?: string;
  start_time: string;
  end_time: string;
  status: "scheduled" | "completed" | "cancelled" | "no-show";
  appointment_type: string;
  mode: "in-person" | "video" | "phone";
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface Addendum {
  id: string;
  clinician_id: string;
  content: string;
  signed_at: string;
}

export interface ClinicalNote {
  id: string;
  organization_id: string;
  patient_id: string;
  clinician_id: string;
  appointment_id?: string;
  note_type: string;
  subjective?: string;
  objective?: string;
  assessment?: string;
  plan?: string;
  is_signed: boolean;
  signed_at?: string;
  addendums?: Addendum[];
  created_at: string;
  updated_at: string;
}

export interface Invoice {
  id: string;
  organization_id: string;
  patient_id: string;
  appointment_id?: string;
  amount_cents: number;
  status: "draft" | "sent" | "pending" | "paid" | "void" | "overdue" | "cancelled";
  due_date?: string;
  paid_at?: string;
  payment_method?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}
