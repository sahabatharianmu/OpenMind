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

export interface RequestHandoffRequest {
  receiving_clinician_id: string;
  message?: string;
  role?: 'primary' | 'secondary';
}

export interface ApproveHandoffRequest {
  reason?: string;
}

export interface RejectHandoffRequest {
  reason: string;
}

export interface Handoff {
  id: string;
  patient_id: string;
  patient_name: string;
  requesting_clinician_id: string;
  requesting_clinician_name: string;
  requesting_clinician_email: string;
  receiving_clinician_id: string;
  receiving_clinician_name: string;
  receiving_clinician_email: string;
  status: 'requested' | 'approved' | 'rejected' | 'cancelled';
  requested_role?: 'primary' | 'secondary';
  message?: string;
  requested_at: string;
  responded_at?: string;
  responded_by?: string;
  created_at: string;
  updated_at: string;
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

  // Handoff methods
  requestHandoff: async (patientId: string, data: RequestHandoffRequest) => {
    const response = await api.post<{ data: Handoff }>(`/patients/${patientId}/handoff`, data);
    return response.data.data;
  },

  approveHandoff: async (handoffId: string, data: ApproveHandoffRequest) => {
    const response = await api.post<{ data: null }>(`/patients/handoffs/${handoffId}/approve`, data);
    return response.data.data;
  },

  rejectHandoff: async (handoffId: string, data: RejectHandoffRequest) => {
    const response = await api.post<{ data: null }>(`/patients/handoffs/${handoffId}/reject`, data);
    return response.data.data;
  },

  cancelHandoff: async (handoffId: string) => {
    const response = await api.post<{ data: null }>(`/patients/handoffs/${handoffId}/cancel`, {});
    return response.data.data;
  },

  getHandoff: async (handoffId: string) => {
    const response = await api.get<{ data: Handoff }>(`/patients/handoffs/${handoffId}`);
    return response.data.data;
  },

  listHandoffs: async (patientId: string) => {
    const response = await api.get<{ data: Handoff[] }>(`/patients/${patientId}/handoffs`);
    return response.data.data;
  },

  listPendingHandoffs: async () => {
    const response = await api.get<{ data: Handoff[] }>("/patients/handoffs/pending");
    return response.data.data;
  },

  isAssigned: async (patientId: string, userId?: string): Promise<boolean> => {
    try {
      const assignments = await patientService.getAssignedClinicians(patientId);
      
      // Get user ID from parameter or localStorage
      let currentUserId: string;
      if (userId) {
        currentUserId = String(userId);
      } else {
        const userStr = localStorage.getItem("user_profile");
        if (!userStr) {
          console.warn("No user profile found in localStorage and no userId provided");
          return false;
        }
        const user = JSON.parse(userStr);
        currentUserId = String(user.id);
      }
      
      // Compare IDs (handle both string and UUID formats)
      const isAssigned = assignments.some(a => {
        const assignmentClinicianId = String(a.clinician_id);
        return assignmentClinicianId === currentUserId;
      });
      
      console.log("Assignment check:", { 
        patientId, 
        userId: currentUserId, 
        assignments: assignments.map(a => ({ id: String(a.clinician_id), name: a.full_name })),
        isAssigned 
      });
      
      return isAssigned;
    } catch (error) {
      console.error("Error checking assignment:", error);
      return false;
    }
  },
};

export default patientService;
