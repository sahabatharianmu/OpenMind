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
};

export default patientService;
