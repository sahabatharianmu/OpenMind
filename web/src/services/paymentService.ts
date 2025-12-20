import api from "@/api/client";

export interface PaymentMethod {
  id: string;
  provider: string;
  last4: string;
  brand: string;
  expiry_month: number;
  expiry_year: number;
  is_default: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreatePaymentMethodRequest {
  token: string;
  provider?: string; // Optional: payment provider (stripe, midtrans). Defaults to configured default provider
}

export interface ListPaymentMethodsResponse {
  payment_methods: PaymentMethod[];
  total: number;
}

export const paymentService = {
  createPaymentMethod: async (data: CreatePaymentMethodRequest) => {
    const response = await api.post<{ data: PaymentMethod }>("/payment-methods", data);
    return response.data.data;
  },

  listPaymentMethods: async () => {
    const response = await api.get<{ data: ListPaymentMethodsResponse }>("/payment-methods");
    return response.data.data;
  },

  getPaymentMethod: async (id: string) => {
    const response = await api.get<{ data: PaymentMethod }>(`/payment-methods/${id}`);
    return response.data.data;
  },

  deletePaymentMethod: async (id: string) => {
    const response = await api.delete<{ data: null }>(`/payment-methods/${id}`);
    return response.data.data;
  },

  setDefaultPaymentMethod: async (id: string) => {
    const response = await api.put<{ data: null }>(`/payment-methods/${id}/default`, {});
    return response.data.data;
  },
};

