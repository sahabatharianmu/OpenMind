import api from "@/api/client";
import { Invoice } from "@/types";
import type { PaginatedResponse } from "@/types/api";

export interface CreateInvoiceRequest {
  patient_id: string;
  appointment_id?: string;
  amount_cents: number;
  status?: string;
  due_date?: string;
  payment_method?: string;
  notes?: string;
}

export interface UpdateInvoiceRequest extends Partial<CreateInvoiceRequest> {
  paid_at?: string;
}

const invoiceService = {
  list: async () => {
    const response = await api.get<{ data: PaginatedResponse<Invoice> }>("/invoices");
    return response.data.data.items;
  },

  get: async (id: string) => {
    const response = await api.get<{ data: Invoice }>(`/invoices/${id}`);
    return response.data.data;
  },

  create: async (data: CreateInvoiceRequest) => {
    const response = await api.post<{ data: Invoice }>("/invoices", data);
    return response.data.data;
  },

  update: async (id: string, data: UpdateInvoiceRequest) => {
    const response = await api.put<{ data: Invoice }>(`/invoices/${id}`, data);
    return response.data.data;
  },

  delete: async (id: string) => {
    const response = await api.delete(`/invoices/${id}`);
    return response.data;
  },
};

export default invoiceService;
