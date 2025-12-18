import api from "@/api/client";
import { ClinicalNote } from "@/types";
import type { PaginatedResponse } from "@/types/api";

export interface CreateClinicalNoteRequest {
  patient_id: string;
  clinician_id: string;
  appointment_id?: string;
  note_type: string;
  subjective?: string;
  objective?: string;
  assessment?: string;
  plan?: string;
}

export interface UpdateClinicalNoteRequest extends Partial<CreateClinicalNoteRequest> {
  is_signed?: boolean;
}

const clinicalNoteService = {
  list: async () => {
    const response = await api.get<{ data: PaginatedResponse<ClinicalNote> }>("/clinical-notes");
    return response.data.data.items;
  },

  get: async (id: string) => {
    const response = await api.get<{ data: ClinicalNote }>(`/clinical-notes/${id}`);
    return response.data.data;
  },

  create: async (data: CreateClinicalNoteRequest) => {
    const response = await api.post<{ data: ClinicalNote }>("/clinical-notes", data);
    return response.data.data;
  },

  update: async (id: string, data: UpdateClinicalNoteRequest) => {
    const response = await api.put<{ data: ClinicalNote }>(`/clinical-notes/${id}`, data);
    return response.data.data;
  },

  delete: async (id: string) => {
    const response = await api.delete(`/clinical-notes/${id}`);
    return response.data;
  },
};

export default clinicalNoteService;
