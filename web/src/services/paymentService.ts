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

export interface PaymentTransaction {
  id: string;
  organization_id: string;
  amount: number;
  currency: string;
  status: 'pending' | 'paid' | 'failed' | 'cancelled';
  payment_method: string;
  type: string;
  partner_reference_no?: string;
  created_at: string;
  paid_at?: string;
}

export interface PaginatedPaymentTransactionResponse {
  data: PaymentTransaction[];
  total: number;
  limit: number;
  offset: number;
  total_pages: number;
}

export interface QRISDataResponse {
  id: string;
  transaction_id: string;
  partner_reference_no: string;
  qr_code: string;
  qr_code_url: string;
  qr_code_image: string;
  amount: number;
  currency: string;
  status: string;
  expires_at?: string;
  created_at: string;
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

  cancelPayment: async (transactionId: string) => {
    const response = await api.post<{ data: null }>(`/payments/${transactionId}/cancel`);
    return response.data;
  },

  listTransactions: async (limit: number = 10, offset: number = 0) => {
    const response = await api.get<{ data: PaginatedPaymentTransactionResponse }>(
      `/payments?limit=${limit}&offset=${offset}`
    );
    return response.data.data;
  },

  getQRISData: async (transactionId: string) => {
    const response = await api.get<{ data: QRISDataResponse }>(`/payments/${transactionId}/qris-data`);
    return response.data.data;
  },
};

