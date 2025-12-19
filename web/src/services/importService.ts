import api from "@/api/client";

export interface ImportPreviewRequest {
  type: "patients" | "appointments" | "notes";
  file_data: string; // Base64 encoded
  file_name: string;
}

export interface ImportExecuteRequest {
  type: "patients" | "appointments" | "notes";
  file_data: string; // Base64 encoded
  file_name: string;
}

export interface RowError {
  row: number;
  field?: string;
  message: string;
}

export interface RowWarning {
  row: number;
  field?: string;
  message: string;
}

export interface ImportPreviewResponse {
  total_rows: number;
  valid_rows: number;
  invalid_rows: number;
  preview: Record<string, any>[];
  errors?: RowError[];
  warnings?: RowWarning[];
}

export interface ImportExecuteResponse {
  total_rows: number;
  success_count: number;
  error_count: number;
  errors?: RowError[];
  imported_ids?: string[];
}

export const importService = {
  downloadTemplate: async (type: "patients" | "notes", format: "csv" | "xlsx" = "csv"): Promise<Blob> => {
    const response = await api.get(`/import/template/${type}?format=${format}`, {
      responseType: "blob",
    });
    return response.data;
  },

  previewImport: async (req: ImportPreviewRequest): Promise<ImportPreviewResponse> => {
    const response = await api.post("/import/preview", req);
    return response.data.data;
  },

  executeImport: async (req: ImportExecuteRequest): Promise<ImportExecuteResponse> => {
    const response = await api.post("/import/execute", req);
    return response.data.data;
  },
};

