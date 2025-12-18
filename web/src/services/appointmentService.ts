import api from "@/api/client";
import { Appointment } from "@/types";
import type { PaginatedResponse } from "@/types/api";

export interface CreateAppointmentRequest {
  patient_id: string;
  clinician_id: string; // usually inferred from context but needed for admin/scheduling
  start_time: string;
  end_time: string;
  status?: string;
  appointment_type: string;
  mode: string;
  notes?: string;
}

export interface UpdateAppointmentRequest extends Partial<CreateAppointmentRequest> {}

const appointmentService = {
  list: async () => {
    const response = await api.get<{ data: PaginatedResponse<Appointment> }>("/appointments");
    return response.data.data.items;
  },

  get: async (id: string) => {
    const response = await api.get<{ data: Appointment }>(`/appointments/${id}`);
    return response.data.data;
  },

  create: async (data: CreateAppointmentRequest) => {
    const response = await api.post<{ data: Appointment }>("/appointments", data);
    return response.data.data;
  },

  update: async (id: string, data: UpdateAppointmentRequest) => {
    const response = await api.put<{ data: Appointment }>(`/appointments/${id}`, data);
    return response.data.data;
  },

  delete: async (id: string) => {
    const response = await api.delete(`/appointments/${id}`);
    return response.data;
  },
};

export default appointmentService;
