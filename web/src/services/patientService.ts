import api from "@/api/client";
import { Patient } from "@/types";
import type { PaginatedResponse } from "@/types/api";

export interface CreatePatientRequest {
  first_name: string;
  last_name: string;
  date_of_birth: string;
  email?: string;
  phone?: string;
  address?: string;
}

export interface UpdatePatientRequest extends Partial<CreatePatientRequest> {
  status?: 'active' | 'inactive' | 'archived';
}

export interface AssignClinicianRequest {
  clinician_id: string;
  role: 'primary' | 'secondary';
}

export interface ClinicianAssignment {
  clinician_id: string;
  full_name: string;
  email: string;
  role: 'primary' | 'secondary';
  assigned_at: string;
  assigned_by: string;
}

const patientService = {
  list: async () => {
    const response = await api.get<{ data: PaginatedResponse<Patient> }>("/patients");
    return response.data.data.items;
  },

  get: async (id: string) => {
    const response = await api.get<{ data: Patient }>(`/patients/${id}`);
    return response.data.data;
  },

  create: async (data: CreatePatientRequest) => {
    const response = await api.post<{ data: Patient }>("/patients", data);
    return response.data.data;
  },

  update: async (id: string, data: UpdatePatientRequest) => {
    const response = await api.put<{ data: Patient }>(`/patients/${id}`, data);
    return response.data.data;
  },

  delete: async (id: string) => {
    const response = await api.delete(`/patients/${id}`);
    return response.data;
  },

  assignClinician: async (patientId: string, data: AssignClinicianRequest) => {
    const response = await api.post<{ data: null }>(`/patients/${patientId}/assign`, data);
    return response.data.data;
  },

  unassignClinician: async (patientId: string, clinicianId: string) => {
    const response = await api.delete<{ data: null }>(`/patients/${patientId}/assign/${clinicianId}`);
    return response.data.data;
  },

  getAssignedClinicians: async (patientId: string) => {
    const response = await api.get<{ data: ClinicianAssignment[] }>(`/patients/${patientId}/assignments`);
    return response.data.data;
  },
};

export default patientService;
