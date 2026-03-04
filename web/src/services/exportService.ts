import api from "@/api/client";

export const exportService = {
  exportAllData: async () => {
    const response = await api.get("/export", {
      responseType: "blob",
    });
    return response.data;
  },
};
